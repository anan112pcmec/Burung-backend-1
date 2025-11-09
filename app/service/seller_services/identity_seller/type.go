package identity_seller

import (
	"context"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

type IdentitySeller struct {
	IdSeller    int32  `json:"id_seller"`
	Username    string `json:"username_seller"`
	EmailSeller string `json:"email_seller"`
}

func (i IdentitySeller) Validating(ctx context.Context, db *gorm.DB) (model models.Seller, status bool) {
	var seller models.Seller
	if i.IdSeller == 0 {
		return seller, false
	}

	if i.Username == "" {
		return seller, false
	}

	if i.EmailSeller == "" {
		return seller, false
	}

	_ = db.WithContext(ctx).Model(&models.Seller{}).Where(&models.Seller{
		ID:       i.IdSeller,
		Username: i.Username,
		Email:    i.EmailSeller,
	}).Limit(1).Take(&seller)

	if seller.ID == 0 {
		return seller, false
	}

	if seller.StatusSeller != "Online" {
		return seller, false
	}

	return seller, true
}
