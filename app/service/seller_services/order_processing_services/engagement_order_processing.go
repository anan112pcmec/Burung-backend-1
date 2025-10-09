package seller_order_processing_services

import (
	"log"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_order_processing_seller "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/order_processing_services/response_order_processing"
)

func ApproveOrderBarang(data PayloadApproveOrder, db *gorm.DB) *response.ResponseForm {
	services := "ApproveOrderBarang"
	var mu sync.Mutex
	var wg sync.WaitGroup

	var approveddata []response_order_processing_seller.ApprovedStatus

	seller, status := data.IdentitasSeller.Validating(db)
	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_order_processing_seller.ResponseApproveTransaksiSeller{
				Message: "Gagal, kredensial seller tidak valid.",
			},
		}
	}
	for _, transaksi := range data.DataTransaction {
		wg.Add(1)
		go func(transaksi models.Transaksi) {
			defer wg.Done()
			var approvingstatus response_order_processing_seller.ApprovedStatus
			if err_approve := db.Model(models.Transaksi{}).Where(models.Transaksi{
				IdSeller:      seller.ID,
				IdPengguna:    transaksi.IdPengguna,
				IdBarangInduk: transaksi.IdBarangInduk,
				Status:        "Dibayar",
				Kuantitas:     transaksi.Kuantitas,
				KodeOrder:     transaksi.KodeOrder,
				Total:         transaksi.Total,
			}).Update("status", "Diproses").Error; err_approve != nil {
				log.Printf("[ERROR] Gagal approve transaksi ID %d: %v", transaksi.ID, err_approve)
				approvingstatus.Status = false
			} else {
				log.Printf("[INFO] Transaksi ID %d berhasil di-approve oleh seller ID %d", transaksi.ID, seller.ID)
				approvingstatus.Status = true
			}
			approvingstatus.DataApproved = transaksi
			mu.Lock()
			approveddata = append(approveddata, approvingstatus)
			mu.Unlock()
		}(transaksi)
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_order_processing_seller.ResponseApproveTransaksiSeller{
			Message: "Proses approve transaksi selesai.",
			Hasil:   &approveddata,
		},
	}
}

func UnApproveOrderBarang(data PayloadUnApproveOrder, db *gorm.DB) *response.ResponseForm {
	services := "UnApproveOrderBarang"
	var mu sync.Mutex
	var wg sync.WaitGroup

	seller, status := data.IdentitasSeller.Validating(db)
	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_order_processing_seller.ResponseUnApproveTransaksiSeller{
				Message: "Gagal, kredensial seller tidak valid.",
			},
		}
	}

	var namaSeller string = ""
	if err_nama := db.Model(models.Seller{}).Select("nama").
		Where(models.Seller{ID: seller.ID, Username: seller.Username}).
		Limit(1).
		Take(&namaSeller).Error; err_nama != nil {
		log.Printf("[WARN] Data seller tidak valid untuk ID %d", seller.ID)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_order_processing_seller.ResponseUnApproveTransaksiSeller{
				Message: "Gagal, data seller tidak valid.",
			},
		}
	}

	var unApprovedData []response_order_processing_seller.UnApprovedStatus

	for _, transaksi := range data.DataTransaction {
		wg.Add(1)
		go func(transaksi models.Transaksi) {
			defer wg.Done()
			if err := db.Transaction(func(tx *gorm.DB) error {
				var unapprovingstatus response_order_processing_seller.UnApprovedStatus

				if errUpdate := tx.Model(&models.Transaksi{}).Where(&models.Transaksi{
					IdSeller:      seller.ID,
					IdPengguna:    transaksi.IdPengguna,
					IdBarangInduk: transaksi.IdBarangInduk,
					Status:        "Dibayar",
					Kuantitas:     transaksi.Kuantitas,
					KodeOrder:     transaksi.KodeOrder,
				}).Limit(int(transaksi.Kuantitas)).Update("status", "Dibatalkan").Error; errUpdate != nil {
					log.Printf("[ERROR] Gagal unapprove transaksi ID %d: %v", transaksi.ID, errUpdate)
					unapprovingstatus.Status = false
				} else {
					log.Printf("[INFO] Transaksi ID %d berhasil di-unapprove oleh seller ID %d", transaksi.ID, seller.ID)
					unapprovingstatus.Status = true
				}

				unapprovingstatus.DataUnApproved = transaksi
				mu.Lock()
				unApprovedData = append(unApprovedData, unapprovingstatus)
				mu.Unlock()

				if errBatal := tx.Create(&models.BatalTransaksi{
					IdTransaksi:    transaksi.ID,
					DibatalkanOleh: "Seller",
					Alasan:         data.Alasan,
					CreatedAt:      time.Now(),
				}).Error; errBatal != nil {
					log.Printf("[ERROR] Gagal mencatat pembatalan transaksi ID %d: %v", transaksi.ID, errBatal)
					return errBatal
				}

				return nil
			}); err != nil {
				log.Printf("[ERROR] Gagal membatalkan transaksi ID %d: %v", transaksi.ID, err)
			}
		}(transaksi)
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_order_processing_seller.ResponseUnApproveTransaksiSeller{
			Message: "Proses unapprove transaksi selesai.",
			Hasil:   &unApprovedData,
		},
	}

}
