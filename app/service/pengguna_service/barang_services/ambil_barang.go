package pengguna_service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services/response_barang_user"
)

func AmbilRandomBarang(ctx context.Context, rds *redis.Client) *response.ResponseForm {
	services := "AmbilRandomBarang"
	const maxBarang = 30
	const scanCount = 1000
	var cursor uint64
	var barang []response_barang_user.ResponseUserBarang

	rand.Seed(time.Now().UnixNano())
	keysTaken := 0

	for keysTaken < maxBarang {
		keysBatch, nextCursor, err := rds.Scan(ctx, cursor, "barang:*", scanCount).Result()
		if err != nil {
			fmt.Println("❌ Gagal scan redis:", err)
			break
		}

		if len(keysBatch) > 0 {
			rand.Shuffle(len(keysBatch), func(i, j int) { keysBatch[i], keysBatch[j] = keysBatch[j], keysBatch[i] })

			for _, key := range keysBatch {
				if keysTaken >= maxBarang {
					break
				}

				data, err := rds.HGetAll(ctx, key).Result()
				if err != nil || len(data) == 0 {
					continue
				}

				idBarang, _ := strconv.Atoi(data["id_barang_induk"])
				idSeller, _ := strconv.Atoi(data["id_seller_barang_induk"])
				viewed, _ := strconv.Atoi(data["viewed_barang_induk"])
				likes, _ := strconv.Atoi(data["likes_barang_induk"])
				totalKomentar, _ := strconv.Atoi(data["total_komentar_barang_induk"])

				b := response_barang_user.ResponseUserBarang{
					BarangInduk: models.BarangInduk{
						ID:               int32(idBarang),
						SellerID:         int32(idSeller),
						NamaBarang:       data["nama_barang_induk"],
						JenisBarang:      data["jenis_barang_induk"],
						OriginalKategori: data["original_kategori"],
						Deskripsi:        data["deskripsi_barang_induk"],
						TanggalRilis:     data["tanggal_rilis_barang_induk"],
						Viewed:           int32(viewed),
						Likes:            int32(likes),
						TotalKomentar:    int32(totalKomentar),
					},
					Harga: data["harga"],
				}

				barang = append(barang, b)
				keysTaken++
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangJenis(ctx context.Context, rds *redis.Client, jenis string) *response.ResponseForm {
	services := "AmbilBarangJenis"

	keys, _ := rds.SMembers(ctx, fmt.Sprintf("jenis_%s_barang", helper.ConvertJenisBarangReverse(jenis))).Result()
	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []response_barang_user.ResponseUserBarang{},
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	if len(keys) > 30 {
		keys = keys[:30]
	}

	var barang []response_barang_user.ResponseUserBarang
	var wg sync.WaitGroup

	for _, key := range keys {
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

			barang = append(barang, b)
		}(key)
	}

	wg.Wait()

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
