package kurir_social_media_services

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_social_media_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/social_media_services/response_social_media_services"
)

func EngagementSocialMediaKurir(data PayloadEngageSocialMedia, db *gorm.DB) *response.ResponseForm {
	services := "EngagementSocialMediaKurir"

	if _, status := data.DataIdentitas.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var id_sosmed_table int64 = 0
	_ = db.Model(models.EntitySocialMedia{}).Select("id").Where(models.EntitySocialMedia{
		EntityId:   data.DataIdentitas.IdKurir,
		EntityType: "Kurir",
	}).Take(&id_sosmed_table)

	if id_sosmed_table == 0 {
		if err_buat_kolom := db.Create(&models.EntitySocialMedia{
			EntityId:   data.DataIdentitas.IdKurir,
			Whatsapp:   data.Data.Whatsapp,
			Facebook:   data.Data.Facebook,
			TikTok:     data.Data.TikTok,
			Instagram:  data.Data.Instagram,
			EntityType: "Kurir",
		}).Error; err_buat_kolom != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_social_media_kurir.ResponseEngageSocialMedia{
					Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
				},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_social_media_kurir.ResponseEngageSocialMedia{
				Message: "Berhasil",
			},
		}
	}

	if err_update := db.Model(models.EntitySocialMedia{}).Where(models.EntitySocialMedia{
		ID: id_sosmed_table,
	}).Updates(&models.EntitySocialMedia{
		Whatsapp:  data.Data.Whatsapp,
		Facebook:  data.Data.Facebook,
		TikTok:    data.Data.TikTok,
		Instagram: data.Data.Instagram,
	}).Error; err_update != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_social_media_kurir.ResponseEngageSocialMedia{
				Message: "Gagal, Server Sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}
}
