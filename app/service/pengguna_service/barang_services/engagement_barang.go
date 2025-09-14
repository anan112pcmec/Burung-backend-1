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

func LikesBarang(ctx context.Context, data PayloadLikesBarang, db *gorm.DB) *response.ResponseForm {
	services := "Likes Barang"

	var existing models.BarangDisukai
	err := db.Where("id_pengguna = ? AND id_barang_induk = ?", data.IDUser, data.IDBarang).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&models.BarangDisukai{
			IdPengguna:    data.IDUser,
			IdBarangInduk: data.IDBarang,
		}).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  "Gagal menambah like",
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_barang_user.ResponseLikesBarangUser{
				Message: "Disukai",
			},
		}
	}

	if err == nil {
		// sudah ada, maka hapus
		if err := db.Delete(&existing).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  "Gagal menghapus like",
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_barang_user.ResponseLikesBarangUser{
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

	result := db.WithContext(ctx).
		Model(&models.Komentar{}).
		Where(models.Komentar{ID: data.DataEditKomentar.ID, IdBarangInduk: data.DataEditKomentar.IdBarangInduk, IdEntity: data.DataEditKomentar.IdEntity}).Statement.Not(models.Komentar{Komentar: data.DataEditKomentar.Komentar}).
		Update("komentar", &data.DataEditKomentar.Komentar)

	if result.Error != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_user.ResponseEditKomentarBarangUser{
				Message: "Gagal mengedit komentar, coba lagi nanti",
			},
		}
	}

	if result.RowsAffected == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_user.ResponseEditKomentarBarangUser{
				Message: "Komentar tidak ditemukan",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseEditKomentarBarangUser{
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
			Payload: response_barang_user.ResponseHapusKomentarBarangUser{
				Message: "Gagal hapus komentar",
			},
		}
	}

	if result.RowsAffected == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_user.ResponseHapusKomentarBarangUser{
				Message: "Komentar tidak ditemukan",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseHapusKomentarBarangUser{
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
			Payload: response_barang_user.ResponseTambahKeranjangUser{
				Message: "Terjadi kesalahan saat cek jumlah keranjang",
			},
		}
	}

	if total >= 30 {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang_user.ResponseTambahKeranjangUser{
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
			Payload: response_barang_user.ResponseTambahKeranjangUser{
				Message: "Kamu sudah memiliki barang ini di keranjang",
			},
		}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_user.ResponseTambahKeranjangUser{
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
			Payload: response_barang_user.ResponseTambahKeranjangUser{
				Message: "Gagal menambahkan barang ke keranjang",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusCreated,
		Services: services,
		Payload: response_barang_user.ResponseTambahKeranjangUser{
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
				Payload: response_barang_user.ResponseEditKeranjangUser{
					Message: "Barang tersebut tidak ada di keranjang",
				},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_user.ResponseEditKeranjangUser{
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
			Payload: response_barang_user.ResponseEditKeranjangUser{
				Message: "Terjadi kesalahan saat cek stok",
			},
		}
	}

	if stok < int64(data.Jumlah) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang_user.ResponseEditKeranjangUser{
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
			Payload: response_barang_user.ResponseEditKeranjangUser{
				Message: "Gagal memperbarui jumlah barang",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseEditKeranjangUser{
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
			Payload: response_barang_user.ResponseHapusKeranjangUser{
				Message: "Gagal Menghapus Barang Keranjang, Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseHapusKeranjangUser{
			Message: "Barang berhasil dihapus dari keranjang",
		},
	}
}
