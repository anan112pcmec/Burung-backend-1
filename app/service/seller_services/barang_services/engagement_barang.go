package seller_service

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	barang_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/barang"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services/response_barang_service"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Masukan Barang Seller
// Berfungsi untuk melayani seller yang hendak memasukan barang nya ke sistem burung
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MasukanBarangInduk(db *gorm.DB, data PayloadMasukanBarangInduk) *response.ResponseForm {
	services := "MasukanBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseMasukanBarangInduk{
				Message: "Gagal memasukkan barang karena kredensial seller tidak valid",
			},
		}
	}

	data.BarangInduk.SellerID = data.IdentitasSeller.IdSeller

	if validasi := data.BarangInduk.Validating(); validasi != "Data Lengkap" {
		log.Printf("[WARN] Data barang tidak lengkap untuk seller ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal: Data barang harus dilengkapi",
		}
	}

	var namaBarang string
	errNama := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			SellerID:   data.IdentitasSeller.IdSeller,
			NamaBarang: data.BarangInduk.NamaBarang,
		}).
		Select("nama_barang").
		Take(&namaBarang).Error

	if errNama == nil {
		log.Printf("[WARN] Seller ID %d sudah memiliki barang dengan nama '%s'", data.IdentitasSeller.IdSeller, data.BarangInduk.NamaBarang)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload:  "Gagal: Nama barang sudah terdaftar untuk seller ini",
		}
	}

	localData := data

	go func(d PayloadMasukanBarangInduk) {
		if err := db.Transaction(func(tx *gorm.DB) error {

			if d.BarangInduk.OriginalKategori == "" {
				return fmt.Errorf("OriginalKategori kosong, rollback transaksi")
			}

			d.BarangInduk.SellerID = d.IdentitasSeller.IdSeller

			if err := tx.Create(&d.BarangInduk).Error; err != nil {
				return err
			}

			idInduk := d.BarangInduk.ID

			for _, kategori := range d.KategoriBarang {
				kategori.IdBarangInduk = int32(idInduk)

				// Validasi Rekening
				var jumlah_rek int64 = 0
				if err := tx.Model(&models.RekeningSeller{}).Where(&models.RekeningSeller{
					ID:       kategori.IDRekening,
					IDSeller: d.IdentitasSeller.IdSeller,
				}).Count(&jumlah_rek).Error; err != nil {
					continue
				}

				if jumlah_rek != 1 {
					continue
				}

				// Validasi Alamat Gudang
				var jumlah_alamat int64 = 0
				if err := tx.Model(&models.AlamatGudang{}).Where(&models.AlamatGudang{
					ID:       kategori.IDAlamat,
					IDSeller: d.IdentitasSeller.IdSeller,
				}).Count(&jumlah_alamat).Error; err != nil {
					continue
				}

				if jumlah_alamat != 1 {
					continue
				}

				if err := tx.Create(&kategori).Error; err != nil {
					return err
				}

				idKategori := kategori.ID

				var varianBatch []models.VarianBarang
				for i := 0; i < int(kategori.Stok); i++ {
					if kategori.IDRekening != 0 && kategori.IDAlamat != 0 {
						varianBatch = append(varianBatch, models.VarianBarang{
							IdBarangInduk: int32(idInduk),
							IdKategori:    idKategori,
							Status:        barang_enums.Ready,
							Sku:           kategori.Sku,
						})
					} else {
						varianBatch = append(varianBatch, models.VarianBarang{
							IdBarangInduk: int32(idInduk),
							IdKategori:    idKategori,
							Status:        barang_enums.Pending,
							Sku:           kategori.Sku,
						})
					}
				}

				if len(varianBatch) > 0 {
					if err := tx.Create(&varianBatch).Error; err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] MasukanBarang gagal insert data untuk seller ID %d: %v", d.IdentitasSeller.IdSeller, err)
			return
		}

		log.Printf("[INFO] Barang berhasil dimasukkan untuk seller ID %d", d.IdentitasSeller.IdSeller)
	}(localData)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseMasukanBarangInduk{
			Message: "Barang berhasil ditambahkan.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Edit Barang Induk
// Berfungsi untuk seller dalam melakukan edit atau pembaruan informasi seputar barang induknya
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EditBarangInduk(db *gorm.DB, data PayloadEditBarangInduk) *response.ResponseForm {
	services := "EditBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditBarangInduk{
				Message: "Gagal: Kredensial Seller Tidak Valid",
			},
		}
	}

	data.BarangInduk.SellerID = data.IdentitasSeller.IdSeller

	var count int64 = 0
	err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID: data.BarangInduk.ID,
		}).
		Count(&count).Error

	if err != nil {
		log.Printf("[ERROR] Terjadi kesalahan server saat validasi barang: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditBarangInduk{
				Message: "Terjadi Kesalahan Server Saat Validasi Barang",
			},
		}
	}

	if count != 1 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditBarangInduk{
				Message: "Gagal: Barang Tidak Ada atau Tidak Ditemukan",
			},
		}
	}

	go func() {
		data.BarangInduk.TanggalRilis = ""
		data.BarangInduk.Viewed = 0
		data.BarangInduk.Likes = 0
		data.BarangInduk.TotalKomentar = 0
		data.BarangInduk.HargaKategoris = 0

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.BarangInduk{}).
				Where(&models.BarangInduk{
					ID:       data.BarangInduk.ID,
					SellerID: data.IdentitasSeller.IdSeller,
				}).
				Updates(&data.BarangInduk).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] EditBarang gagal update BarangInduk ID %d: %v", data.BarangInduk.ID, err)
			return
		}

		log.Printf("[INFO] Barang berhasil diubah untuk seller ID %d", data.IdentitasSeller.IdSeller)
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditBarangInduk{
			Message: "Barang Berhasil Diubah.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Hapus Barang Induk
// Berfungsi untuk seller dalam menghapus barang induknya, akan otomatis menghapus kategori barang dan varian barang
// nya
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusBarangInduk(db *gorm.DB, data PayloadHapusBarangInduk) *response.ResponseForm {
	services := "HapusBarang"

	// Validasi kredensial seller
	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseHapusBarangInduk{
				Message: "Gagal: Kredensial seller tidak valid",
			},
		}
	}

	// Cek apakah masih ada varian dalam transaksi (status: Dipesan/Diproses)
	var jumlahDalamTransaksi int64
	if err := db.Model(&models.VarianBarang{}).
		Where("id_barang_induk = ? AND status IN ?", data.BarangInduk.ID, []string{"Dipesan", "Diproses"}).
		Count(&jumlahDalamTransaksi).Error; err != nil {
		log.Printf("[ERROR] Gagal memeriksa stok dalam transaksi: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Terjadi kesalahan pada database",
		}
	}

	// Jika masih ada barang dalam transaksi, hentikan proses
	if jumlahDalamTransaksi != 0 {
		log.Printf("[WARN] Masih ada %v barang dalam transaksi untuk Barang ID %d", jumlahDalamTransaksi, data.BarangInduk.ID)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_barang_service.ResponseHapusBarangInduk{
				Message: fmt.Sprintf("Masih ada %v barang dalam transaksi", jumlahDalamTransaksi),
			},
		}
	}

	// Jalankan proses penghapusan dalam goroutine (asynchronous)
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {

			// Hapus permanen semua varian barang terkait
			if err := tx.Unscoped().Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{IdBarangInduk: data.BarangInduk.ID}).
				Delete(&models.VarianBarang{}).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.KategoriBarang{}).
				Where(&models.KategoriBarang{
					IdBarangInduk: data.BarangInduk.ID,
				}).
				Delete(&models.KategoriBarang{}).Error; err != nil {
				return fmt.Errorf("gagal update deleted_at kategori: %w", err)
			}

			// Update kolom deleted_at barang induk
			if err := tx.Model(&models.BarangInduk{}).
				Where(&models.BarangInduk{
					ID: data.BarangInduk.ID,
				}).
				Delete(&models.BarangInduk{}).Error; err != nil {
				return fmt.Errorf("gagal update deleted_at barang induk: %w", err)
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] Rollback: Gagal menghapus barang '%s' (ID %d): %v",
				data.BarangInduk.NamaBarang, data.BarangInduk.ID, err)
			return
		}

		log.Printf("[INFO] Barang berhasil dihapus (soft delete) untuk Seller ID %d (Barang ID %d)",
			data.IdentitasSeller.IdSeller, data.BarangInduk.ID)
	}()

	// Kembalikan respons sukses
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusBarangInduk{
			Message: "Barang berhasil dihapus.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Tambah Kategori Barang Induk
// Berfungsi untuk seller menambahkan kategori barang pada barang induk
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TambahKategoriBarang(db *gorm.DB, data PayloadTambahKategori) *response.ResponseForm {
	services := "TambahKategoriBarang"

	// Validasi kredensial seller
	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseTambahKategori{
				Message: "Gagal: kredensial seller tidak valid",
			},
		}
	}

	// Pastikan barang induk milik seller
	var idBarangIndukGet int64
	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan. IdBarangInduk=%d, IdSeller=%d",
			data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseTambahKategori{
				Message: "Gagal: Barang induk tidak ditemukan",
			},
		}
	}

	// 👉 Buat salinan data untuk mencegah race di goroutine
	dataCopy := PayloadTambahKategori{
		IdentitasSeller: data.IdentitasSeller,
		IdBarangInduk:   data.IdBarangInduk,
		KategoriBarang:  append([]models.KategoriBarang(nil), data.KategoriBarang...), // copy slice
	}

	// Jalankan async tapi dengan salinan data yang aman
	go func(d PayloadTambahKategori) {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range d.KategoriBarang {
				d.KategoriBarang[i].IdBarangInduk = d.IdBarangInduk

				// Gunakan FOR UPDATE untuk lock kategori yang sedang diperiksa (mencegah race)
				var existingKategori models.KategoriBarang
				err := tx.Model(&models.KategoriBarang{}).
					Where(&models.KategoriBarang{
						IdBarangInduk: d.IdBarangInduk,
						Nama:          d.KategoriBarang[i].Nama,
					}).
					First(&existingKategori).Error

				if err == nil {
					log.Printf("[WARN] Kategori '%s' sudah ada pada barang induk ID %d",
						d.KategoriBarang[i].Nama, d.IdBarangInduk)
					continue
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("[ERROR] Gagal cek kategori '%s': %v", d.KategoriBarang[i].Nama, err)
					continue
				}

				var jumlah_rek int64 = 0
				if err := tx.Model(&models.RekeningSeller{}).Where(&models.RekeningSeller{
					ID:       d.KategoriBarang[i].IDRekening,
					IDSeller: d.IdentitasSeller.IdSeller,
				}).Count(&jumlah_rek).Error; err != nil {
					continue
				}

				if jumlah_rek != 1 {
					continue
				}

				var jumlah_alamat int64 = 0
				if err := tx.Model(&models.AlamatGudang{}).Where(&models.AlamatGudang{
					ID:       d.KategoriBarang[i].IDAlamat,
					IDSeller: d.IdentitasSeller.IdSeller,
				}).Count(&jumlah_alamat).Error; err != nil {
					continue
				}

				if jumlah_alamat != 1 {
					continue
				}

				if err := tx.Create(&d.KategoriBarang[i]).Error; err != nil {
					log.Printf("[ERROR] Gagal menambah kategori '%s' untuk barang induk %d: %v",
						d.KategoriBarang[i].Nama, d.IdBarangInduk, err)
					return err
				}

				var kategoriBaru models.KategoriBarang
				if err := tx.Where(&models.KategoriBarang{
					Nama:          d.KategoriBarang[i].Nama,
					IdBarangInduk: d.IdBarangInduk,
				}).
					First(&kategoriBaru).Error; err != nil {
					log.Printf("[ERROR] Gagal mengambil kategori baru '%s': %v", d.KategoriBarang[i].Nama, err)
					return err
				}

				var varianBatch []models.VarianBarang
				for j := 0; j < int(d.KategoriBarang[i].Stok); j++ {
					status := barang_enums.Pending
					if d.KategoriBarang[i].IDRekening != 0 && d.KategoriBarang[i].IDAlamat != 0 {
						status = barang_enums.Ready
					}
					varianBatch = append(varianBatch, models.VarianBarang{
						IdBarangInduk: int32(d.IdBarangInduk),
						IdKategori:    kategoriBaru.ID,
						Status:        status,
						Sku:           d.KategoriBarang[i].Sku,
					})
				}
				if len(varianBatch) > 0 {
					if err := tx.Create(&varianBatch).Error; err != nil {
						log.Printf("[ERROR] Gagal membuat varian kategori '%s': %v",
							d.KategoriBarang[i].Nama, err)
						return err
					}
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] TambahKategoriBarang gagal untuk BarangInduk ID %v: %v", d.IdBarangInduk, err)
		}
	}(dataCopy)

	log.Printf("[INFO] Permintaan tambah kategori diterima untuk BarangInduk ID %d oleh Seller ID %d",
		data.IdBarangInduk, data.IdentitasSeller.IdSeller)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseTambahKategori{
			Message: "Kategori barang berhasil ditambahkan (async-safe).",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Edit Kategori Barang
// Berfungsi untuk mengedit data informasi tentang kategori barang induk yang dituju
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EditKategoriBarang(db *gorm.DB, data PayloadEditKategori) *response.ResponseForm {
	services := "EditKategoriBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var idBarangIndukGet int64 = 0

	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Baang Induk Tidak Ditemukan",
			},
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				data.KategoriBarang[i].Stok = 0
				if err := tx.Model(&models.KategoriBarang{}).
					Where(&models.KategoriBarang{
						ID: data.KategoriBarang[i].ID,
					}).
					Updates(&data.KategoriBarang[i]).Error; err != nil {
					log.Printf("[ERROR] Gagal update kategori ID %d: %v", data.KategoriBarang[i].ID, err)
					return err
				}

				if err := tx.Model(&models.VarianBarang{}).
					Where(&models.VarianBarang{
						IdBarangInduk: data.IdBarangInduk,
						IdKategori:    data.KategoriBarang[i].ID,
					}).
					Update("sku", data.KategoriBarang[i].Sku).Error; err != nil {
					log.Printf("[ERROR] Gagal update SKU varian untuk kategori ID %d: %v", data.KategoriBarang[i].ID, err)
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] EditKategoriBarang gagal update kategori untuk BarangInduk ID %v: %v", idBarangIndukGet, err)
		}
	}()

	log.Printf("[INFO] Kategori barang berhasil diubah pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditKategori{
			Message: "Kategori barang berhasil diubah.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Hapus Kategori Barang Induk
// Berfungsi untuk menghapus kategori barang induk yang ada
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusKategoriBarang(db *gorm.DB, data PayloadHapusKategori) *response.ResponseForm {
	services := "HapusKategoriBarang"

	// Validasi kredensial seller
	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal: Kredensial seller tidak valid",
			},
		}
	}

	// Pastikan barang induk valid untuk seller ini
	var idBarangIndukGet int64
	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk hapus kategori. IdBarangInduk=%d, IdSeller=%d",
			data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal: Barang induk tidak ditemukan",
			},
		}
	}

	// Cek apakah kategori yang akan dihapus masih punya varian dalam transaksi
	for _, kat := range data.KategoriBarang {
		var barangDalamTransaksi int64
		if errStock := db.Model(&models.VarianBarang{}).
			Where("id_barang_induk = ? AND id_kategori = ? AND status IN ?", data.IdBarangInduk, kat.ID, []string{"Dipesan", "Diproses"}).
			Count(&barangDalamTransaksi).Error; errStock != nil {
			log.Printf("[ERROR] Gagal cek stok dalam transaksi untuk kategori ID %d: %v", kat.ID, errStock)
			continue
		}

		if barangDalamTransaksi != 0 {
			log.Printf("[WARN] Tidak bisa hapus kategori %s, masih ada %v stok dalam transaksi", kat.Nama, barangDalamTransaksi)
			return &response.ResponseForm{
				Status:   http.StatusConflict,
				Services: services,
				Payload: response_barang_service.ResponseHapusKategori{
					Message: fmt.Sprintf("Tidak bisa menghapus kategori %s, masih ada %v stok dalam transaksi", kat.Nama, barangDalamTransaksi),
				},
			}
		}
	}

	// Jalankan proses penghapusan di goroutine
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {

			for _, kat := range data.KategoriBarang {
				var existingKategori models.KategoriBarang
				err := tx.Where(&models.KategoriBarang{
					ID:            kat.ID,
					IdBarangInduk: data.IdBarangInduk,
				}).First(&existingKategori).Error

				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						log.Printf("[WARN] Kategori ID %d tidak ditemukan untuk barang induk ID %d", kat.ID, data.IdBarangInduk)
						continue
					}
					log.Printf("[ERROR] Error cek kategori ID %d: %v", kat.ID, err)
					continue
				}

				// Hard delete semua varian di kategori ini
				if err := tx.Unscoped().
					Where("id_kategori = ? AND id_barang_induk = ?", existingKategori.ID, data.IdBarangInduk).
					Delete(&models.VarianBarang{}).Error; err != nil {
					log.Printf("[ERROR] Gagal hapus varian untuk kategori ID %d: %v", existingKategori.ID, err)
					return err
				}

				// Soft delete manual kategori (update kolom deleted_at)
				if err := tx.Model(&models.KategoriBarang{}).
					Where(&models.KategoriBarang{
						ID:            existingKategori.ID,
						IdBarangInduk: data.IdBarangInduk,
					}).
					Delete(&models.KategoriBarang{}).Error; err != nil {
					log.Printf("[ERROR] Gagal update deleted_at kategori ID %d: %v", existingKategori.ID, err)
					return err
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] HapusKategoriBarang gagal proses hapus kategori & varian untuk BarangInduk ID %v: %v",
				idBarangIndukGet, err)
		}
	}()

	log.Printf("[INFO] Kategori barang berhasil dihapus (soft delete manual) pada BarangInduk ID %d oleh Seller ID %d",
		data.IdBarangInduk, data.IdentitasSeller.IdSeller)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusKategori{
			Message: "Kategori barang berhasil dihapus (soft delete manual).",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// STOK BARANG
// ////////////////////////////////////////////////////////////////////////////////

func EditStokBarang(db *gorm.DB, data PayloadEditStokBarang) *response.ResponseForm {
	services := "EditStokBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditStokBarang{
				Message: "Gagal: kredensial seller tidak valid",
			},
		}
	}

	var idBarangIndukGet int64 = 0

	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan. IdBarangInduk=%d, IdSeller=%d, Error=%v", data.IdBarangInduk, data.IdentitasSeller.IdSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditStokBarang{
				Message: "Gagal: Barang induk tidak ditemukan",
			},
		}
	}

	// Validasi setiap kategori yang dikirim oleh client
	for _, b := range data.Barang {
		if err := db.
			Where(&models.KategoriBarang{IdBarangInduk: data.IdBarangInduk, ID: b.IdKategoriBarang}).
			Take(&models.KategoriBarang{}).Error; err != nil {
			log.Printf("[WARN] Kategori tidak ditemukan: IdKategori=%d, NamaKategori=%s, Error=%v", b.IdKategoriBarang, b.NamaKategoriBarang, err)
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  fmt.Sprintf("Kategori %s tidak ditemukan", b.NamaKategoriBarang),
			}
		}
		log.Printf("[INFO] Kategori ditemukan: IdKategori=%d, NamaKategori=%s", b.IdKategoriBarang, b.NamaKategoriBarang)
	}

	// Jalankan perubahan stok secara async (fire-and-forget)
	go func() {
		for _, b := range data.Barang {
			var jumlah int64
			// Hitung varian dengan status Ready atau Pending
			if err := db.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{
					IdBarangInduk: data.IdBarangInduk,
					IdKategori:    b.IdKategoriBarang,
					Status:        barang_enums.Ready,
				}).
				Or(&models.VarianBarang{
					IdBarangInduk: data.IdBarangInduk,
					IdKategori:    b.IdKategoriBarang,
					Status:        barang_enums.Pending,
				}).
				Count(&jumlah).Error; err != nil {
				log.Printf("[ERROR] Gagal menghitung stok: IdKategori=%d, Error=%v", b.IdKategoriBarang, err)
				continue
			}

			// Jika stok sama -> lanjut ke kategori berikutnya
			if jumlah == int64(b.JumlahStok) {
				continue
			}

			// Jika stok yang ada > stok target -> hapus kelebihan varian
			if jumlah > int64(b.JumlahStok) {
				hapus := jumlah - int64(b.JumlahStok)
				log.Printf("[INFO] Perlu hapus varian: IdKategori=%d, JumlahHapus=%d", b.IdKategoriBarang, hapus)
				for j := int64(0); j < hapus; j++ {
					if err := db.
						Where(&models.VarianBarang{
							IdBarangInduk: data.IdBarangInduk,
							IdKategori:    b.IdKategoriBarang,
							Status:        barang_enums.Ready,
						}).
						Or(&models.VarianBarang{
							IdBarangInduk: data.IdBarangInduk,
							IdKategori:    b.IdKategoriBarang,
							Status:        barang_enums.Pending,
						}).
						Limit(1).
						Delete(&models.VarianBarang{}).Error; err != nil {
						log.Printf("[ERROR] Gagal menghapus varian: IdKategori=%d, Iterasi=%d, Error=%v", b.IdKategoriBarang, j, err)
					} else {
						log.Printf("[INFO] Berhasil menghapus varian: IdKategori=%d, Iterasi=%d", b.IdKategoriBarang, j)
					}
				}
			}

			// Jika stok target > stok yang ada -> buat varian baru
			if int64(b.JumlahStok) > jumlah {
				buat := int64(b.JumlahStok) - jumlah
				for j := int64(0); j < buat; j++ {
					var kategori models.KategoriBarang
					if err := db.
						Where(&models.KategoriBarang{IdBarangInduk: data.IdBarangInduk, ID: b.IdKategoriBarang}).
						Take(&kategori).Error; err != nil {
						log.Printf("[ERROR] Gagal mengambil kategori saat pembuatan varian: IdKategori=%d, Iterasi=%d, Error=%v", b.IdKategoriBarang, j, err)
						continue
					}

					newVarian := models.VarianBarang{}
					// Jika kategori belum lengkap data rekening/alamat -> set status Pending
					if kategori.IDRekening == 0 || kategori.IDAlamat == 0 {
						newVarian = models.VarianBarang{
							IdBarangInduk: data.IdBarangInduk,
							IdKategori:    b.IdKategoriBarang,
							Status:        barang_enums.Pending,
							Sku:           b.SkuKategoriBarang,
						}
					} else {
						newVarian = models.VarianBarang{
							IdBarangInduk: data.IdBarangInduk,
							IdKategori:    b.IdKategoriBarang,
							Status:        barang_enums.Ready,
							Sku:           b.SkuKategoriBarang,
						}
					}

					if err := db.Create(&newVarian).Error; err != nil {
						log.Printf("[ERROR] Gagal membuat varian: IdKategori=%d, Iterasi=%d, Error=%v", b.IdKategoriBarang, j, err)
					} else {
						log.Printf("[INFO] Berhasil membuat varian: IdKategori=%d, Iterasi=%d", b.IdKategoriBarang, j)
					}
				}
			}
		}
	}()

	log.Printf("[INFO] Proses update stok sedang berjalan untuk barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditStokBarang{
			Message: "Proses update stok sedang berjalan",
		},
	}
}

func DownStokBarangInduk(db *gorm.DB, data PayloadDownBarangInduk) *response.ResponseForm {
	services := "DownStokBarangInduk"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var count_barang int64
	if err_db := db.Model(&models.BarangInduk{}).Where(&models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdentitasSeller.IdSeller}).Count(&count_barang).Limit(1).Error; err_db != nil {
		log.Printf("[ERROR] Gagal mengambil data barang induk ID %d untuk seller ID %d: %v", data.IdBarangInduk, data.IdentitasSeller.IdSeller, err_db)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count_barang != 1 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk down stok. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal, barang tidak ditemukan.",
			},
		}
	} else {
		go func() {
			if err := db.Transaction(func(tx *gorm.DB) error {
				if err_updates := db.Model(models.VarianBarang{}).
					Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, Status: barang_enums.Ready}).
					Or(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, Status: barang_enums.Terjual}).
					Or(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, Status: barang_enums.Pending}).
					Update("status", barang_enums.Down).Error; err_updates != nil {
					log.Printf("[ERROR] Gagal menurunkan stok semua varian barang induk ID %d: %v", data.IdBarangInduk, err_updates)
					return err_updates
				}
				return nil
			}); err != nil {
				log.Printf("[ERROR] Gagal downkan semua stok barang induk ID %d: %v", data.IdBarangInduk, err)
			}
		}()
	}

	log.Printf("[INFO] Semua stok barang induk ID %d berhasil di-down-kan oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseDownBarang{
			Message: "Berhasil",
		},
	}
}

func DownKategoriBarang(db *gorm.DB, data PayloadDownKategoriBarang) *response.ResponseForm {
	services := "DownKategoriBarang"

	var count_barang int64
	if err_db := db.Model(models.BarangInduk{}).Where(models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdentitasSeller.IdSeller}).Count(&count_barang).Error; err_db != nil {
		log.Printf("[ERROR] Gagal mengambil data barang induk ID %d untuk seller ID %d: %v", data.IdBarangInduk, data.IdentitasSeller.IdSeller, err_db)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count_barang != 1 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk down stok kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, barang tidak ditemukan.",
			},
		}
	} else {
		go func() {
			if err_db := db.Transaction(func(tx *gorm.DB) error {
				if err_updates := db.Model(models.VarianBarang{}).
					Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: data.IdKategoriBarang, Status: barang_enums.Ready}).
					Or(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: data.IdKategoriBarang, Status: barang_enums.Terjual}).
					Or(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: data.IdKategoriBarang, Status: barang_enums.Pending}).
					Update("status", barang_enums.Down).Error; err_updates != nil {
					log.Printf("[ERROR] Gagal menurunkan stok kategori ID %d pada barang induk ID %d: %v", data.IdKategoriBarang, data.IdBarangInduk, err_updates)
					return err_updates
				}
				return nil
			}); err_db != nil {
				log.Printf("[ERROR] Gagal downkan stok kategori ID %d pada barang induk ID %d: %v", data.IdKategoriBarang, data.IdBarangInduk, err_db)
			}
		}()
	}

	log.Printf("[INFO] Semua stok kategori ID %d pada barang induk ID %d berhasil di-down-kan oleh seller ID %d", data.IdKategoriBarang, data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseDownKategori{
			Message: "Berhasil",
		},
	}
}

func EditRekeningBarangInduk(data PayloadEditRekeningBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "EditRekeningBarangInduk"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditRekeningBarangInduk{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	if data.IdRekeningSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditRekeningBarangInduk{
				Message: "Gagal DataRekening Tidak Valid",
			},
		}
	}

	var count int64 = 0
	if err := db.Model(&models.RekeningSeller{}).Where(&models.RekeningSeller{
		ID:       data.IdRekeningSeller,
		IDSeller: data.IdentitasSeller.IdSeller,
	}).Count(&count).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditRekeningBarangInduk{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if count != 1 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditRekeningBarangInduk{
				Message: "Gagal Rekening tidak Ditemukan",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		var hitung int64 = 0
		if err_barang_induk := tx.Model(&models.BarangInduk{}).Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).Count(&hitung).Error; err_barang_induk != nil {
			return err_barang_induk
		}

		if hitung != 1 {
			return fmt.Errorf("gagal data barang induk tidak valis")
		}

		if err_kategori := tx.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
			IdBarangInduk: data.IdBarangInduk,
		}).Update("id_rekening", data.IdRekeningSeller).Error; err_kategori != nil {
			return err_kategori
		}
		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditRekeningBarangInduk{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditRekeningBarangInduk{
			Message: "Berhasil",
		},
	}
}

func EditAlamatGudangBarangInduk(data PayloadEditAlamatBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudangBarangInduk"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, kredensial seller tidak valid.",
			},
		}
	}

	var valid_id int32 = 0
	_ = db.Model(models.BarangInduk{}).Select("id").Where(models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdentitasSeller.IdSeller,
	}).Take(&valid_id)

	if valid_id == 0 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit alamat gudang. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, barang induk tidak ditemukan atau kredensial seller tidak valid.",
			},
		}
	}

	var count int64 = 0

	if errCheck := db.Model(&models.AlamatGudang{}).
		Where(&models.AlamatGudang{
			ID: data.IdAlamatGudang,
		}).Count(&count).Error; errCheck != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count == 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, kredensial alamat tidak valid.",
			},
		}
	}

	if err_edit := db.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
		IdBarangInduk: valid_id,
	}).Update("id_alamat_gudang", data.IdAlamatGudang).Error; err_edit != nil {
		log.Printf("[ERROR] Gagal update alamat gudang untuk semua kategori barang induk ID %d: %v", valid_id, err_edit)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil diubah untuk semua kategori barang induk ID %d oleh seller ID %d", valid_id, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditAlamatBarangInduk{
			Message: "Alamat gudang berhasil diubah.",
		},
	}
}

func EditAlamatGudangBarangKategori(data PayloadEditAlamatBarangKategori, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudangBarangKategori"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, kredensial seller tidak valid.",
			},
		}
	}

	var valid_id int32 = 0
	_ = db.Model(models.BarangInduk{}).Select("id").Where(models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdentitasSeller.IdSeller,
	}).Take(&valid_id)

	if valid_id == 0 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit alamat gudang kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, barang induk tidak ditemukan atau kredensial seller tidak valid.",
			},
		}
	}

	var count int64 = 0

	if errCheck := db.Model(&models.AlamatGudang{}).
		Where(&models.AlamatGudang{
			ID: data.IdAlamatGudang,
		}).Count(&count).Error; errCheck != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count == 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, kredensial alamat tidak valid.",
			},
		}
	}

	if err_edit := db.Model(models.KategoriBarang{}).Where(models.KategoriBarang{
		IdBarangInduk: valid_id,
		ID:            data.IdKategoriBarang,
	}).Update("id_alamat_gudang", data.IdAlamatGudang).Error; err_edit != nil {
		log.Printf("[ERROR] Gagal update alamat gudang untuk kategori ID %d pada barang induk ID %d: %v", data.IdKategoriBarang, valid_id, err_edit)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil diubah untuk kategori ID %d pada barang induk ID %d oleh seller ID %d", data.IdKategoriBarang, valid_id, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditAlamatBarangKategori{
			Message: "Alamat gudang berhasil diubah.",
		},
	}
}
