package pengguna_service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/barang_service/response_barang"

)

var fieldBarangViewed = "viewed_barang_induk"

func ViewBarang(data PayloadWatchBarang, rds *redis.Client, db *gorm.DB) {
	ctx := context.Background()
	key := fmt.Sprintf("barang:%d", data.ID)

	// coba increment di Redis
	if err := rds.HIncrBy(ctx, key, fieldBarangViewed, 1).Err(); err != nil {
		// kalau Redis error, fallback ke DB
		go func() {
			_ = db.Model(&models.BarangInduk{}).
				Where("id = ?", data.ID).
				UpdateColumn("viewed", gorm.Expr("viewed + 1")).Error
		}()
	}
}

func LikesBarang(data PayloadLikesBarang, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "Likes Barang"

	var existing models.BarangDisukai
	err := db.Where("id_pengguna = ? AND id_barang_induk = ?", data.IDUser, data.IDBarang).First(&existing).Error

	// ✅ case: belum pernah like
	if errors.Is(err, gorm.ErrRecordNotFound) {
		go func() {
			ctx := context.Background()

			// insert barang disukai
			if err := db.Create(&models.BarangDisukai{
				IdPengguna:    data.IDUser,
				IdBarangInduk: data.IDBarang,
			}).Error; err != nil {
				fmt.Println("Gagal likes:", err)
			}

			var likes int64
			if err_ambil := db.Model(models.BarangInduk{}).
				Select("likes").
				Where(models.BarangInduk{ID: data.IDBarang}).
				Take(&likes).Error; err_ambil != nil {
				return
			}

			// increment likes di DB
			if err_incr_db := db.Model(&models.BarangInduk{}).
				Where("id = ?", data.IDBarang).
				Update("likes", likes+1).Error; err_incr_db != nil {
				fmt.Println("Gagal increment likes di DB:", err_incr_db)
				return
			}

			// increment likes di Redis
			if err_incr_rds := rds.HIncrBy(
				ctx,
				fmt.Sprintf("barang:%v", data.IDBarang),
				"likes_barang_induk",
				likes+1,
			).Err(); err_incr_rds != nil {
				fmt.Println("Gagal increment likes di Redis:", err_incr_rds)
			}
		}()

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_barang.ResponseLikesBarangUser{
				Message: "Disukai",
			},
		}
	}

	// ✅ case: sudah pernah like
	if err == nil {
		go func() {
			if err := db.Model(models.BarangDisukai{}).
				Delete(&existing, models.BarangDisukai{
					IdPengguna:    existing.IdPengguna,
					IdBarangInduk: existing.IdBarangInduk,
				}).Error; err != nil {
				fmt.Println("Gagal Hapus Likes")
				return
			}

			var likes int64
			if err_ambil := db.Model(models.BarangInduk{}).
				Select("likes").
				Where(models.BarangInduk{ID: data.IDBarang}).
				Take(&likes).Error; err_ambil != nil {
				return
			}

			// decrement likes di DB (⚠️ sebelumnya salah +1, sekarang -1)
			if err_decr_db := db.Model(&models.BarangInduk{}).
				Where("id = ?", data.IDBarang).
				Update("likes", likes-1).Error; err_decr_db != nil {
				fmt.Println("Gagal decrement likes di DB:", err_decr_db)
			}

			// decrement likes di Redis
			if err_decr_rds := rds.HIncrBy(
				context.Background(),
				fmt.Sprintf("barang:%v", data.IDBarang),
				"likes_barang_induk",
				-1,
			).Err(); err_decr_rds != nil {
				fmt.Println("Gagal decrement likes di Redis:", err_decr_rds)
			}
		}()

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_barang.ResponseLikesBarangUser{
				Message: "Tidak Disukai",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusInternalServerError,
		Services: services,
		Payload:  "Terjadi kesalahan",
	}
}

func TambahKomentarBarang(ctx context.Context, data PayloadKomentarBarang, db *gorm.DB) *response.ResponseForm {
	services := "KomentarBarang"
	data.DataKomentar.JenisEntity = "User"
	if err_db := db.WithContext(ctx).Model(models.Komentar{}).Create(&data.DataKomentar).Error; err_db != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseKomentarBarangUser{
				Message: "Gagal Unggah Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang.ResponseKomentarBarangUser{
			Message: "Berhasil Unggah Komentar",
		},
	}
}

func EditKomentarBarang(ctx context.Context, data PayloadEditKomentarBarang, db *gorm.DB) *response.ResponseForm {
	services := "EditKomentarBarang"

	result := db.WithContext(ctx).
		Model(&models.Komentar{}).
		Where(models.Komentar{ID: data.DataEditKomentar.ID, IdBarangInduk: data.DataEditKomentar.IdBarangInduk, IdEntity: data.DataEditKomentar.IdEntity}).Statement.Not(models.Komentar{Komentar: data.DataEditKomentar.Komentar}).
		Update("komentar", &data.DataEditKomentar.Komentar)

	if result.Error != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseEditKomentarBarangUser{
				Message: "Gagal mengedit komentar, coba lagi nanti",
			},
		}
	}

	if result.RowsAffected == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang.ResponseEditKomentarBarangUser{
				Message: "Komentar tidak ditemukan",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang.ResponseEditKomentarBarangUser{
			Message: "Berhasil mengedit komentar",
		},
	}
}

func HapusKomentarBarang(ctx context.Context, data PayloadHapusKomentarBarang, db *gorm.DB) *response.ResponseForm {
	services := "HapusKomentarBarang"

	result := db.WithContext(ctx).
		Where(&models.Komentar{
			ID:            data.IDKomentar,
			IdBarangInduk: data.IdBarang,
			IdEntity:      data.IDEntity,
		}).
		Delete(&models.Komentar{})

	if result.Error != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseHapusKomentarBarangUser{
				Message: "Gagal hapus komentar",
			},
		}
	}

	if result.RowsAffected == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang.ResponseHapusKomentarBarangUser{
				Message: "Komentar tidak ditemukan",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang.ResponseHapusKomentarBarangUser{
			Message: "Berhasil hapus komentar",
		},
	}
}

func TambahKeranjangBarang(ctx context.Context, data PayloadTambahDataKeranjangBarang, db *gorm.DB) *response.ResponseForm {
	services := "TambahKeranjangBarang"

	var total int64
	if err := db.WithContext(ctx).
		Model(&models.Keranjang{}).
		Where(models.Keranjang{IdPengguna: data.DataTambahKeranjang.IdPengguna}).
		Limit(31).
		Count(&total).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseTambahKeranjangUser{
				Message: "Terjadi kesalahan saat cek jumlah keranjang",
			},
		}
	}

	if total >= 30 {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang.ResponseTambahKeranjangUser{
				Message: "Maksimal barang dalam keranjang adalah 30",
			},
		}
	}

	err := db.WithContext(ctx).
		Where(&models.Keranjang{
			IdPengguna:    data.DataTambahKeranjang.IdPengguna,
			IdBarangInduk: data.DataTambahKeranjang.IdBarangInduk,
			IdKategori:    data.DataTambahKeranjang.IdKategori,
		}).
		Take(&models.Keranjang{}).Error

	if err == nil {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_barang.ResponseTambahKeranjangUser{
				Message: "Kamu sudah memiliki barang ini di keranjang",
			},
		}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseTambahKeranjangUser{
				Message: "Terjadi kesalahan saat cek keranjang",
			},
		}
	}

	data.DataTambahKeranjang.Count = 0
	data.DataTambahKeranjang.Status = "Ready"

	if err := db.WithContext(ctx).Create(&data.DataTambahKeranjang).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseTambahKeranjangUser{
				Message: "Gagal menambahkan barang ke keranjang",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusCreated,
		Services: services,
		Payload: response_barang.ResponseTambahKeranjangUser{
			Message: "Berhasil menambahkan barang ke keranjang",
		},
	}
}

func EditKeranjangBarang(ctx context.Context, data PayloadEditDataKeranjangBarang, db *gorm.DB) *response.ResponseForm {
	services := "EditKeranjangBarang"

	var keranjang models.Keranjang
	if err := db.WithContext(ctx).
		Where(&models.Keranjang{IdPengguna: data.IdPengguna, IdBarangInduk: data.IdBarangInduk}).
		First(&keranjang).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload: response_barang.ResponseEditKeranjangUser{
					Message: "Barang tersebut tidak ada di keranjang",
				},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseEditKeranjangUser{
				Message: "Terjadi kesalahan saat cek keranjang",
			},
		}
	}

	var stok int64
	if err := db.WithContext(ctx).
		Model(&models.VarianBarang{}).
		Where(&models.VarianBarang{
			IdKategori:    data.IdKategori,
			IdBarangInduk: data.IdBarangInduk,
			Status:        "Ready",
		}).
		Count(&stok).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseEditKeranjangUser{
				Message: "Terjadi kesalahan saat cek stok",
			},
		}
	}

	if stok < int64(data.Jumlah) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang.ResponseEditKeranjangUser{
				Message: "Jumlah melebihi stok tersedia",
			},
		}
	}

	if err := db.WithContext(ctx).
		Model(&models.Keranjang{}).
		Where(&models.Keranjang{IdPengguna: data.IdPengguna, IdBarangInduk: data.IdBarangInduk}).
		Update("count", data.Jumlah).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseEditKeranjangUser{
				Message: "Gagal memperbarui jumlah barang",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang.ResponseEditKeranjangUser{
			Message: "Jumlah barang berhasil diperbarui",
		},
	}
}

func HapusKeranjangBarang(ctx context.Context, data PayloadHapusDataKeranjangBarang, db *gorm.DB) *response.ResponseForm {
	services := "HapusKeranjangBarang"

	if err_hapus := db.Where(models.Keranjang{
		IdPengguna:    data.DataHapusKeranjang.IdPengguna,
		IdBarangInduk: data.DataHapusKeranjang.IdBarangInduk,
		IdKategori:    data.DataHapusKeranjang.IdKategori,
	}).Delete(&models.Keranjang{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang.ResponseHapusKeranjangUser{
				Message: "Gagal Menghapus Barang Keranjang, Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang.ResponseHapusKeranjangUser{
			Message: "Barang berhasil dihapus dari keranjang",
		},
	}
}
