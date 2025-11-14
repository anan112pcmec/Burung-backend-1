package pengguna_service

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	transaksi_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/transaksi"
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

func LikesBarang(ctx context.Context, data PayloadLikesBarang, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "LikesBarang"
	var wg sync.WaitGroup

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
				Message: "Gagal Menghapus Alamat, Identitas Mu Tidak Sesuai.",
			},
		}
	}

	var id_pengguna_disukai int64 = 0
	if err := db.WithContext(ctx).Model(&models.BarangDisukai{}).Select("id_pengguna").Where(&models.BarangDisukai{
		IdPengguna:    data.IdentitasPengguna.ID,
		IdBarangInduk: data.IDBarangInduk,
	}).Limit(1).Scan(&id_pengguna_disukai).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_pengguna_disukai == 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err_incr_rds := rds.HIncrBy(
				ctx,
				fmt.Sprintf("barang:%d", data.IDBarangInduk),
				"likes_barang_induk",
				1,
			).Err(); err_incr_rds != nil {
				fmt.Println("Gagal increment likes di Redis:", err_incr_rds)
			}
		}()

		if err := db.WithContext(ctx).Create(&models.BarangDisukai{
			IdPengguna:    data.IdentitasPengguna.ID,
			IdBarangInduk: data.IDBarangInduk,
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

		wg.Wait()

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
				Message: "Disukai",
			},
		}
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err_decr_rds := rds.HIncrBy(
				ctx,
				fmt.Sprintf("barang:%d", data.IDBarangInduk),
				"likes_barang_induk",
				-1,
			).Err(); err_decr_rds != nil {
				fmt.Println("Gagal decrement likes di Redis:", err_decr_rds)
			}
		}()

		wg.Wait()
		if err := db.WithContext(ctx).Model(&models.BarangDisukai{}).
			Where(models.BarangDisukai{
				IdPengguna:    data.IdentitasPengguna.ID,
				IdBarangInduk: data.IDBarangInduk,
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

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseLikesBarangUser{
				Message: "Tidak Disukai",
			},
		}
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Engagement Barang Level Critical
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MasukanKomentarBarang(ctx context.Context, data PayloadMasukanKomentarBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "TambahKomentarBarang"

	if err := db.WithContext(ctx).Create(&models.Komentar{
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

	var id_komentar int64 = 0
	if err := db.WithContext(ctx).Model(&models.Komentar{}).Select("id").Where(&models.Komentar{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Limit(1).Scan(&id_komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKomentarBarangUser{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_komentar == 0 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKomentarBarangUser{
				Message: "Gagal Data Komentar Tidak Ditemukan",
			},
		}
	}

	if err := db.WithContext(ctx).Model(&models.Komentar{}).Where(&models.Komentar{
		ID: data.IdKomentar,
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

	var id_komentar int64 = 0
	if err := db.WithContext(ctx).Model(&models.Komentar{}).Select("id").Where(&models.Komentar{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Limit(1).Scan(&id_komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKomentarBarangUser{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_komentar == 0 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKomentarBarangUser{
				Message: "Gagal Data Komentar Tidak Ditemukan",
			},
		}
	}

	if err := db.WithContext(ctx).Model(&models.Komentar{}).Where(&models.Komentar{
		ID: data.IdKomentar,
	}).Delete(&models.Komentar{}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKomentarBarangUser{
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

	if err := db.WithContext(ctx).Create(&models.KomentarChild{
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

	if err := db.WithContext(ctx).Create(&models.KomentarChild{
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

	var id_edit_child_komentar int64 = 0
	if err := db.WithContext(ctx).Model(&models.KomentarChild{}).Select("id").Where(&models.KomentarChild{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Limit(1).Scan(&id_edit_child_komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditChildKomentar{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_edit_child_komentar == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditChildKomentar{
				Message: "Gagal Data Komentar Tidak Valid",
			},
		}
	}

	if err := db.WithContext(ctx).Model(&models.KomentarChild{}).Where(&models.KomentarChild{
		ID: data.IdKomentar,
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

	var id_edit_child_komentar int64 = 0
	if err := db.WithContext(ctx).Model(&models.KomentarChild{}).Select("id").Where(&models.KomentarChild{
		ID:          data.IdKomentar,
		IdEntity:    data.IdentitasPengguna.ID,
		JenisEntity: "Pengguna",
	}).Limit(1).Scan(&id_edit_child_komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusChildKomentar{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_edit_child_komentar == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusChildKomentar{
				Message: "Gagal Data Komentar Tidak Valid",
			},
		}
	}
	if err := db.Model(&models.KomentarChild{}).Where(&models.KomentarChild{
		ID: data.IdKomentar,
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

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Gagal: Data kamu tidak valid.",
			},
		}
	}

	var id_total []int64
	if err := db.WithContext(ctx).Model(&models.Keranjang{}).Select("id").Where(models.Keranjang{
		IdPengguna: data.IdentitasPengguna.ID,
	}).Limit(LIMITKERANJANG).Scan(&id_total).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Terjadi kesalahan saat cek jumlah keranjang.",
			},
		}
	}

	if len(id_total) >= LIMITKERANJANG {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: fmt.Sprintf("Maksimal barang dalam keranjang adalah %v.", LIMITKERANJANG),
			},
		}
	}

	var id_data_keranjang int64 = 0
	if err := db.WithContext(ctx).Model(&models.Keranjang{}).Select("id").Where(&models.Keranjang{
		IdPengguna:    data.IdentitasPengguna.ID,
		IdSeller:      data.IdSeller,
		IdBarangInduk: data.IdBarangInduk,
		IdKategori:    data.IdKategori,
	}).Limit(1).Scan(&id_data_keranjang).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_keranjang != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseTambahKeranjangUser{
				Message: "Gagal Kamu sudah memiliki itu di keranjang mu",
			},
		}
	}

	if err := db.WithContext(ctx).Create(&models.Keranjang{
		IdPengguna:    data.IdentitasPengguna.ID,
		IdSeller:      data.IdSeller,
		IdBarangInduk: data.IdBarangInduk,
		IdKategori:    data.IdKategori,
		Status:        "Ready",
		Jumlah:        0,
	}).Error; err != nil {
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

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Gagal: Data kamu tidak valid.",
			},
		}
	}

	var id_data_keranjang int64 = 0
	if err := db.WithContext(ctx).Model(&models.Keranjang{}).Select("id").Where(&models.Keranjang{
		ID:         data.IdKeranjang,
		IdPengguna: data.IdentitasPengguna.ID,
	}).Limit(1).Scan(&id_data_keranjang).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Gagal Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_keranjang == 0 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Gagal Data Keranjang Tidak Ditemukan",
			},
		}
	}

	var id_stok []int64
	if err := db.WithContext(ctx).Model(&models.VarianBarang{}).Select("id").Where(&models.VarianBarang{
		IdKategori:    data.IdKategori,
		IdBarangInduk: data.IdBarangInduk,
		Status:        "Ready",
	}).Limit(int(data.Jumlah)).Scan(&id_stok).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseEditKeranjangUser{
				Message: "Terjadi kesalahan saat cek stok.",
			},
		}
	}

	if int64(len(id_stok)) < data.Jumlah {
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
		Where(&models.Keranjang{
			ID: data.IdKeranjang,
		}).
		Update("jumlah", data.Jumlah).Error; err != nil {
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

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
				Message: "Gagal: Data kamu tidak valid.",
			},
		}
	}

	var id_data_keranjang int64 = 0
	if err := db.WithContext(ctx).Model(&models.Keranjang{}).Select("id").Where(&models.Keranjang{
		ID:         data.IdKeranjang,
		IdPengguna: data.IdentitasPengguna.ID,
	}).Limit(1).Scan(&id_data_keranjang).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
				Message: "Gagal Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_keranjang == 0 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseHapusKeranjangUser{
				Message: "Gagal Data Keranjang Tidak Ditemukan",
			},
		}
	}

	if err_hapus := db.WithContext(ctx).Model(&models.Keranjang{}).Where(&models.Keranjang{
		ID:         data.IdKeranjang,
		IdPengguna: data.IdentitasPengguna.ID,
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

func BerikanReviewBarang(ctx context.Context, data PayloadBerikanReviewBarang, db *gorm.DB) *response.ResponseForm {
	services := "BerikanReviewBarang"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseBerikanReviewBarang{
				Message: "Gagal data pengguna tidak valid",
			},
		}
	}

	var id_transaksi_data_selesai int64 = 0
	if err := db.WithContext(ctx).Model(&models.Transaksi{}).Select("id").Where(&models.Transaksi{
		IdBarangInduk: data.IdBarangInduk,
		IdPengguna:    data.IdentitasPengguna.ID,
		Status:        transaksi_enums.Selesai,
		Reviewed:      false,
	}).Limit(1).Scan(&id_transaksi_data_selesai).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseBerikanReviewBarang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_transaksi_data_selesai == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseBerikanReviewBarang{
				Message: "Gagal kamu tidak memiliki otoritas itu",
			},
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.Review{
			IdPengguna:    data.IdentitasPengguna.ID,
			IdBarangInduk: int32(data.IdBarangInduk),
			Rating:        data.Rating,
			Ulasan:        data.Ulasan,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Transaksi{}).Where(&models.Transaksi{
			ID: id_transaksi_data_selesai,
		}).Updates(&models.Transaksi{
			Reviewed: true,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_engagement_barang_pengguna.ResponseBerikanReviewBarang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_engagement_barang_pengguna.ResponseBerikanReviewBarang{
			Message: "Berhasil",
		},
	}
}

func LikeReviewBarang(ctx context.Context, data PayloadLikeReviewBarang, db *gorm.DB) *response.ResponseForm {
	services := "LikeReviewBarang"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal data pengguna tidak valid",
		}
	}

	var id_review_like int64 = 0
	if err := db.Model(&models.ReviewLike{}).
		Select("id").
		Where(&models.ReviewLike{
			IdPengguna: data.IdentitasPengguna.ID,
			IdReview:   data.IdReview,
		}).
		Limit(1).
		Scan(&id_review_like).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_review_like != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal kamu sudah like review itu",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Insert like
		if err := tx.Create(&models.ReviewLike{
			IdPengguna: data.IdentitasPengguna.ID,
			IdReview:   data.IdReview,
		}).Error; err != nil {
			return err
		}

		// Update kolom "like"
		if err := tx.Model(&models.Review{}).
			Where("id = ?", data.IdReview).
			UpdateColumn(`"like"`, gorm.Expr(`"like" + 1`)).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil",
	}
}

func UnlikeReviewBarang(ctx context.Context, data PayloadUnlikeReviewBarang, db *gorm.DB) *response.ResponseForm {
	services := "UnlikeReviewBarang"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal data pengguna tidak valid",
		}
	}

	var id_review_like int64 = 0
	if err := db.WithContext(ctx).Model(&models.ReviewLike{}).
		Select("id").
		Where(&models.ReviewLike{
			IdPengguna: data.IdentitasPengguna.ID,
			IdReview:   data.IdReview,
		}).
		Limit(1).
		Scan(&id_review_like).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_review_like == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data like tidak ditemukan",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Hapus like
		if err := tx.Delete(&models.ReviewLike{}, id_review_like).Error; err != nil {
			return err
		}

		// Decrement kolom "like", pastikan tidak negatif
		if err := tx.Model(&models.Review{}).
			Where("id = ?", data.IdReview).
			UpdateColumn(`"like"`, gorm.Expr(`CASE WHEN "like" > 0 THEN "like" - 1 ELSE 0 END`)).
			Error; err != nil {
			return err
		}

		return nil
	}); err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil",
	}
}
