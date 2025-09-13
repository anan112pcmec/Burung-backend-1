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
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services/response_barang_user"
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

func LikesBarang(ctx context.Context, data PayloadLikesBarang, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	key := fmt.Sprintf("barang_disukai:%v>%v", data.IDUser, data.IDBarang)

	_, err_rds := rds.Get(ctx, key).Result()
	if err_rds == redis.Nil {
		if err := rds.Set(ctx, key, "disukai", 0).Err(); err != nil {
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: "Likes Barang",
				Payload:  "Gagal",
			}
		}

		if err := db.Create(&models.BarangDisukai{
			IdPengguna:    data.IDUser,
			IdBarangInduk: data.IDBarang,
		}).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: "Likes Barang",
				Payload:  "Gagal",
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: "Likes Barang",
			Payload: response_barang_user.ResponseLikesBarangUser{
				Message: "Disukai",
			},
		}
	}

	if err := rds.Del(ctx, key).Err(); err == nil {
		if err := db.Delete(&models.BarangDisukai{
			IdPengguna:    data.IDUser,
			IdBarangInduk: data.IDBarang,
		}).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: "Likes Barang",
				Payload:  "Gagal",
			}
		}
	} else {
		var existing models.BarangDisukai
		if err_db := db.Where("id_pengguna = ? AND id_barang_induk = ?", data.IDUser, data.IDBarang).First(&existing).Error; err_db == nil {
			if err := db.Delete(&existing).Error; err != nil {
				return &response.ResponseForm{
					Status:   http.StatusOK,
					Services: "Likes Barang",
					Payload:  "Gagal",
				}
			}
		} else if errors.Is(err_db, gorm.ErrRecordNotFound) {
			if err := db.Create(&models.BarangDisukai{
				IdPengguna:    data.IDUser,
				IdBarangInduk: data.IDBarang,
			}).Error; err != nil {
				return &response.ResponseForm{
					Status:   http.StatusOK,
					Services: "Likes Barang",
					Payload:  "Gagal",
				}
			}
			return &response.ResponseForm{
				Status:   http.StatusOK,
				Services: "Likes Barang",
				Payload: response_barang_user.ResponseLikesBarangUser{
					Message: "Disukai",
				},
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: "Likes Barang",
		Payload: response_barang_user.ResponseLikesBarangUser{
			Message: "Tidak Disukai",
		},
	}
}

func TambahKomentarBarang(ctx context.Context, data PayloadKomentarBarang, db *gorm.DB) *response.ResponseForm {
	services := "KomentarBarang"
	data.DataKomentar.JenisEntity = "User"
	if err_db := db.Model(models.Komentar{}).Create(&data.DataKomentar).Error; err_db != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_user.ResponseKomentarBarangUser{
				Message: "Gagal Unggah Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseKomentarBarangUser{
			Message: "Berhasil Unggah Komentar",
		},
	}
}

func EditKomentarBarang(ctx context.Context, data PayloadEditKomentarBarang, db *gorm.DB) *response.ResponseForm {
	services := "EditKomentarBarang"

	if err_db := db.Model(models.Komentar{}).Update("komentar", data.DataEditKomentar.Komentar).Error; err_db != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_user.ResponseEditKomentarBarangUser{
				Message: "Gagal Mengedit Komentar Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseEditKomentarBarangUser{
			Message: "Gagal Mengedit Komentar Coba Lagi Nanti",
		},
	}
}

func HapusKomentarBarang(ctx context.Context, data PayloadHapusKomentarBarang, db *gorm.DB, rds *redis.Client) {
}

func TambahKeranjangBarang(ctx context.Context, data PayloadTambahDataKeranjangBarang, db *gorm.DB) {}

func HapusKeranjangBarang(ctx context.Context) {}
