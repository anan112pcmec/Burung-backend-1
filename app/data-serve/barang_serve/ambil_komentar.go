package barang_serve

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func AmbilKomentarBarangInduk(IdBarangInduk int32, db *gorm.DB) *response.ResponseForm {
	services := "AmbilKomentarBarang"
	var komentar models.Komentar

	if err := db.Model(&models.Komentar{}).Where(&models.Komentar{
		IdBarangInduk: IdBarangInduk,
	}).Find(&komentar).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Komentar Tidak Ditemukan",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  komentar,
	}
}
