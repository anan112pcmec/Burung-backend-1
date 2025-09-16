package profiling_services

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type PayloadUpdateProfilingSeller struct {
	DataUpdate models.Seller `json:"data_update_profile"`
}
