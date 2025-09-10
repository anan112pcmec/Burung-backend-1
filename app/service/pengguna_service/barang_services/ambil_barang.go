package pengguna_service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services/response_barang_user"
)

func AmbilRandomBarang(rds *redis.Client) *response.ResponseForm {
	services := "AmbilRandomBarang"
	ctx := context.Background()

	var cursor uint64
	var keys []string
	for {
		var batch []string
		var err error
		batch, cursor, err = rds.Scan(ctx, cursor, "barang:*", 1000).Result()
		if err != nil {
			fmt.Println("❌ Gagal scan redis:", err)
			break
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	if len(keys) > 30 {
		keys = keys[:30]
	}

	var barang []response_barang_user.ResponseUserBarang

	for _, key := range keys {
		data, err := rds.HGetAll(ctx, key).Result()
		if err != nil {
			fmt.Println("❌ Gagal ambil HGetAll:", err)
			continue
		}
		if len(data) == 0 {
			continue
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
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}

func AmbilBarangJenis(rds *redis.Client, jenis string) *response.ResponseForm {
	services := "AmbilBarangJenis"
	ctx := context.Background()

	// Key list redis (isi nya kumpulan key barang)
	listKey := fmt.Sprintf("jenis_%s_barang", helper.ConvertJenisBarangReverse(jenis))

	// Ambil semua key barang dalam list
	keys, err := rds.SMembers(ctx, listKey).Result()
	if err != nil {
		fmt.Println("❌ Gagal ambil SMEMBERS:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "gagal ambil data barang dari redis",
		}
	}

	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  []response_barang_user.ResponseUserBarang{},
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	// Maksimal 30
	finalKeys := keys
	if len(finalKeys) > 30 {
		finalKeys = finalKeys[:30]
	}

	var barang []response_barang_user.ResponseUserBarang

	for _, key := range finalKeys {
		data, err := rds.HGetAll(ctx, key).Result()
		if err != nil {
			fmt.Printf("❌ Gagal HGetAll key=%s: %v\n", key, err)
			continue
		}
		if len(data) == 0 {
			continue
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
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  barang,
	}
}
