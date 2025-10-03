package pengguna_social_media_service

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_social_media_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/social_media_services/response_social_media_services"
)

func EngageSocialMediaPengguna(data PayloadEngageSocialMedia, db *gorm.DB) *response.ResponseForm {
	services := "TambahkanSocialMediaPenguna"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_social_media_pengguna.ResponseEngageSocialMedia{
				Message: "Gagal, Kredensial Pengguna tidak valid",
			},
		}
	}

	var id_sosmed_table int64 = 0
	_ = db.Model(models.EntitySocialMedia{}).Select("id").Where(models.EntitySocialMedia{
		EntityId:   data.IdentitasPengguna.ID,
		EntityType: "Pengguna",
	}).Take(&id_sosmed_table)

	if id_sosmed_table == 0 {
		if err_buat_kolom := db.Create(&models.EntitySocialMedia{
			EntityId:   data.IdentitasPengguna.ID,
			Whatsapp:   data.Data.Whatsapp,
			Facebook:   data.Data.Facebook,
			TikTok:     data.Data.TikTok,
			Instagram:  data.Data.Instagram,
			EntityType: "Pengguna",
		}).Error; err_buat_kolom != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_social_media_pengguna.ResponseEngageSocialMedia{
					Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
				},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_social_media_pengguna.ResponseEngageSocialMedia{
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
			Payload: response_social_media_pengguna.ResponseEngageSocialMedia{
				Message: "Gagal, Server Sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_social_media_pengguna.ResponseEngageSocialMedia{
			Message: "Berhasil",
		},
	}
}
