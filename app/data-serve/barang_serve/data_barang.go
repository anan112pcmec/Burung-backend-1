package barang_serve

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/data-serve/barang_serve/response_barang"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/search_engine"
)

const MAXTAKE = 10

func MatchSearch(barang []models.BarangInduk, SearchRes []search_engine.BarangInduk) []int32 {
	var hasil []int32
	BarangIndukMap := make(map[int32]struct{})

	for _, s := range barang {
		BarangIndukMap[s.ID] = struct{}{}
	}

	for _, data := range SearchRes {
		if _, found := BarangIndukMap[int32(data.IDKey)]; !found {
			hasil = append(hasil, int32(data.IDKey))
		}
	}

	return hasil
}

func NotIn(barang []models.BarangInduk) []int32 {
	var hasil []int32

	for _, data := range barang {
		hasil = append(hasil, data.ID)
	}

	return hasil
}

// /////////////////////////////////////////////////////////////////////////////////////////////////
// Searching Barang Service
// /////////////////////////////////////////////////////////////////////////////////////////////////

func AmbilRandomBarang(FinalTake int, ctx context.Context, rds *redis.Client, db *gorm.DB) *response.ResponseForm {

	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilRandomBarang"
	var barang []models.BarangInduk

	keys, _ := rds.SMembers(ctx, "barang_keys").Result()
	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal, mungkin server sedang sibuk atau maintenance. Coba lagi nanti.",
		}
	}

	if FinalTake >= len(keys) {
		_ = db.Model(&models.BarangInduk{}).Order("viewed DESC, likes DESC").Offset(FinalTake).Limit(MAXTAKE).Find(&barang)
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  barang,
		}
	}

	end := FinalTake + MAXTAKE
	if end > len(keys) {
		end = len(keys)
	}

	for _, key := range keys[FinalTake:end] {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()

			data, err := rds.HGetAll(ctx, k).Result()
			if err != nil || len(data) == 0 {
				return
			}

			id_barang, _ := strconv.Atoi(data["id_barang_induk"])
			id_seller, _ := strconv.Atoi(data["id_seller_barang_induk"])
			viewed, _ := strconv.Atoi(data["viewed_barang_induk"])
			likes, _ := strconv.Atoi(data["likes_barang_induk"])
			total_komentar, _ := strconv.Atoi(data["total_komentar_barang_induk"])
			harga_kategori, _ := strconv.Atoi(data["harga"])

			b := models.BarangInduk{
				ID:               int32(id_barang),
				SellerID:         int32(id_seller),
				NamaBarang:       data["nama_barang_induk"],
				JenisBarang:      data["jenis_barang_induk"],
				OriginalKategori: data["original_kategori"],
				Deskripsi:        data["deskripsi_barang_induk"],
				TanggalRilis:     data["tanggal_rilis_barang_induk"],
				Viewed:           int32(viewed),
				Likes:            int32(likes),
				TotalKomentar:    int32(total_komentar),
				HargaKategoris:   int32(harga_kategori),
			}

			mu.Lock()
			if len(barang) < MAXTAKE {
				barang = append(barang, b)
			}
			mu.Unlock()
		}(key)
	}

	wg.Wait()
	length := len(barang)
	if length < MAXTAKE {
		var Hasil []models.BarangInduk
		db.Model(&models.BarangInduk{}).Order("viewed DESC, likes DESC").Offset(FinalTake).Limit(MAXTAKE - length).Find(&Hasil)
		barang = append(barang, Hasil...)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangJenis(FinalTake int, ctx context.Context, db *gorm.DB, rds *redis.Client, jenis string) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangJenis"

	var barang []models.BarangInduk

	// Ambil semua key dari Redis
	keys, err := rds.SMembers(ctx, fmt.Sprintf("jenis_%s_barang", helper.ConvertJenisBarangReverse(jenis))).Result()
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Gagal mengambil data dari cache",
		}
	}

	if len(keys) == 0 {
		// Ambil dari DB langsung jika cache kosong
		var id_barang []int32
		if err := db.Model(&models.BarangInduk{}).
			Where("jenis_barang = ?", jenis).
			Order("viewed DESC, likes DESC").
			Limit(MAXTAKE).
			Pluck("id", &id_barang).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  barang,
			}
		}

		if len(id_barang) > 0 {
			var hasil []models.BarangInduk
			if err := db.Where("id IN ?", id_barang).Find(&hasil).Error; err != nil {
				return &response.ResponseForm{
					Status:   http.StatusInternalServerError,
					Services: services,
					Payload:  barang,
				}
			}
			barang = hasil
		}

	} else {
		// Ambil slice sesuai paging
		if FinalTake > len(keys) {
			var Hasil []models.BarangInduk

			if err := db.Model(&models.BarangInduk{}).Order("viewed DESC, likes DESC").Offset(FinalTake).Where(&models.BarangInduk{JenisBarang: jenis}).Limit(MAXTAKE).Find(&Hasil).Error; err != nil {
				return &response.ResponseForm{
					Status:   http.StatusNotFound,
					Services: services,
					Payload:  []models.BarangInduk{},
				}
			}

			barang = append(barang, Hasil...)

			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: services,
				Payload:  barang,
			}
		}
		end := FinalTake + MAXTAKE
		if len(keys) < end {
			end = len(keys)
		}

		if FinalTake >= end {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  barang,
			}
		}

		for _, key := range keys[FinalTake:end] {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				data, err := rds.HGetAll(ctx, k).Result()
				if err != nil || len(data) == 0 {
					return
				}

				id_barang, _ := strconv.Atoi(data["id_barang_induk"])
				id_seller, _ := strconv.Atoi(data["id_seller_barang_induk"])
				viewed, _ := strconv.Atoi(data["viewed_barang_induk"])
				likes, _ := strconv.Atoi(data["likes_barang_induk"])
				total_komentar, _ := strconv.Atoi(data["total_komentar_barang_induk"])
				harga_kategori, _ := strconv.Atoi(data["harga"])

				b := models.BarangInduk{
					ID:               int32(id_barang),
					SellerID:         int32(id_seller),
					NamaBarang:       data["nama_barang_induk"],
					JenisBarang:      data["jenis_barang_induk"],
					OriginalKategori: data["original_kategori"],
					Deskripsi:        data["deskripsi_barang_induk"],
					TanggalRilis:     data["tanggal_rilis_barang_induk"],
					Viewed:           int32(viewed),
					Likes:            int32(likes),
					TotalKomentar:    int32(total_komentar),
					HargaKategoris:   int32(harga_kategori),
				}

				mu.Lock()
				barang = append(barang, b)
				mu.Unlock()
			}(key)
		}
		wg.Wait()

		if len(barang) < MAXTAKE {
			var Hasil []models.BarangInduk
			if err := db.Model(&models.BarangInduk{}).Where(&models.BarangInduk{
				JenisBarang: jenis}).
				Order("viewed DESC, likes DESC").
				Limit(MAXTAKE - len(barang)).
				Find(&Hasil).Error; err != nil {
				fmt.Println("Gagal Fallback")
			} else {
				barang = append(barang, Hasil...)
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangSeller(FinalTake int, ctx context.Context, db *gorm.DB, rds *redis.Client, sellerId int32) *response.ResponseForm {
	services := "AmbilBarangSeller"

	if sellerId == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal Mengambil Data Barang seller ini",
		}
	}

	keys, _ := rds.SMembers(ctx, fmt.Sprintf("barang_seller:%v", sellerId)).Result()

	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Gagal Keys Tidak Valid",
		}
	}

	var barang []models.BarangInduk
	var wg sync.WaitGroup
	if FinalTake > len(keys) {
		var Hasil []models.BarangInduk

		if err := db.Model(&models.BarangInduk{}).Order("viewed DESC, likes DESC").Offset(FinalTake).Where(&models.BarangInduk{SellerID: sellerId}).Limit(MAXTAKE).Find(&Hasil).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  []models.BarangInduk{},
			}
		}
		barang = append(barang, Hasil...)
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  barang,
		}
	}
	end := FinalTake + MAXTAKE
	if end > len(keys) {
		end = len(keys)
	}

	for _, key := range keys[FinalTake:end] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, _ := rds.HGetAll(ctx, key).Result()
			if len(data) == 0 {
				return
			}

			id_barang, _ := strconv.Atoi(data["id_barang_induk"])
			id_seller, _ := strconv.Atoi(data["id_seller_barang_induk"])
			viewed, _ := strconv.Atoi(data["viewed_barang_induk"])
			likes, _ := strconv.Atoi(data["likes_barang_induk"])
			total_komentar, _ := strconv.Atoi(data["total_komentar_barang_induk"])
			harga_kategori, _ := strconv.Atoi(data["harga"])

			b := models.BarangInduk{
				ID:               int32(id_barang),
				SellerID:         int32(id_seller),
				NamaBarang:       data["nama_barang_induk"],
				JenisBarang:      data["jenis_barang_induk"],
				OriginalKategori: data["original_kategori"],
				Deskripsi:        data["deskripsi_barang_induk"],
				TanggalRilis:     data["tanggal_rilis_barang_induk"],
				Viewed:           int32(viewed),
				Likes:            int32(likes),
				TotalKomentar:    int32(total_komentar),
				HargaKategoris:   int32(harga_kategori),
			}
			barang = append(barang, b)
		}()
	}

	wg.Wait()

	if len(barang) < MAXTAKE {
		var Hasil []models.BarangInduk
		if err := db.Model(&models.BarangInduk{}).Where(&models.BarangInduk{
			SellerID: sellerId}).
			Order("viewed DESC, likes DESC").
			Limit(MAXTAKE - len(barang)).
			Find(&Hasil).Error; err != nil {
			fmt.Println("Gagal Fallback")
		} else {
			barang = append(barang, Hasil...)
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangNama(FinalTake int, ctx context.Context, rds *redis.Client, db *gorm.DB, Nama_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "Ambil Barang Nama"
	var wg sync.WaitGroup
	var mu sync.Mutex
	var barang []models.BarangInduk
	var idBarangList []search_engine.BarangInduk

	if Nama_Barang == "" {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	// 1. Cari di Meilisearch
	searchRes, err := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{
		Limit:  MAXTAKE,
		Sort:   []string{"viewed_barang_induk:desc", "likes_barang_induk:desc"},
		Offset: int64(FinalTake),
	})
	if err != nil {
		fmt.Println("❌ Gagal search Meilisearch:", err)
	}

	if searchRes != nil && len(searchRes.Hits) > 0 {
		hitsJSON, _ := json.Marshal(searchRes.Hits)
		if err := json.Unmarshal(hitsJSON, &idBarangList); err != nil {
			fmt.Println("❌ Gagal unmarshal Meilisearch hits:", err)
		} else {
			fmt.Println("✅ Data dari Meili ditemukan")

			for _, dataID := range idBarangList {
				// Pastikan IdBarang valid
				if dataID.IdBarang == 0 {
					continue
				}

				wg.Add(1)
				go func(id int64) {
					defer wg.Done()
					data, err := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", id)).Result()
					if err != nil || len(data) == 0 {
						return
					}

					idBarangInt, err := strconv.Atoi(data["id_barang_induk"])
					if err != nil || idBarangInt == 0 {
						return
					}
					idSellerInt, _ := strconv.Atoi(data["id_seller_barang_induk"])
					viewedInt, _ := strconv.Atoi(data["viewed_barang_induk"])
					likesInt, _ := strconv.Atoi(data["likes_barang_induk"])
					totalKomentarInt, _ := strconv.Atoi(data["total_komentar_barang_induk"])
					hargaKategoriInt, _ := strconv.Atoi(data["harga"])

					b := models.BarangInduk{
						ID:               int32(idBarangInt),
						SellerID:         int32(idSellerInt),
						NamaBarang:       data["nama_barang_induk"],
						JenisBarang:      data["jenis_barang_induk"],
						OriginalKategori: data["original_kategori"],
						Deskripsi:        data["deskripsi_barang_induk"],
						TanggalRilis:     data["tanggal_rilis_barang_induk"],
						Viewed:           int32(viewedInt),
						Likes:            int32(likesInt),
						TotalKomentar:    int32(totalKomentarInt),
						HargaKategoris:   int32(hargaKategoriInt),
					}

					mu.Lock()
					barang = append(barang, b)
					mu.Unlock()
				}(dataID.IdBarang)
			}

			wg.Wait()
		}
	}

	// 2. Fallback ke DB jika Meili kosong atau kurang
	if len(barang) < MAXTAKE {
		var dbResult []models.BarangInduk
		idList := MatchSearch(barang, idBarangList)
		if len(idList) > 0 {
			if err := db.Model(&models.BarangInduk{}).
				Where("id IN ?", idList).
				Find(&dbResult).Error; err != nil {
				fmt.Println("❌ Gagal Mengambil Data DB:", err)
			} else {
				barang = append(barang, dbResult...)
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangNamaDanJenis(FinalTake int, ctx context.Context, rds *redis.Client, db *gorm.DB, Nama_Barang, Jenis_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangNamaDanJenis"
	var barang []models.BarangInduk
	var id_barang []search_engine.BarangInduk

	// Cari di Meilisearch
	searchRes, err := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("jenis_barang_induk = '%s'", Jenis_Barang),
		Sort:   []string{"viewed_barang_induk:desc", "likes_barang_induk:desc"},
		Offset: int64(FinalTake),
		Limit:  MAXTAKE,
	})
	if err != nil {
		fmt.Println("❌ Gagal search Meilisearch:", err)
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	if len(searchRes.Hits) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	hitsJSON, err := json.Marshal(searchRes.Hits)
	if err != nil {
		fmt.Println("❌ Gagal marshal Meili hits:", err)
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	if err := json.Unmarshal(hitsJSON, &id_barang); err != nil {
		fmt.Println("❌ Gagal unmarshal Meili hits:", err)
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	fmt.Println("✅ Data dari Meili ditemukan")

	// Ambil detail dari Redis
	for _, data_id := range id_barang {
		if data_id.IdBarang == 0 {
			continue
		}

		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			data, err := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", id)).Result()
			if err != nil || len(data) == 0 {
				return
			}

			id_barang_int, err := strconv.Atoi(data["id_barang_induk"])
			if err != nil || id_barang_int == 0 {
				return
			}
			id_seller_int, _ := strconv.Atoi(data["id_seller_barang_induk"])
			viewed_int, _ := strconv.Atoi(data["viewed_barang_induk"])
			likes_int, _ := strconv.Atoi(data["likes_barang_induk"])
			total_komentar_int, _ := strconv.Atoi(data["total_komentar_barang_induk"])
			harga_kategori_int, _ := strconv.Atoi(data["harga"])

			b := models.BarangInduk{
				ID:               int32(id_barang_int),
				SellerID:         int32(id_seller_int),
				NamaBarang:       data["nama_barang_induk"],
				JenisBarang:      data["jenis_barang_induk"],
				OriginalKategori: data["original_kategori"],
				Deskripsi:        data["deskripsi_barang_induk"],
				TanggalRilis:     data["tanggal_rilis_barang_induk"],
				Viewed:           int32(viewed_int),
				Likes:            int32(likes_int),
				TotalKomentar:    int32(total_komentar_int),
				HargaKategoris:   int32(harga_kategori_int),
			}

			mu.Lock()
			barang = append(barang, b)
			mu.Unlock()
		}(data_id.IdBarang)
	}

	wg.Wait()

	if len(barang) < MAXTAKE {
		var dbResult []models.BarangInduk
		idList := MatchSearch(barang, id_barang)
		if len(idList) > 0 {
			if err := db.Model(&models.BarangInduk{}).
				Where("id IN ?", idList).
				Find(&dbResult).Error; err != nil {
				fmt.Println("❌ Gagal Mengambil Data DB:", err)
			} else {
				barang = append(barang, dbResult...)
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangNamaDanSeller(FinalTake int, ctx context.Context, rds *redis.Client, db *gorm.DB, id_seller int32, Nama_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangNamaDanSeller"
	var barang []models.BarangInduk
	var id_barang []search_engine.BarangInduk

	searchRes, err := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("id_seller_barang_induk = '%v'", id_seller),
		Limit:  MAXTAKE,
		Offset: int64(FinalTake),
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	if len(searchRes.Hits) > 0 {
		hitsjson, _ := json.Marshal(searchRes.Hits)

		_ = json.Unmarshal(hitsjson, &id_barang)

		fmt.Println("Jalan Dari Meili")
	} else {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  []models.BarangInduk{},
		}
	}

	for _, barang_key := range id_barang {
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, _ := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", barang_key.IdBarang)).Result()
			if len(data) == 0 {
				return
			}

			id_barang, _ := strconv.Atoi(data["id_barang_induk"])
			id_seller, _ := strconv.Atoi(data["id_seller_barang_induk"])
			viewed, _ := strconv.Atoi(data["viewed_barang_induk"])
			likes, _ := strconv.Atoi(data["likes_barang_induk"])
			total_komentar, _ := strconv.Atoi(data["total_komentar_barang_induk"])
			harga_kategori, _ := strconv.Atoi(data["harga"])

			b := models.BarangInduk{
				ID:               int32(id_barang),
				SellerID:         int32(id_seller),
				NamaBarang:       data["nama_barang_induk"],
				JenisBarang:      data["jenis_barang_induk"],
				OriginalKategori: data["original_kategori"],
				Deskripsi:        data["deskripsi_barang_induk"],
				TanggalRilis:     data["tanggal_rilis_barang_induk"],
				Viewed:           int32(viewed),
				Likes:            int32(likes),
				TotalKomentar:    int32(total_komentar),
				HargaKategoris:   int32(harga_kategori),
			}
			mu.Lock()
			barang = append(barang, b)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(barang) < MAXTAKE {
		var dbResult []models.BarangInduk
		idList := MatchSearch(barang, id_barang)
		if len(idList) > 0 {
			if err := db.Model(&models.BarangInduk{}).
				Where("id IN ?", idList).
				Find(&dbResult).Error; err != nil {
				fmt.Println("❌ Gagal Mengambil Data DB:", err)
			} else {
				barang = append(barang, dbResult...)
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
	}
}

func AmbilBarangJenisDanSeller(FinalTake int, ctx context.Context, rds *redis.Client, db *gorm.DB, id_seller int32, jenis string) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangJenisDanSeller"
	var barang []models.BarangInduk

	// Ambil semua ID barang seller dari Redis
	id_barang_rds, err := rds.SMembers(ctx, fmt.Sprintf("barang_seller:%v", id_seller)).Result()
	if err != nil {
		fmt.Println("❌ Gagal ambil barang dari Redis:", err)
		id_barang_rds = []string{}
	}

	// Jika FinalTake melebihi jumlah barang di Redis, fallback langsung ke DB
	if FinalTake >= len(id_barang_rds) {
		if err := db.Model(&models.BarangInduk{}).
			Where(&models.BarangInduk{SellerID: id_seller, JenisBarang: jenis}).
			Offset(FinalTake).
			Order("viewed DESC, likes DESC").
			Limit(MAXTAKE).
			Find(&barang).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  []models.BarangInduk{},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  barang,
		}
	}

	// Ambil range slice
	end := FinalTake + MAXTAKE
	if end > len(id_barang_rds) {
		end = len(id_barang_rds)
	}

	// Ambil data dari Redis dengan goroutine
	for _, key := range id_barang_rds[FinalTake:end] {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			data, err := rds.HGetAll(ctx, fmt.Sprintf("barang:%s", key)).Result()
			if err != nil || len(data) == 0 {
				return
			}

			if data["jenis_barang_induk"] != jenis {
				return
			}

			id_barang_int, err := strconv.Atoi(data["id_barang_induk"])
			if err != nil || id_barang_int == 0 {
				return
			}
			id_seller_int, _ := strconv.Atoi(data["id_seller_barang_induk"])
			viewed_int, _ := strconv.Atoi(data["viewed_barang_induk"])
			likes_int, _ := strconv.Atoi(data["likes_barang_induk"])
			total_komentar_int, _ := strconv.Atoi(data["total_komentar_barang_induk"])
			harga_kategori_int, _ := strconv.Atoi(data["harga"])

			b := models.BarangInduk{
				ID:               int32(id_barang_int),
				SellerID:         int32(id_seller_int),
				NamaBarang:       data["nama_barang_induk"],
				JenisBarang:      data["jenis_barang_induk"],
				OriginalKategori: data["original_kategori"],
				Deskripsi:        data["deskripsi_barang_induk"],
				TanggalRilis:     data["tanggal_rilis_barang_induk"],
				Viewed:           int32(viewed_int),
				Likes:            int32(likes_int),
				TotalKomentar:    int32(total_komentar_int),
				HargaKategoris:   int32(harga_kategori_int),
			}

			mu.Lock()
			barang = append(barang, b)
			mu.Unlock()
		}(key)
	}

	wg.Wait()

	// Fallback ke DB jika jumlah barang kurang dari MAXTAKE
	if len(barang) < MAXTAKE {
		var Hasil []models.BarangInduk
		limitDB := MAXTAKE - len(barang)

		err := db.Model(&models.BarangInduk{}).
			Where("jenis_barang = ? AND id_seller = ? AND id NOT IN ?", jenis, id_seller, NotIn(barang)).
			Limit(limitDB).
			Find(&Hasil).Error
		if err == nil && len(Hasil) > 0 {
			barang = append(barang, Hasil...)
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////
// Ambil Barang Di Tap
// /////////////////////////////////////////////////////////////////////////////////////////////////

func AmbilDataBarangInduk(ctx context.Context, id_barang_induk int32, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilDataBarangInduk"

	if id_barang_induk == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Data Barang Induk Tidak Ditemukan",
		}
	}

	var barangInduk models.BarangInduk
	var kategoriBarang []response_barang.KategoriBarangDiambil

	cacheKey := fmt.Sprintf("barang:%v", id_barang_induk)
	dataRedis, _ := rds.HGetAll(ctx, cacheKey).Result()

	if len(dataRedis) > 0 {
		idBarang, _ := strconv.Atoi(dataRedis["id_barang_induk"])
		idSeller, _ := strconv.Atoi(dataRedis["id_seller_barang_induk"])
		viewed, _ := strconv.Atoi(dataRedis["viewed_barang_induk"])
		likes, _ := strconv.Atoi(dataRedis["likes_barang_induk"])
		totalKomentar, _ := strconv.Atoi(dataRedis["total_komentar_barang_induk"])
		hargaKategori, _ := strconv.Atoi(dataRedis["harga"])

		barangInduk = models.BarangInduk{
			ID:               int32(idBarang),
			SellerID:         int32(idSeller),
			NamaBarang:       dataRedis["nama_barang_induk"],
			JenisBarang:      dataRedis["jenis_barang_induk"],
			OriginalKategori: dataRedis["original_kategori"],
			Deskripsi:        dataRedis["deskripsi_barang_induk"],
			TanggalRilis:     dataRedis["tanggal_rilis_barang_induk"],
			Viewed:           int32(viewed),
			Likes:            int32(likes),
			TotalKomentar:    int32(totalKomentar),
			HargaKategoris:   int32(hargaKategori),
		}
	} else {
		var hasilDB models.BarangInduk

		if err := db.Model(models.BarangInduk{}).
			Select("*").
			Where("id = ?", id_barang_induk).
			Take(&hasilDB).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  "Server Sedang Sibuk",
			}
		}

		var hargaBarang int32
		if err := db.Model(models.KategoriBarang{}).
			Select("harga").
			Where(&models.KategoriBarang{
				IdBarangInduk: hasilDB.ID,
				Nama:          hasilDB.OriginalKategori,
			}).
			Take(&hargaBarang).Error; err != nil {
			fmt.Println("❌ Gagal ambil harga kategori:", err)
		}

		barangInduk = models.BarangInduk{
			ID:               hasilDB.ID,
			SellerID:         hasilDB.SellerID,
			NamaBarang:       hasilDB.NamaBarang,
			JenisBarang:      hasilDB.JenisBarang,
			OriginalKategori: hasilDB.OriginalKategori,
			Deskripsi:        hasilDB.Deskripsi,
			TanggalRilis:     hasilDB.TanggalRilis,
			Viewed:           hasilDB.Viewed,
			Likes:            hasilDB.Likes,
			TotalKomentar:    hasilDB.TotalKomentar,
			HargaKategoris:   hargaBarang,
		}
	}

	var kategoriCheck []models.KategoriBarang
	if err := db.Model(models.KategoriBarang{}).
		Where("id_barang_induk = ?", id_barang_induk).
		Find(&kategoriCheck).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Coba lain waktu, server sedang sibuk",
		}
	}

	if len(kategoriCheck) > 0 {
		hitsjson, _ := json.Marshal(kategoriCheck)
		_ = json.Unmarshal(hitsjson, &kategoriBarang)
	}

	found := false
	for i, kategori := range kategoriBarang {
		wg.Add(1)
		go func(idx int, kat response_barang.KategoriBarangDiambil) {
			defer wg.Done()
			if kat.Nama == barangInduk.OriginalKategori {
				mu.Lock()
				if !found {
					kategoriBarang[idx].Original = true
					found = true
				}
				mu.Unlock()
			}
		}(i, kategori)
	}
	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang.ResponseUserBarangInduk{
			DataBarangInduk: barangInduk,
			DataKategori:    kategoriBarang,
		},
	}
}
