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

const MAXTAKE = 10

func MatchSearch(sellers []models.Seller, SearchRes []search_engine.Seller) []int32 {
	var hasil []int32
	sellerMap := make(map[int32]struct{})

	for _, s := range sellers {
		sellerMap[s.ID] = struct{}{}
	}

	for _, data := range SearchRes {
		if _, found := sellerMap[data.IDKey]; !found {
			hasil = append(hasil, data.IDKey)
		}
	}

	return hasil
}

func AmbilRandomSeller(FinalTake int64, ctx context.Context, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilRandomSeller"
	var sellers []models.Seller

	// ambil random key dari Redis
	keys, err := rds.SMembers(ctx, "all_seller_keys").Result()
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

	if int(FinalTake) >= len(keys) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "FinalTake melebihi jumlah seller yang tersedia di Redis",
		}
	}

	end := int(FinalTake) + MAXTAKE
	if end > len(keys) {
		end = len(keys)
	}
	for _, rawKey := range keys[FinalTake:end] {
		key := strings.Trim(rawKey, `"`)
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
		if errDB := db.Model(&models.Seller{}).Limit(MAXTAKE).Find(&sellers).Error; errDB != nil {
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
	FinalTake int64,
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
	searchRes, err := SE.Index("seller_all").Search(nama_seller, &meilisearch.SearchRequest{
		Limit:  MAXTAKE,
		Sort:   []string{"follower_total:desc"},
		Offset: FinalTake,
	})
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
	hitsBytes, err := json.Marshal(searchRes.Hits)
	if err != nil {
		fmt.Printf("[ERROR] ❌ Gagal marshal hasil Meili ke bytes: %v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  fmt.Sprintf("Gagal memproses hasil Meilisearch: %v", err),
		}
	}

	if err := json.Unmarshal(hitsBytes, &idsellers); err != nil {
		fmt.Printf("[ERROR] ❌ Gagal unmarshal hasil Meili ke struct idsellers: %v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  fmt.Sprintf("Gagal decode hasil Meilisearch: %v", err),
		}
	}

	fmt.Printf("[TRACE] ✅ Meilisearch mengembalikan %d hasil\n", len(idsellers))

	// --- Ambil data dari Redis secara paralel ---
	startIdx := int(FinalTake)
	endIdx := startIdx + int(MAXTAKE)
	if endIdx > len(idsellers) {
		endIdx = len(idsellers)
	}

	for _, ds := range idsellers[startIdx:endIdx] {
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

			idSeller, errID := strconv.Atoi(data["id_seller"])
			follower, _ := strconv.Atoi(data["follower_total_seller"])
			if errID != nil || idSeller == 0 {
				return
			}

			seller := models.Seller{
				ID:               int32(idSeller),
				Username:         data["username_seller"],
				Nama:             data["nama_seller"],
				JamOperasional:   data["jam_operasional_seller"],
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

	// --- Fallback: ambil dari DB kalau Redis tidak lengkap ---
	if len(sellers) < int(MAXTAKE) {
		var hasilDB []models.Seller
		missingIDs := MatchSearch(sellers, idsellers)
		if len(missingIDs) > 0 {
			if err := db.Model(&models.Seller{}).Where("id IN ?", missingIDs).Find(&hasilDB).Error; err != nil {
				fmt.Printf("[ERROR] ❌ Gagal ambil fallback DB: %v\n", err)
			} else {
				mu.Lock()
				sellers = append(sellers, hasilDB...)
				mu.Unlock()
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

func AmbilSellerByJenis(FinalTake int64, ctx context.Context, jenis string, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "AmbilSellerByJenis"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller

	keys, err := rds.SMembers(ctx, fmt.Sprintf("seller_jenis:%s", jenis)).Result()
	if err != nil {
		fmt.Printf("[ERROR] Gagal ambil key Redis: %v\n", err)
		keys = nil
	}

	if len(keys) == 0 {
		if errDB := db.Model(&models.Seller{}).
			Where(&models.Seller{Jenis: jenis}).
			Offset(int(FinalTake)).
			Limit(MAXTAKE).
			Find(&sellers).Error; errDB != nil {
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

	if int(FinalTake) >= len(keys) {
		var dbSellers []models.Seller

		if errDB := db.Where("jenis = ?", jenis).
			Offset(int(FinalTake)).
			Limit(MAXTAKE).
			Find(&dbSellers).Error; errDB != nil {
			fmt.Printf("[INFO] Tidak ada data seller jenis '%s' di offset %d (DB fallback error: %v)\n", jenis, FinalTake, errDB)
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: services,
				Payload:  []models.Seller{},
			}
		}

		if len(dbSellers) == 0 {
			fmt.Printf("[INFO] Data seller jenis '%s' habis di offset %d\n", jenis, FinalTake)
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: services,
				Payload:  []models.Seller{},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  dbSellers,
		}
	}

	end := int(FinalTake) + MAXTAKE
	if end > len(keys) {
		end = len(keys)
	}

	for _, key := range keys[FinalTake:end] {
		if key == "_init_" || strings.TrimSpace(key) == "" {
			continue
		}

		wg.Add(1)
		go func(k string) {
			defer wg.Done()

			data, errData := rds.HGetAll(ctx, k).Result()
			if errData != nil || len(data) == 0 {
				return
			}

			idSeller, errID := strconv.Atoi(data["id_seller"])
			if errID != nil {
				return
			}

			followerTotal, errFollower := strconv.Atoi(data["follower_total_seller"])
			if errFollower != nil {
				return
			}

			sellerData := models.Seller{
				ID:               int32(idSeller),
				Username:         data["username_seller"],
				Email:            data["email_seller"],
				JamOperasional:   data["jam_operasional_seller"],
				Jenis:            data["jenis_seller"],
				Punchline:        data["punchline_seller"],
				Deskripsi:        data["deskripsi_seller"],
				Nama:             data["nama_seller"],
				SellerDedication: data["seller_dedication_seller"],
				FollowerTotal:    int32(followerTotal),
			}

			if sellerData.Username != "" {
				mu.Lock()
				sellers = append(sellers, sellerData)
				mu.Unlock()
			}
		}(key)
	}

	wg.Wait()

	fmt.Println("Jalannih", keys)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByDedication(FinalTake int64, ctx context.Context, dedication string, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "AmbilSellerByDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller

	keySet := fmt.Sprintf("seller_dedication:%s", dedication)
	keys, err := rds.SMembers(ctx, keySet).Result()
	if err != nil {
		fmt.Printf("[ERROR][%s] Gagal ambil key Redis: %v\n", services, err)
		keys = nil
	}

	if len(keys) == 0 {
		fmt.Printf("[INFO][%s] Redis kosong, fallback ke DB (dedication=%s)\n", services, dedication)
		if errDB := db.Model(&models.Seller{}).
			Where("seller_dedication = ?", dedication).Order("follower_total DESC").
			Offset(int(FinalTake)).
			Limit(MAXTAKE).
			Find(&sellers).Error; errDB != nil {
			fmt.Printf("[ERROR][%s] Gagal ambil data dari DB: %v\n", services, errDB)
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: services,
				Payload:  []models.Seller{},
			}
		}

		if len(sellers) == 0 {
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: services,
				Payload:  []models.Seller{},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  sellers,
		}
	}

	if int(FinalTake) >= len(keys) {
		fmt.Printf("[INFO][%s] Offset %d melebihi jumlah key Redis, fallback DB\n", services, FinalTake)
		if errDB := db.Model(&models.Seller{}).Order("follower_total DESC").
			Where("seller_dedication = ?", dedication).
			Offset(int(FinalTake)).
			Limit(MAXTAKE).
			Find(&sellers).Error; errDB != nil || len(sellers) == 0 {
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: services,
				Payload:  []models.Seller{},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  sellers,
		}
	}

	end := int(FinalTake) + MAXTAKE
	if end > len(keys) {
		end = len(keys)
	}

	for _, key := range keys[FinalTake:end] {
		if key == "_init_" || strings.TrimSpace(key) == "" {
			continue
		}

		wg.Add(1)
		go func(k string) {
			defer wg.Done()

			data, err := rds.HGetAll(ctx, k).Result()
			if err != nil || len(data) == 0 {
				return
			}

			idSeller, errID := strconv.Atoi(data["id_seller"])
			if errID != nil {
				return
			}

			followerTotal, errFollower := strconv.Atoi(data["follower_total_seller"])
			if errFollower != nil {
				return
			}

			sellerData := models.Seller{
				ID:               int32(idSeller),
				Username:         data["username_seller"],
				Email:            data["email_seller"],
				JamOperasional:   data["jam_operasional_seller"],
				Jenis:            data["jenis_seller"],
				Punchline:        data["punchline_seller"],
				Deskripsi:        data["deskripsi_seller"],
				Nama:             data["nama_seller"],
				SellerDedication: data["seller_dedication_seller"],
				FollowerTotal:    int32(followerTotal),
			}

			if sellerData.Username != "" {
				mu.Lock()
				sellers = append(sellers, sellerData)
				mu.Unlock()
			}
		}(key)
	}

	wg.Wait()

	if len(sellers) < MAXTAKE {
		fmt.Printf("[INFO][%s] Redis gak punya data valid, fallback DB\n", services)
		var dbSellers []models.Seller
		if errDB := db.Model(&models.Seller{}).
			Order("follower_total DESC").
			Where(&models.Seller{SellerDedication: dedication}).
			Offset(len(sellers)).
			Limit(MAXTAKE - len(sellers)).
			Find(&dbSellers).Error; errDB == nil && len(dbSellers) > 0 {
			mu.Lock()
			sellers = append(sellers, dbSellers...)
			mu.Unlock()
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByNamaDanJenis(FinalTake int64, ctx context.Context, nama, jenis string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilSellerByNamaDanJenis"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var idSeller []search_engine.Seller
	var sellers []models.Seller

	start := time.Now()

	ResultSe, err := SE.Index("seller_all").Search(nama, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("jenis_seller = '%s'", jenis),
		Sort:   []string{"follower_total:desc"},
		Limit:  MAXTAKE,
		Offset: FinalTake,
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

	if len(sellers) < MAXTAKE {
		var Hasil []models.Seller

		if err := db.Model(&models.Seller{}).Where("id IN ?", MatchSearch(sellers, idSeller)).Find(&Hasil).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  []models.Seller{},
			}
		}
		mu.Lock()
		sellers = append(sellers, Hasil...)
		mu.Unlock()
	}

	log.Printf("[%s] ✅ Selesai. Dapat %d seller dalam %v\n", services, len(sellers), time.Since(start))

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByNamaDanDedication(FinalTake int64, ctx context.Context, nama, dedication string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilSellerByNameDanDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var idSeller []search_engine.Seller
	var sellers []models.Seller

	ResultSe, err := SE.Index("seller_all").Search(nama, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("seller_dedication_seller = '%s'", dedication),
		Sort:   []string{"follower_total:desc"},
		Offset: FinalTake,
		Limit:  MAXTAKE,
	})
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  []models.Seller{},
		}
	}

	hitsjson, _ := json.Marshal(ResultSe.Hits)

	if err := json.Unmarshal(hitsjson, &idSeller); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  []models.Seller{},
		}
	}

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

	if len(sellers) < MAXTAKE {
		var Hasil []models.Seller

		if err := db.Model(&models.Seller{}).Where("id IN ?", MatchSearch(sellers, idSeller)).Find(&Hasil).Error; err != nil {
			fmt.Println("Gagal Ambil Fallback")
		} else {
			mu.Lock()
			sellers = append(sellers, Hasil...)
			mu.Unlock()
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

func AmbilSellerByJenisDanDedication(FinalTake int64, ctx context.Context, jenis, dedication string, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "AmbilSellerByJenisDanDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller
	var keysSeller []int64

	if err := db.Model(&models.Seller{}).Select("id").Where(&models.Seller{Jenis: jenis, SellerDedication: dedication}).Find(&keysSeller).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []models.Seller{},
		}
	}

	if len(keysSeller) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  []models.Seller{},
		}
	}

	if int(FinalTake) >= len(keysSeller) {
		var hasilDB []models.Seller
		if err := db.Model(&models.Seller{}).Offset(int(FinalTake)).Order("follower_total DESC").
			Where(&models.Seller{Jenis: jenis, SellerDedication: dedication}).
			Limit(int(MAXTAKE)).
			Find(&hasilDB).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  []models.Seller{},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  hasilDB,
		}
	}

	endIndex := int(FinalTake) + MAXTAKE
	if endIndex > len(keysSeller) {
		endIndex = len(keysSeller)
	}

	for _, ks := range keysSeller[FinalTake:endIndex] {
		wg.Add(1)
		go func(key int64) {
			defer wg.Done()
			seller, err := rds.HGetAll(ctx, fmt.Sprintf("seller_data:%v", key)).Result()
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

func AmbilSellerByNamaJenisDedication(FinalTake int64, ctx context.Context, nama, jenis, dedication string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilSellerByNamaJenisDedication"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sellers []models.Seller
	var idSeller []search_engine.Seller

	ResultSe, err := SE.Index("seller_all").Search(nama, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("seller_dedication_seller = \"%s\" AND jenis_seller = \"%s\"", dedication, jenis),
		Sort:   []string{"follower_total:desc"},
		Offset: FinalTake,
		Limit:  MAXTAKE,
	})

	if err != nil {
		log.Printf("[%s] ⚠️ Gagal melakukan pencarian di Meilisearch: %v\n", services, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  []models.Seller{},
		}
	}

	hitsjson, _ := json.Marshal(ResultSe.Hits)
	if err := json.Unmarshal(hitsjson, &idSeller); err != nil {
		log.Printf("[%s] ⚠️ Gagal unmarshal hasil Meilisearch: %v\n", services, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  []models.Seller{},
		}
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

	if len(sellers) < MAXTAKE {
		var Hasil []models.Seller

		if err := db.Model(&models.Seller{}).Where("id IN ?", MatchSearch(sellers, idSeller)).Find(&Hasil).Error; err != nil {
			fmt.Println("Gagal Ambil Fallback")
		} else {
			mu.Lock()
			sellers = append(sellers, Hasil...)
			mu.Unlock()
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}
