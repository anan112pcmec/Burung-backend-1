package seller_social_media_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"
)

type PayloadEngageSocialMedia struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	Data            models.EntitySocialMedia       `json:"data_social_media"`
}
