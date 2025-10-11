package pengguna_validate

import (
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

func Validate(Id int64, username string, db *gorm.DB) (models.Pengguna, bool) {
	var pengguna models.Pengguna
	if err := db.Model(&models.Pengguna{}).Where(&models.Pengguna{
		ID:       Id,
		Username: username,
	}).Take(&pengguna).Error; err != nil {
		return pengguna, false
	}

	return pengguna, true
}
