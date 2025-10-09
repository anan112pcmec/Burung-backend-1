package kurir_social_media_services

import (
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_social_media_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/social_media_services/response_social_media_services"
)

func EngagementSocialMediaKurir(data PayloadEngageSocialMedia, db *gorm.DB) *response.ResponseForm {
	services := "EngagementSocialMediaKurir"

	if _, status := data.DataIdentitas.Validating(db); !status {
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.DataIdentitas.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_social_media_kurir.ResponseEngageSocialMedia{
				Message: "Kredensial kurir tidak valid.",
			},
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
			log.Printf("[ERROR] Gagal menambah data social media untuk kurir ID %d: %v", data.DataIdentitas.IdKurir, err_buat_kolom)
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_social_media_kurir.ResponseEngageSocialMedia{
					Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
				},
			}
		}

		log.Printf("[INFO] Data social media berhasil ditambahkan untuk kurir ID %d", data.DataIdentitas.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_social_media_kurir.ResponseEngageSocialMedia{
				Message: "Data social media berhasil ditambahkan.",
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
		log.Printf("[ERROR] Gagal memperbarui data social media untuk kurir ID %d: %v", data.DataIdentitas.IdKurir, err_update)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_social_media_kurir.ResponseEngageSocialMedia{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Data social media berhasil diperbarui untuk kurir ID %d", data.DataIdentitas.IdKurir)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_social_media_kurir.ResponseEngageSocialMedia{
			Message: "Data social media berhasil diperbarui.",
		},
	}
}
