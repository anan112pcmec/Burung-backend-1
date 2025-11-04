package open_route_direction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Getenvi(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var ApiKey = Getenvi("OPEN_ROUTE_KEY", "apalah")

func HitungJarakHargaDirection(TitikMulai [2]float64, TitikSelesai [2]float64) (string, int64, bool) {
	var (
		Jarak  string = ""
		Harga  int64  = 0
		Status bool   = false
	)

	// --- Siapkan request body ---
	reqBody := RequestDirection{
		Koordinat: [][2]float64{
			TitikMulai, TitikSelesai,
		},
		Bahasa:       "id",
		Satuan:       "km",
		Geometry:     true,
		Instructions: false,
	}

	// --- Encode ke JSON ---
	body, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Gagal encode JSON:", err)
		return Jarak, Harga, false
	}

	// --- Kirim request ke ORS ---
	url := "https://api.openrouteservice.org/v2/directions/driving-car"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Gagal buat request:", err)
		return Jarak, Harga, false
	}

	req.Header.Set("Authorization", ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Gagal kirim request:", err)
		return Jarak, Harga, false
	}
	defer resp.Body.Close()

	// --- Baca respon body ---
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error dari API:", string(data))
		return Jarak, Harga, false
	}

	// --- Decode hasil JSON ---
	var hasil ResponseDirection
	if err := json.Unmarshal(data, &hasil); err != nil {
		fmt.Println("Gagal decode JSON:", err)
		return Jarak, Harga, false
	}

	// --- Ambil jarak total ---
	if len(hasil.Features) > 0 {
		distanceMeter := hasil.Features[0].Properties.Summary.Distance
		jarakKm := distanceMeter / 1000

		Jarak = fmt.Sprintf("%.2f km", jarakKm)
		Harga = int64(jarakKm * 1000) // Rp 1000 per km
		Status = true
	}

	return Jarak, Harga, Status
}
