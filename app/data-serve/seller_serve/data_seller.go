package seller_serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/data-serve/seller_serve/response_seller"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/search_engine"
)

func AmbilRandomSeller(ctx context.Context, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilRandomSeller"
	var sellers []models.Seller
	const maxSeller = 20

	// ambil random key dari Redis
	keys, err := rds.SRandMemberN(ctx, "all_seller_keys", int64(maxSeller)).Result()
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  fmt.Sprintf("Gagal mengambil kunci seller dari Redis: %v", err),
		}
	}

	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Tidak ditemukan seller di cache Redis.",
		}
	}

	for _, rawKey := range keys {
		key := strings.Trim(rawKey, `"`) // hapus kutip ganda kalau ada
		if key == "_init_" {
			continue
		}

		// debug: tampilkan key yang ditemukan (untuk diagnosa)
		fmt.Printf("[DEBUG] Redis key: %s\n", rawKey)
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			var seller models.Seller

			// cek apakah key valid dan bertipe hash
			keyType, errType := rds.Type(ctx, k).Result()
			if errType != nil || keyType != "hash" {
				fmt.Printf("[WARN] Key '%s' bukan hash atau tidak dapat diambil: %v\n", k, errType)
				return
			}

			data, errData := rds.HGetAll(ctx, k).Result()
			if errData != nil || len(data) == 0 {
				fmt.Printf("[WARN] HGetAll kosong atau error untuk key '%s': %v\n", k, errData)
				return
			}

			idSeller, errID := strconv.Atoi(data["id_seller"])
			follower, _ := strconv.Atoi(data["follower_total_seller"])
			if errID != nil || idSeller == 0 {
				return
			}

			seller = models.Seller{
				ID:               int32(idSeller),
				Username:         data["username_seller"],
				Nama:             data["nama_seller"],
				JamOperasional:   data["jam_operasional_seller"],
				Password:         "",
				Norek:            "",
				SellerDedication: data["seller_dedication_seller"],
				Deskripsi:        data["deskripsi_seller"],
				FollowerTotal:    int32(follower),
				Email:            data["email_seller"],
				Jenis:            data["jenis_seller"],
				Punchline:        data["punchline_seller"],
			}

			mu.Lock()
			sellers = append(sellers, seller)
			mu.Unlock()
		}(key)
	}

	wg.Wait()

	// fallback ke database jika data Redis kosong total
	if len(sellers) == 0 {
		fmt.Println("[INFO] Data cache kosong, mengambil seller dari database sebagai fallback")
		if errDB := db.Model(&models.Seller{}).Limit(maxSeller).Find(&sellers).Error; errDB != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  fmt.Sprintf("Gagal mengambil seller dari database: %v", errDB),
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByNama(
	ctx context.Context,
	nama_seller string,
	db *gorm.DB,
	rds *redis.Client,
	SE meilisearch.ServiceManager,
) *response.ResponseForm {
	services := "AmbilSellerByNama"
	var sellers []models.Seller
	var idsellers []search_engine.Seller
	var wg sync.WaitGroup
	var mu sync.Mutex

	fmt.Printf("[TRACE] 🔍 Memulai pencarian seller nama: %s\n", nama_seller)

	// --- Pencarian Meilisearch ---
	searchRes, err := SE.Index("seller_all").Search(nama_seller, &meilisearch.SearchRequest{})
	if err != nil {
		fmt.Printf("[ERROR] ❌ Meilisearch gagal: %v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  fmt.Sprintf("Gagal melakukan pencarian di Meilisearch: %v", err),
		}
	}

	if len(searchRes.Hits) == 0 {
		fmt.Println("[TRACE] ⚠️ Tidak ditemukan hasil di Meilisearch.")
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Seller tidak ditemukan di indeks pencarian.",
		}
	}

	// --- Parsing hasil Meilisearch ---
	hitsjson, errMarshal := json.Marshal(searchRes.Hits)
	if errMarshal != nil {
		fmt.Printf("[ERROR] ❌ Gagal marshal hasil pencarian Meili: %v\n", errMarshal)
	}
	if errUnmarshal := json.Unmarshal(hitsjson, &idsellers); errUnmarshal != nil {
		fmt.Printf("[ERROR] ❌ Gagal unmarshal hasil Meili ke struct idsellers: %v\n", errUnmarshal)
	}

	fmt.Printf("[TRACE] ✅ Meilisearch mengembalikan %d hasil\n", len(idsellers))

	for _, ds := range idsellers {
		wg.Add(1)
		go func(ds search_engine.Seller) {
			defer wg.Done()
			key := fmt.Sprintf("seller_data:%v", ds.IdSeller)

			data, errData := rds.HGetAll(ctx, key).Result()
			if errData != nil {
				fmt.Printf("[WARN] ⚠️ Gagal ambil Redis key %s: %v\n", key, errData)
				return
			}
			if len(data) == 0 {
				fmt.Printf("[TRACE] ⚠️ Redis key %s kosong, mungkin belum di-cache\n", key)
				return
			}

			var seller models.Seller
			idSeller, errID := strconv.Atoi(data["id_seller"])
			follower, _ := strconv.Atoi(data["follower_total_seller"])
			if errID != nil || idSeller == 0 {
				return
			}

			seller = models.Seller{
				ID:               int32(idSeller),
				Username:         data["username_seller"],
				Nama:             data["nama_seller"],
				JamOperasional:   data["jam_operasional_seller"],
				Password:         "",
				Norek:            "",
				SellerDedication: data["seller_dedication_seller"],
				Deskripsi:        data["deskripsi_seller"],
				FollowerTotal:    int32(follower),
				Email:            data["email_seller"],
				Jenis:            data["jenis_seller"],
				Punchline:        data["punchline_seller"],
			}

			mu.Lock()
			sellers = append(sellers, seller)
			mu.Unlock()

			fmt.Printf("[TRACE] ✅ Seller ID %v (%s) berhasil dimuat dari Redis\n", seller.ID, seller.Nama)
		}(ds)
	}

	wg.Wait()

	// --- Jika Redis kosong semua, fallback ke database ---
	if len(sellers) == 0 {
		fmt.Println("[TRACE] ⚠️ Semua data Redis kosong, fallback ambil dari database...")
		if errDB := db.Model(&models.Seller{}).
			Where("nama_seller LIKE ?", "%"+nama_seller+"%").
			Limit(10).Find(&sellers).Error; errDB != nil {
			fmt.Printf("[ERROR] ❌ Gagal ambil dari database: %v\n", errDB)
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  fmt.Sprintf("Gagal ambil dari database: %v", errDB),
			}
		}
	}

	fmt.Printf("[TRACE] 🎯 Total seller ditemukan: %d\n", len(sellers))

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByJenis(ctx context.Context, jenis string, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "AmbilSellerByJenis"
	var wg sync.WaitGroup
	var mu sync.Mutex

	var sellers []models.Seller

	keys, err := rds.SRandMemberN(ctx, fmt.Sprintf("seller_jenis:%s", jenis), 10).Result()
	if err != nil || len(keys) == 0 {
		if err_db := db.Model(&models.Seller{}).Where(&models.Seller{
			Jenis: jenis,
		}).Limit(10).Find(&sellers).Error; err_db != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  response_seller.ResponseAmbilJenisSeller{Message: "Gagal mengambil data dari DB"},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  sellers,
		}
	}

	for _, key := range keys {
		if key == "_init_" {
			continue
		}

		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			data, errData := rds.HGetAll(ctx, key).Result()
			if errData != nil || len(data) == 0 {
				return
			}

			var seller models.Seller
			jsonData, _ := json.Marshal(data)
			if errUnmarshal := json.Unmarshal(jsonData, &seller); errUnmarshal != nil {
				return
			}

			mu.Lock()
			sellers = append(sellers, seller)
			mu.Unlock()
		}(key)
	}

	wg.Wait()

	if len(sellers) < 10 {
		var dbSellers []models.Seller
		if err_db := db.Where("jenis = ?", jenis).Limit(10 - len(sellers)).Find(&dbSellers).Error; err_db == nil {
			sellers = append(sellers, dbSellers...)
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByDedication(ctx context.Context, dedication string, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "AmbilSellerByDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller
	var key_seller []string

	keys := fmt.Sprintf(`"seller_dedication:%s"`, dedication)

	if hasil, err := rds.SRandMemberN(ctx, keys, 10).Result(); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	} else {
		key_seller = append(key_seller, hasil...)
	}

	for _, ks := range key_seller {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			seller, err := rds.HGetAll(ctx, key).Result()
			if err != nil || len(seller) == 0 {
				return
			}

			idSeller, errID := strconv.Atoi(seller["id_seller"])
			if errID != nil {
				return
			}
			followerTotal, errFollower := strconv.Atoi(seller["follower_total_seller"])
			if errFollower != nil {
				return
			}

			sellerData := models.Seller{
				ID:               int32(idSeller),
				Username:         seller["username_seller"],
				Email:            seller["email_seller"],
				JamOperasional:   seller["jam_operasional_seller"],
				Jenis:            seller["jenis_seller"],
				Punchline:        seller["punchline_seller"],
				Deskripsi:        seller["deskripsi_seller"],
				Nama:             seller["nama_seller"],
				SellerDedication: seller["seller_dedication_seller"],
				FollowerTotal:    int32(followerTotal),
			}

			if sellerData.Username != "" {
				mu.Lock()
				sellers = append(sellers, sellerData)
				mu.Unlock()
			}
		}(ks)
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByNamaDanJenis(ctx context.Context, nama, jenis string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilSellerByNamaDanJenis"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var idSeller []search_engine.Seller
	var sellers []models.Seller

	start := time.Now()

	ResultSe, err := SE.Index("seller_all").Search(nama, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("jenis_seller = '%s'", jenis),
		Limit:  10,
	})
	if err != nil {
		log.Printf("[%s] ❌ Meilisearch error: %v\n", services, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	hitsjson, _ := json.Marshal(ResultSe.Hits)
	if err := json.Unmarshal(hitsjson, &idSeller); err != nil {
		log.Printf("[%s] ⚠️ Gagal unmarshal hasil Meilisearch: %v\n", services, err)
	}

	if len(idSeller) == 0 {
		log.Printf("[%s] ⚠️ Tidak ada hasil ditemukan untuk '%s' jenis '%s'\n", services, nama, jenis)
	}

	for _, id := range idSeller {
		wg.Add(1)
		go func(id search_engine.Seller) {
			defer wg.Done()

			seller, errRds := rds.HGetAll(ctx, fmt.Sprintf("seller_data:%v", id.IdSeller)).Result()
			if errRds != nil {
				log.Printf("[%s] ⚠️ Redis miss untuk seller ID %v: %v\n", services, id.IdSeller, errRds)
				return
			}

			if len(seller) == 0 {
				log.Printf("[%s] ⚠️ Redis kosong untuk seller ID %v\n", services, id.IdSeller)
				return
			}

			id_seller, _ := strconv.Atoi(seller["id_seller"])
			follower_total, _ := strconv.Atoi(seller["follower_total_seller"])

			mu.Lock()
			sellers = append(sellers, models.Seller{
				ID:               int32(id_seller),
				Username:         seller["username_seller"],
				Email:            seller["email_seller"],
				JamOperasional:   seller["jam_operasional_seller"],
				Jenis:            seller["jenis_seller"],
				Punchline:        seller["punchline_seller"],
				Deskripsi:        seller["deskripsi_seller"],
				Nama:             seller["nama_seller"],
				SellerDedication: seller["seller_dedication_seller"],
				FollowerTotal:    int32(follower_total),
			})
			mu.Unlock()
		}(id)
	}

	wg.Wait()

	log.Printf("[%s] ✅ Selesai. Dapat %d seller dalam %v\n", services, len(sellers), time.Since(start))

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByNamaDanDedication(ctx context.Context, nama, dedication string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilSellerByNameDanDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var idSeller []search_engine.Seller
	var sellers []models.Seller

	ResultSe, err := SE.Index("seller_all").Search(nama, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("seller_dedication_seller = '%s'", dedication),
		Limit:  10,
	})
	if err == nil {
		hitsjson, _ := json.Marshal(ResultSe.Hits)

		_ = json.Unmarshal(hitsjson, &idSeller)
		for _, id := range idSeller {
			wg.Add(1)
			go func(id search_engine.Seller) {
				defer wg.Done()
				seller, _ := rds.HGetAll(ctx, fmt.Sprintf("seller_data:%v", id.IdSeller)).Result()

				id_seller, err_id := strconv.Atoi(seller["id_seller"])
				if err_id != nil {
					return
				}
				follower_total, err_follower := strconv.Atoi(seller["follower_total_seller"])
				if err_follower != nil {
					return
				}

				seller_data := models.Seller{
					ID:               int32(id_seller),
					Username:         seller["username_seller"],
					Email:            seller["email_seller"],
					JamOperasional:   seller["jam_operasional_seller"],
					Jenis:            seller["jenis_seller"],
					Punchline:        seller["punchline_seller"],
					Deskripsi:        seller["deskripsi_seller"],
					Nama:             seller["nama_seller"],
					SellerDedication: seller["seller_dedication_seller"],
					FollowerTotal:    int32(follower_total),
				}

				if seller_data.Username != "" {
					mu.Lock()
					sellers = append(sellers, seller_data)
					mu.Unlock()
				} else {
					return
				}
			}(id)
		}
		wg.Wait()

		if len(sellers) == 0 {
			var fallbackid []int64
			for _, id := range idSeller {
				fallbackid = append(fallbackid, int64(id.IdSeller))
			}

			_ = db.Model(&models.Seller{}).Where("id IN ?", fallbackid).Find(&sellers)
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByJenisDanDedication(ctx context.Context, jenis, dedication string, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "AmbilSellerByJenisDanDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller
	var rawKeysSeller []string
	var keysSeller []string

	keyJenis := fmt.Sprintf("seller_jenis:%s", jenis)
	if hasil, err := rds.SRandMemberN(ctx, keyJenis, 10).Result(); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	} else {
		rawKeysSeller = append(rawKeysSeller, hasil...)
	}

	keyDedication := fmt.Sprintf("seller_dedication:%s", dedication)
	if hasilDed, err := rds.SMembers(ctx, keyDedication).Result(); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	} else {
		dedMap := make(map[string]struct{}, len(hasilDed))
		for _, k := range hasilDed {
			dedMap[k] = struct{}{}
		}

		for _, ks := range rawKeysSeller {
			if _, ok := dedMap[ks]; ok {
				keysSeller = append(keysSeller, ks)
			}
		}
	}

	for _, ks := range keysSeller {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			seller, err := rds.HGetAll(ctx, key).Result()
			if err != nil || len(seller) == 0 {
				return
			}

			idSeller, errID := strconv.Atoi(seller["id_seller"])
			if errID != nil {
				return
			}
			followerTotal, errFollower := strconv.Atoi(seller["follower_total_seller"])
			if errFollower != nil {
				return
			}

			sellerData := models.Seller{
				ID:               int32(idSeller),
				Username:         seller["username_seller"],
				Email:            seller["email_seller"],
				JamOperasional:   seller["jam_operasional_seller"],
				Jenis:            seller["jenis_seller"],
				Punchline:        seller["punchline_seller"],
				Deskripsi:        seller["deskripsi_seller"],
				Nama:             seller["nama_seller"],
				SellerDedication: seller["seller_dedication_seller"],
				FollowerTotal:    int32(followerTotal),
			}

			if sellerData.Username != "" {
				mu.Lock()
				sellers = append(sellers, sellerData)
				mu.Unlock()
			}
		}(ks)
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByNamaJenisDedication(ctx context.Context, nama, jenis, dedication string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilSellerByNamaJenisDedication"
	const LIMIT = 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller
	var idSeller []search_engine.Seller
	var hitsjson []byte

	ResultSe, err := SE.Index("seller_all").Search(nama, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("seller_dedication_seller = '%s' AND jenis_seller = '%s' ", dedication, jenis),
		Limit:  LIMIT,
	})

	if err != nil {
		log.Printf("[%s] ⚠️ Gagal melakukan pencarian di Meilisearch: %v\n", services, err)
		goto FALLBACKDB
	}

	hitsjson, _ = json.Marshal(ResultSe.Hits)
	if err := json.Unmarshal(hitsjson, &idSeller); err != nil {
		log.Printf("[%s] ⚠️ Gagal unmarshal hasil Meilisearch: %v\n", services, err)
	}

	if len(idSeller) == 0 {
		log.Printf("[%s] ⚠️ Tidak ada hasil ditemukan untuk '%s' jenis '%s'\n", services, nama, jenis)
	}

	for _, ds := range idSeller {
		wg.Add(1)
		go func(id search_engine.Seller) {
			defer wg.Done()

			sellerMap, err := rds.HGetAll(ctx, fmt.Sprintf("seller_data:%v", id.IdSeller)).Result()
			if err != nil {
				log.Printf("[%s] ⚠️ Gagal mengambil data seller dari Redis: %v\n", services, err)
				return
			}

			id_seller, errID := strconv.Atoi(sellerMap["id_seller"])
			if errID != nil {
				return
			}

			follower_total, errFollower := strconv.Atoi(sellerMap["follower_total_seller"])
			if errFollower != nil {
				return
			}

			sellerData := models.Seller{
				ID:               int32(id_seller),
				Username:         sellerMap["username_seller"],
				Email:            sellerMap["email_seller"],
				JamOperasional:   sellerMap["jam_operasional_seller"],
				Jenis:            sellerMap["jenis_seller"],
				Punchline:        sellerMap["punchline_seller"],
				Deskripsi:        sellerMap["deskripsi_seller"],
				Nama:             sellerMap["nama_seller"],
				SellerDedication: sellerMap["seller_dedication_seller"],
				FollowerTotal:    int32(follower_total),
			}

			if sellerData.Username != "" {
				mu.Lock()
				sellers = append(sellers, sellerData)
				mu.Unlock()
			}
		}(ds)
	}

	wg.Wait()

	if len(sellers) > 0 {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  sellers,
		}
	}

FALLBACKDB:
	var sellersFallback []models.Seller
	if errDB := db.Model(&models.Seller{}).
		Where("nama LIKE ? AND jenis = ? AND seller_dedication = ?", "%"+nama+"%", jenis, dedication).
		Limit(LIMIT).
		Find(&sellersFallback).Error; errDB != nil {
		log.Printf("[%s] ⚠️ Gagal fallback DB: %v\n", services, errDB)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellersFallback,
	}
}

// func AmbilMostFollowedSeller(ctx context.Context, SE meilisearch.ServiceManager, rds *redis.Client) *response.ResponseForm {
// 	services := "AmbilMostFollowedSeller"
// 	var sellers models.Seller
// 	return &response.ResponseForm{
// 		Status:   http.StatusOK,
// 		Services: services,

// 	}
// }
