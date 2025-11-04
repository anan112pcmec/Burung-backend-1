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
	response_engagement_barang_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services/response_barang"
)

var fieldBarangViewed = "viewed_barang_induk"

const LIMITKERANJANG = 30

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Engagement Barang Level Uncritical
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur View Barang
// Berfungsi Untuk Menambah View Barang Setiap kali di klik akan menjalankan fungsi ini
// Hanya bersifat menaikan view (increment)
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ViewBarang(data PayloadViewBarang, rds *redis.Client, db *gorm.DB) {
	ctx := context.Background()
	key := fmt.Sprintf("barang:%d", data.ID)

	// Jika gagal increment di Redis -> fallback update ke DB (asynchronous)
	if err := rds.HIncrBy(ctx, key, fieldBarangViewed, 1).Err(); err != nil {
		go func() {
			_ = db.Model(&models.BarangInduk{}).
				Where("id = ?", data.ID).
				UpdateColumn("viewed", gorm.Expr("viewed + 1")).Error
		}()
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Likes Barang
// :Berfungsi Untuk Menambah Dan Mengurangi Likes Barang induk dan mencatat barangdisukai
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func LikesBarang(data PayloadLikesBarang, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "LikesBarang"

	var existing models.BarangDisukai
	err := db.Unscoped().Model(&models.BarangDisukai{}).
		Where(&models.BarangDisukai{IdPengguna: data.IDUser, IdBarangInduk: data.IDBarang}).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&models.BarangDisukai{
			IdPengguna:    data.IDUser,
			IdBarangInduk: data.IDBarang,
		}).Error; err != nil {
			fmt.Println("Gagal likes:", err)
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
					Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
				},
			}
		}

		go func() {
			ctx := context.Background()
			if err_incr_rds := rds.HIncrBy(
				ctx,
				fmt.Sprintf("barang:%d", data.IDBarang),
				"likes_barang_induk",
				1,
			).Err(); err_incr_rds != nil {
				fmt.Println("Gagal increment likes di Redis:", err_incr_rds)
			}
		}()

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
				Message: "Disukai",
			},
		}
	}

	if err == nil {
		if err := db.Unscoped().Model(&models.BarangDisukai{}).
			Where(models.BarangDisukai{
				IdPengguna:    existing.IdPengguna,
				IdBarangInduk: existing.IdBarangInduk,
			}).Delete(&models.BarangDisukai{}).Error; err != nil {
			fmt.Println("Gagal Hapus Likes:", err)
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
					Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
				},
			}
		}

		go func() {
			if err_decr_rds := rds.HIncrBy(
				context.Background(),
				fmt.Sprintf("barang:%d", data.IDBarang),
				"likes_barang_induk",
				-1,
			).Err(); err_decr_rds != nil {
				fmt.Println("Gagal decrement likes di Redis:", err_decr_rds)
			}
		}()

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
				Message: "Tidak Disukai",
			},
		}
	}

	fmt.Println("Error cek existing likes:", err)
	return &response.ResponseForm{
		Status:   http.StatusInternalServerError,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
			Message: "Terjadi kesalahan.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Engagement Barang Level Critical
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MasukanKomentarBarang(ctx context.Context, data PayloadMasukanKomentarBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "TambahKomentarBarang"

	if err := db.Create(&models.Komentar{
		IdBarangInduk: data.IdBarangInduk,
		IdEntity:      data.IdentitasPengguna.ID,
		JenisEntity:   "Pengguna",
		Komentar:      data.Komentar,
		IsSeller:      false,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseMasukanKomentarBarangUser{
				Message: "Gagal Memposting Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseMasukanKomentarBarangUser{
			Message: "Berhasil",
		},
	}
}

func EditKomentarBarang(ctx context.Context, data PayloadEditKomentarBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "EditKomentarBarang"

	if err := db.Model(&models.Komentar{}).Where(&models.Komentar{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Update("komentar", data.Komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKomentarBarangUser{
				Message: "Gagal Mengedit Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseEditKomentarBarangUser{
			Message: "Berhasil",
		},
	}
}

func HapusKomentarBarang(ctx context.Context, data PayloadHapusKomentarBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "HapusKomentarBarang"

	if err := db.Model(&models.Komentar{}).Where(&models.Komentar{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Delete(&models.Komentar{}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
				Message: "Gagal Menghapus Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseHapusKomentarBarangUser{
			Message: "Berhasil",
		},
	}
}

func MasukanChildKomentar(ctx context.Context, data PayloadMasukanChildKomentar, db *gorm.DB) *response.ResponseForm {
	services := "MasukanChildKomentar"

	if err := db.Create(&models.KomentarChild{
		IdKomentar:  data.IdKomentarBarang,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
		IsiKomentar: data.Komentar,
		IsSeller:    false,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseMasukanChildKomentar{
				Message: "Gagal Mengunggah Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseMasukanChildKomentar{
			Message: "Berhasil",
		},
	}
}

func MentionChildKomentar(ctx context.Context, data PayloadMentionChildKomentar, db *gorm.DB) *response.ResponseForm {
	services := "MentionChildKomentar"

	if err := db.Create(&models.KomentarChild{
		IdKomentar:  data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
		IsiKomentar: data.Komentar,
		IsSeller:    false,
		Mention:     data.UsernameMentioned,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseMentionChildKomentar{
				Message: "Gagal Membalas Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseMentionChildKomentar{
			Message: "Berhasil",
		},
	}
}

func EditChildKomentar(ctx context.Context, data PayloadEditChildKomentar, db *gorm.DB) *response.ResponseForm {
	services := "EditChildKomentar"

	if err := db.Model(&models.KomentarChild{}).Where(&models.KomentarChild{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Update("komentar", data.Komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditChildKomentar{
				Message: "Gagal Mengedit Komentar",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseEditChildKomentar{
			Message: "Berhasil",
		},
	}
}

func HapusChildKomentar(ctx context.Context, data PayloadHapusChildKomentar, db *gorm.DB) *response.ResponseForm {
	services := "HapusChildKomentar"
	if err := db.Model(&models.KomentarChild{}).Where(&models.KomentarChild{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Delete(&models.KomentarChild{}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusChildKomentar{
				Message: "Gagal Menghapus Komentar",
			},
		}
	}
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseHapusChildKomentar{
			Message: "Berhasil",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Tambah Keranjang Barang
// :Berfungsi Untuk menambahkan sebuah barang ke keranjang pengguna tertentu
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TambahKeranjangBarang(ctx context.Context, data PayloadTambahDataKeranjangBarang, db *gorm.DB) *response.ResponseForm {
	services := "TambahKeranjangBarang"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Gagal: Data kamu tidak valid.",
			},
		}
	}

	var total int64
	if err := db.WithContext(ctx).
		Model(&models.Keranjang{}).
		Where(models.Keranjang{IdPengguna: data.DataTambahKeranjang.IdPengguna}).
		Limit(LIMITKERANJANG).
		Count(&total).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Terjadi kesalahan saat cek jumlah keranjang.",
			},
		}
	}

	if total >= int64(LIMITKERANJANG) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: fmt.Sprintf("Maksimal barang dalam keranjang adalah %v.", LIMITKERANJANG),
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
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Kamu sudah memiliki barang ini di keranjang.",
			},
		}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Terjadi kesalahan saat cek keranjang.",
			},
		}
	}

	// Hardcoded
	data.DataTambahKeranjang.Count = 0
	data.DataTambahKeranjang.Status = "Ready"
	//

	if err := db.WithContext(ctx).Create(&data.DataTambahKeranjang).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Gagal menambahkan barang ke keranjang.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusCreated,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
			Message: "Berhasil menambahkan barang ke keranjang.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Edit Keranjang Barang
// :Berfungsi Untuk mengedit sebuah count dari keranjang pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EditKeranjangBarang(ctx context.Context, data PayloadEditDataKeranjangBarang, db *gorm.DB) *response.ResponseForm {
	services := "EditKeranjangBarang"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Gagal: Data kamu tidak valid.",
			},
		}
	}

	var keranjang models.Keranjang
	if err := db.WithContext(ctx).
		Where(&models.Keranjang{IdPengguna: data.IdentitasPengguna.ID, IdBarangInduk: data.IdBarangInduk}).
		First(&keranjang).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
					Message: "Barang tersebut tidak ada di keranjang.",
				},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Terjadi kesalahan saat cek keranjang.",
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
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Terjadi kesalahan saat cek stok.",
			},
		}
	}

	if stok < int64(data.Jumlah) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Jumlah melebihi stok tersedia.",
			},
		}
	}

	if err := db.WithContext(ctx).
		Model(&models.Keranjang{}).
		Where(&models.Keranjang{IdPengguna: data.IdentitasPengguna.ID, IdBarangInduk: data.IdBarangInduk}).
		Update("count", data.Jumlah).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Gagal memperbarui jumlah barang.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
			Message: "Jumlah barang berhasil diperbarui.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Hapus Keranjang Barang
// :Berfungsi Untuk menghapus suatu barang dari keranjang pengguna tertentu
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusKeranjangBarang(ctx context.Context, data PayloadHapusDataKeranjangBarang, db *gorm.DB) *response.ResponseForm {
	services := "HapusKeranjangBarang"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
				Message: "Gagal: Data kamu tidak valid.",
			},
		}
	}

	if err_hapus := db.Unscoped().Model(&models.Keranjang{}).Where(&models.Keranjang{
		IdPengguna:    data.DataHapusKeranjang.IdPengguna,
		IdBarangInduk: data.DataHapusKeranjang.IdBarangInduk,
		IdKategori:    data.DataHapusKeranjang.IdKategori,
	}).Delete(&models.Keranjang{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
				Message: "Gagal menghapus barang keranjang, coba lagi nanti.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
			Message: "Barang berhasil dihapus dari keranjang.",
		},
	}
}
