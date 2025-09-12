package pengguna_service

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

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services/response_barang_user"
)

var wg sync.WaitGroup
var mu sync.Mutex

func AmbilRandomBarang(ctx context.Context, rds *redis.Client) *response.ResponseForm {
	services := "AmbilRandomBarang"
	const maxBarang = 30

	keys, _ := rds.SRandMemberN(ctx, "barang_keys", int64(maxBarang)).Result()
	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal, mungkin server sedang sibuk atau maintenance. Coba lagi nanti.",
		}
	}

	// Channel buat nampung hasil dari goroutine
	resultChan := make(chan response_barang_user.ResponseUserBarang, len(keys))

	for i, key := range keys {
		if i == maxBarang {
			break
		}
		wg.Add(1)
		go func(key string) {
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

			b := response_barang_user.ResponseUserBarang{
				BarangInduk: models.BarangInduk{
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
				},
				Harga: data["harga"],
			}
			resultChan <- b
		}(key)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	barang := make([]response_barang_user.ResponseUserBarang, 0, len(keys))
	for b := range resultChan {
		barang = append(barang, b)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangJenis(ctx context.Context, db *gorm.DB, rds *redis.Client, jenis string) *response.ResponseForm {
	services := "AmbilBarangJenis"

	var barang []response_barang_user.ResponseUserBarang
	keys, _ := rds.SMembers(ctx, fmt.Sprintf("jenis_%s_barang", helper.ConvertJenisBarangReverse(jenis))).Result()
	if len(keys) == 0 {
		var id_barang []int32
		if err := db.Model(models.BarangInduk{}).
			Where(models.BarangInduk{JenisBarang: jenis}).
			Order("RANDOM()"). // ✅ pakai RANDOM() untuk PostgreSQL
			Limit(30).
			Pluck("id", &id_barang).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  "Server Sedang Sibuk",
			}
		}

		for _, ids := range id_barang {
			var hasil response_barang_user.ResponseUserBarang

			if err := db.Model(models.BarangInduk{}).Select("*").
				Where("id", ids).
				Find(&hasil).Error; err != nil {

				return &response.ResponseForm{
					Status:   http.StatusInternalServerError,
					Services: services,
					Payload:  "Server Sedang Sibuk",
				}
			}

			var hargabarangnya int32
			if err := db.Model(models.KategoriBarang{}).Select("harga").Where(models.KategoriBarang{IdBarangInduk: hasil.ID, Nama: hasil.OriginalKategori}).Take(&hargabarangnya).Error; err != nil {
				fmt.Println("gagal konversi harga")
			}
			hargafinal := strconv.Itoa(int(hargabarangnya))
			b := response_barang_user.ResponseUserBarang{
				BarangInduk: models.BarangInduk{
					ID:               hasil.ID,
					SellerID:         hasil.SellerID,
					NamaBarang:       hasil.NamaBarang,
					JenisBarang:      hasil.JenisBarang,
					OriginalKategori: hasil.OriginalKategori,
					Deskripsi:        hasil.Deskripsi,
					Viewed:           hasil.Viewed,
					Likes:            hasil.Likes,
					TotalKomentar:    hasil.TotalKomentar,
				},
				Harga: hargafinal,
			}

			barang = append(barang, b)

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

				b := response_barang_user.ResponseUserBarang{
					BarangInduk: models.BarangInduk{
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
					},
					Harga: data["harga"],
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

	var barang []response_barang_user.ResponseUserBarang
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

			b := response_barang_user.ResponseUserBarang{
				BarangInduk: models.BarangInduk{
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
				},
				Harga: data["harga"],
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
	services := "Ambil Barang Nama"
	if Nama_Barang == "" {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Gagal Masukan Nama Barang Yang Hendak Di cari",
		}
	}

	searchRes, _ := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{})

	var barang []response_barang_user.ResponseUserBarang
	var id_barang []models.BarangInduk

	if len(searchRes.Hits) > 0 {
		hitsjson, _ := json.Marshal(searchRes.Hits)

		_ = json.Unmarshal(hitsjson, &id_barang)

		fmt.Println("Jalan Dari Meili")

		for _, data_id := range id_barang {
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

				b := response_barang_user.ResponseUserBarang{
					BarangInduk: models.BarangInduk{
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
					},
					Harga: data["harga"],
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
			Where("nama_barang LIKE ?", "%"+Nama_Barang+"%").
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

		for _, data_id := range id_barang {
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

				b := response_barang_user.ResponseUserBarang{
					BarangInduk: models.BarangInduk{
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
					},
					Harga: data["harga"],
				}

				mu.Lock()
				barang = append(barang, b)
				mu.Unlock()
			}()
		}

		wg.Wait()

	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
	}
}

func AmbilBarangNamaDanJenis(ctx context.Context, rds *redis.Client, db *gorm.DB, Nama_Barang, Jenis_Barang string, SE meilisearch.ServiceManager) *response.ResponseForm {
	services := "AmbilBarangNamaDanJenis"
	var barang []response_barang_user.ResponseUserBarang
	var id_barang []models.BarangInduk

	searchRes, _ := SE.Index("barang_induk_all").Search(Nama_Barang, &meilisearch.SearchRequest{
		Filter: fmt.Sprintf("jenis_barang_induk = '%s'", Jenis_Barang),
		Limit:  30,
	})

	if len(searchRes.Hits) > 0 {
		hitsjson, _ := json.Marshal(searchRes.Hits)

		_ = json.Unmarshal(hitsjson, &id_barang)

		fmt.Println("Jalan Dari Meili")

		for i, data_id := range id_barang {
			if i == 30 {
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

				b := response_barang_user.ResponseUserBarang{
					BarangInduk: models.BarangInduk{
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
					},
					Harga: data["harga"],
				}
				mu.Lock()
				barang = append(barang, b)
				mu.Unlock()
			}()
		}

		wg.Wait()
	} else {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Tidak Ditemukan Barang Dengan Nama Dan Jenis Tersebut",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  &barang,
	}
}
