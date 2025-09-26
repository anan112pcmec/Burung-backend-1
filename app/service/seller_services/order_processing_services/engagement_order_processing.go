package seller_order_processing_services

import (
	"net/http"
	"sync"

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
	if data.Seller.Validating() != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_order_processing_seller.ResponseApproveTransaksiSeller{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var namaSeller string = ""
	if err_nama := db.Model(models.Seller{}).Select("nama").
		Where(models.Seller{ID: data.Seller.ID, Username: data.Seller.Username}).
		Limit(1).
		Take(&namaSeller).Error; err_nama != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_order_processing_seller.ResponseApproveTransaksiSeller{
				Message: "Gagal Data Seller Tidak Valid",
			},
		}
	}

	if namaSeller == "" {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_order_processing_seller.ResponseApproveTransaksiSeller{
				Message: "Gagal Data Seller Tidak Valid",
			},
		}
	}

	for _, transaksi := range data.DataTransaction {
		wg.Add(1)
		go func(transaksi models.Transaksi) {
			defer wg.Done()
			var approvingstatus response_order_processing_seller.ApprovedStatus
			if err_approve := db.Model(models.Transaksi{}).Where(models.Transaksi{
				IdSeller:      data.Seller.ID,
				IdPengguna:    transaksi.IdPengguna,
				IdBarangInduk: transaksi.IdBarangInduk,
				Status:        "Dibayar",
				Kuantitas:     transaksi.Kuantitas,
				KodeOrder:     transaksi.KodeOrder,
				Total:         transaksi.Total,
			}).Update("status", "Diproses").Error; err_approve != nil {
				approvingstatus.Status = false
			} else {
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
			Message: "Berhasil",
			Hasil:   &approveddata,
		},
	}
}

func UnApproveOrderBarang(data PayloadUnApproveOrder, db *gorm.DB) *response.ResponseForm {
	services := "UnApproveOrderBarang"
	var mu sync.Mutex
	var wg sync.WaitGroup
	if data.Seller.Validating() != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_order_processing_seller.ResponseUnApproveTransaksiSeller{
				Message: "Gagal Data Seller Tidak Valid",
			},
		}
	}

	var namaSeller string = ""
	if err_nama := db.Model(models.Seller{}).Select("nama").
		Where(models.Seller{ID: data.Seller.ID, Username: data.Seller.Username}).
		Limit(1).
		Take(&namaSeller).Error; err_nama != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_order_processing_seller.ResponseUnApproveTransaksiSeller{
				Message: "Gagal Data Seller Tidak Valid",
			},
		}
	}

	var unApprovedData []response_order_processing_seller.UnApprovedStatus
	for _, transaksi := range data.DataTransaction {
		wg.Add(1)
		go func(transaksi models.Transaksi) {
			defer wg.Done()
			var unapprovingstatus response_order_processing_seller.UnApprovedStatus
			if err_approve := db.Model(models.Transaksi{}).Where(models.Transaksi{
				IdSeller:      data.Seller.ID,
				IdPengguna:    transaksi.IdPengguna,
				IdBarangInduk: transaksi.IdBarangInduk,
				Status:        "Dibayar",
				Kuantitas:     transaksi.Kuantitas,
				KodeOrder:     transaksi.KodeOrder,
			}).Update("status", "Dibatalkan").Error; err_approve != nil {
				unapprovingstatus.Status = false
			} else {
				unapprovingstatus.Status = true
			}
			unapprovingstatus.DataUnApproved = transaksi
			mu.Lock()
			unApprovedData = append(unApprovedData, unapprovingstatus)
			mu.Unlock()
		}(transaksi)
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_order_processing_seller.ResponseUnApproveTransaksiSeller{
			Message: "Berhasil",
			Hasil:   &unApprovedData,
		},
	}

}
