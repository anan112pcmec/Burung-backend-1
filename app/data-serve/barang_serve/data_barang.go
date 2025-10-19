package barang_serve

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

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

func AmbilBarangJenis(ctx context.Context, db *gorm.DB, rds *redis.Client, jenis string) *response.ResponseForm {

	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangJenis"

	var barang []models.BarangInduk
	keys, _ := rds.SMembers(ctx, fmt.Sprintf("jenis_%s_barang", helper.ConvertJenisBarangReverse(jenis))).Result()
	if len(keys) == 0 {
		var id_barang []int32
		if err := db.Model(models.BarangInduk{}).
			Where(models.BarangInduk{JenisBarang: jenis}).
			Order("RANDOM()").
			Limit(30).
			Pluck("id", &id_barang).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  "Server Sedang Sibuk",
			}
		}

		for _, ids := range id_barang {
			var hasil models.BarangInduk

			if err := db.Model(models.BarangInduk{}).Select("*").
				Where(&models.BarangInduk{ID: ids}).
				Find(&hasil).Error; err != nil {

				return &response.ResponseForm{
					Status:   http.StatusInternalServerError,
					Services: services,
					Payload:  "Server Sedang Sibuk",
				}
			}

			barang = append(barang, hasil)
		}

	} else {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
		if len(keys) > 30 {
			keys = keys[:30]
		}

		for _, key := range keys {
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

				mu.Lock() // kunci sebelum append
				barang = append(barang, b)
				mu.Unlock()
			}()
		}

		wg.Wait()

	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangSeller(ctx context.Context, rds *redis.Client, sellerId int32) *response.ResponseForm {
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

	for _, key := range keys {
		wg.Add(1)
		go func() { // bawa key ke parameter biar aman
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

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangNama(ctx context.Context, rds *redis.Client, db *gorm.DB, Nama_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {

	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "Ambil Barang Nama"
	if Nama_Barang == "" {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Gagal Masukan Nama Barang Yang Hendak Di cari",
		}
	}

	searchRes, _ := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{})

	var barang []models.BarangInduk

	if len(searchRes.Hits) > 0 {
		var id_barang []models.BarangInduk
		hitsjson, _ := json.Marshal(searchRes.Hits)

		_ = json.Unmarshal(hitsjson, &id_barang)

		fmt.Println("Jalan Dari Meili")

		for _, data_id := range id_barang {
			if len(barang) == 30 {
				break
			}
			wg.Add(1)
			go func() {
				defer wg.Done()

				data, _ := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", data_id.ID)).Result()
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
	} else {
		var idBarang []models.BarangInduk
		if err := db.Model(&models.BarangInduk{}).
			Where("nama_barang LIKE ?", "'%"+Nama_Barang+"%'").
			Select("id").
			Order("RANDOM()").
			Limit(30).
			Pluck("id", &idBarang).Error; err != nil {

			fmt.Println("❌ Gagal Mengambil Data:", err)
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  "Barang Yang Dicari Tidak Ada!",
			}
		}

		if len(idBarang) > 0 {
			for _, data_id := range idBarang {
				wg.Add(1)
				go func(d models.BarangInduk) {
					defer wg.Done()
					data, _ := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", d.ID)).Result()
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
				}(data_id)
			}

			wg.Wait()

		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
	}
}

func AmbilBarangNamaDanJenis(ctx context.Context, rds *redis.Client, db *gorm.DB, Nama_Barang, Jenis_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangNamaDanJenis"
	var barang []models.BarangInduk
	var id_barang []search_engine.BarangInduk

	searchRes, _ := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("jenis_barang_induk = '%s'", Jenis_Barang),
	})

	if len(searchRes.Hits) > 0 {
		hitsjson, _ := json.Marshal(searchRes.Hits)
		_ = json.Unmarshal(hitsjson, &id_barang)

		fmt.Println("Jalan Dari Meili")
		fmt.Println(id_barang)
	} else {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Tidak Ditemukan Barang Dengan Nama Dan Jenis Tersebut",
		}
	}

	var hasil int

	if len(id_barang) > 30 {
		hasil = 30
	} else {
		hasil = len(id_barang)
	}

	for _, data_id := range id_barang[:hasil] {

		wg.Add(1)
		go func(data_id search_engine.BarangInduk) {
			defer wg.Done()

			data, _ := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", data_id.IdBarang)).Result()
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
		}(data_id)
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
	}
}

func AmbilBarangNamaDanSeller(ctx context.Context, rds *redis.Client, db *gorm.DB, id_seller int32, Nama_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {

	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangNamaDanSeller"
	var barang []models.BarangInduk
	var id_barang []models.BarangInduk

	searchRes, _ := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("id_seller_barang_induk = '%v'", id_seller),
	})

	if len(searchRes.Hits) > 0 {
		hitsjson, _ := json.Marshal(searchRes.Hits)

		_ = json.Unmarshal(hitsjson, &id_barang)

		fmt.Println("Jalan Dari Meili")
	} else {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Seller Ini Tidak Menjual Barang Demikian",
		}
	}

	for _, barang_key := range id_barang {
		if len(barang) == 30 {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, _ := rds.HGetAll(ctx, fmt.Sprintf("barang:%v", barang_key.ID)).Result()
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

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
	}
}

func AmbilBarangJenisDanSeller(ctx context.Context, rds *redis.Client, db *gorm.DB, id_seller int32, jenis string) *response.ResponseForm {

	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilBarangJenisDanSeller"
	var barang []models.BarangInduk

	id_barang_rds, _ := rds.SRandMemberN(ctx, fmt.Sprintf("barang_seller:%v", id_seller), 1000).Result()

	if len(id_barang_rds) > 0 {
		for _, keys := range id_barang_rds {
			if len(barang) == 30 {
				break
			}
			wg.Add(1)
			go func() {
				defer wg.Done()

				data, _ := rds.HGetAll(ctx, fmt.Sprintf("barang:%s", keys)).Result()
				if len(data) == 0 {
					return
				}

				if data["jenis_barang_induk"] != jenis {
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
	}
	// else {
	// 	if err := db.Model(models.BarangInduk{}).Where(models.BarangInduk)
	// }

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
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
