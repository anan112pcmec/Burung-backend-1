package maintain_cache

import (
	"log"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	data_cache "github.com/anan112pcmec/Burung-backend-1/app/cache/data"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

func DataAlamatEkspedisiUp(db *gorm.DB) {
	var idsAlamatEks []int64
	if err := db.Model(&models.AlamatEkspedisi{}).Pluck("id", &idsAlamatEks).Error; err != nil {
		return
	}

	var alamatEkspedisiData []models.AlamatEkspedisi
	if err := db.Where("id IN ?", idsAlamatEks).Find(&alamatEkspedisiData).Error; err != nil {
		return
	}

	if data_cache.DataAlamatEkspedisi == nil {
		data_cache.DataAlamatEkspedisi = make(map[string]map[int64]models.AlamatEkspedisi, len(alamatEkspedisiData))
	}

	for i := range alamatEkspedisiData {
		kota := alamatEkspedisiData[i].Kota
		if _, ok := data_cache.DataAlamatEkspedisi[kota]; !ok {
			data_cache.DataAlamatEkspedisi[kota] = make(map[int64]models.AlamatEkspedisi)
		}
		data_cache.DataAlamatEkspedisi[kota][alamatEkspedisiData[i].ID] = alamatEkspedisiData[i]
	}
}

func DataAlamatEkspedisiChange(DataEkspedisi models.AlamatEkspedisi, kota string, id int64, db *gorm.DB) {
	data_cache.DataAlamatEkspedisi[kota][id] = DataEkspedisi
}

func DataOperasionalPengirimanUp() {
	// Set path dan nama file
	v := viper.New()
	v.SetConfigFile("C:/Burung_App/Project_Source/Backend-1/app/operational_data/pengiriman_burung.json")
	v.SetConfigType("json") // karena file json

	// Baca config
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Gagal membaca file config: %v", err)
		return
	}
	// Unmarshal JSON ke struct global
	if err := v.Unmarshal(&data_cache.OperationalPengirimanData); err != nil {
		log.Printf("Gagal unmarshal JSON ke struct: %v", err)
		return
	}

	log.Println("OperationalPengirimanData berhasil di-update dari JSON")

	data_cache.DataTarifPengiriman.TarifKurirPerKm = data_cache.OperationalPengirimanData.CommittedOperationalData.DataTarifPengiriman.TarifKurirPerKm
	data_cache.DataTarifPengiriman.TarifKurirPerKg = data_cache.OperationalPengirimanData.CommittedOperationalData.DataTarifPengiriman.TarifKurirPerKg
	data_cache.DataTarifPengiriman.TarifSistem = data_cache.OperationalPengirimanData.CommittedOperationalData.DataTarifPengiriman.TarifSistem

	// Isi DataTarifJenisPengiriman dari JSON
	data_cache.DataTarifJenisPengiriman = make(map[string]data_cache.JenisPengiriman, len(data_cache.OperationalPengirimanData.CommittedOperationalData.DataJenisPengiriman))

	for _, jenis := range data_cache.OperationalPengirimanData.CommittedOperationalData.DataJenisPengiriman {
		data_cache.DataTarifJenisPengiriman[jenis.Nama] = jenis
	}
}
