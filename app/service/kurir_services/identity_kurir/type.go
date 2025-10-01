package identity_kurir

import (
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

type IdentitasKurir struct {
	IdKurir       int64  `json:"id_kurir"`
	UsernameKurir string `json:"username_kurir"`
	EmailKurir    string `json:"email_kurir"`
}

func (ik IdentitasKurir) Validating(db *gorm.DB) (model models.Kurir, status bool) {
	var kurir models.Kurir
	if ik.IdKurir == 0 {
		return kurir, false
	}

	if ik.UsernameKurir == "" {
		return kurir, false
	}

	if ik.EmailKurir == "" {
		return kurir, false
	}

	kurir.ID = 0

	_ = db.Model(models.Kurir{}).Where(models.Kurir{
		ID:       ik.IdKurir,
		Username: ik.UsernameKurir,
		Email:    ik.EmailKurir,
	}).Take(&kurir)

	if kurir.ID == 0 {
		return kurir, false
	}

	return kurir, true
}
