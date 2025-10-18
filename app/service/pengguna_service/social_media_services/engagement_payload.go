package pengguna_social_media_service

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/identity_pengguna"

)

type PayloadEngageSocialMedia struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"data_identitas_pengguna"`
	Data              models.EntitySocialMedia           `json:"data_social_media"`
}

type PayloadFollowOrUnfollowSeller struct {
	IdentitasUser identity_pengguna.IdentityPengguna `json:"data_identitas_pengguna"`
	IdSeller      int32                              `json:"id_seller_follow"`
}
