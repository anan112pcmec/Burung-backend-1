package kurir_profiling_service

import (
	"net/http"
	"sync"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	particular_profiling_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/profiling_services/particular_profiling"
	response_profiling_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/profiling_services/response_profiling"
)

func PersonalProfilingKurir(data PayloadPersonalProfilingKurir, db *gorm.DB) *response.ResponseForm {
	var wg sync.WaitGroup
	services := "GeneralProfilingKurir"

	if data.DataKredensial.IDkurir == 0 && data.DataKredensial.UsernameKurir == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var hasilresponsenama particular_profiling_kurir.ResponseUbahNama
	var hasilresponseusername particular_profiling_kurir.ResponseUbahUsername
	var hasilresponseemail particular_profiling_kurir.ResponseUbahGmail

	if data.Nama != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasilresponsenama = particular_profiling_kurir.UbahNama(data.DataKredensial.IDkurir, data.DataKredensial.UsernameKurir, data.Nama, db)
		}()
	}

	if data.Username != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasilresponseusername = particular_profiling_kurir.UbahUsernameKurir(db, data.DataKredensial.IDkurir, data.Username)
		}()
	}

	if data.Email != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasilresponseemail = particular_profiling_kurir.UbahEmail(data.DataKredensial.IDkurir, data.Username, data.Email, db)
		}()
	}

	wg.Wait()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_profiling_kurir.ResponseProfilingPersonalKurir{
			UpdateNama:     hasilresponsenama,
			UpdateUsername: hasilresponseusername,
			UpdateEmail:    hasilresponseemail,
		},
	}
}

func GeneralProfilingKurir(data PayloadGeneralProfiling, db *gorm.DB) *response.ResponseForm {
	services := "GeneralProfilingKurir"
	var hasil_update_deskripsi particular_profiling_kurir.ResponseUbahDeskripsi

	_, status := data.DataIdentitas.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Deskripsi != "" {
		hasil_update_deskripsi = particular_profiling_kurir.UbahDeskripsi(data.DataIdentitas.IdKurir, data.DataIdentitas.UsernameKurir, data.Deskripsi, db)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_profiling_kurir.ResponseProfilingGeneralKurir{
			UpdateDeskripsi: hasil_update_deskripsi,
		},
	}

}
