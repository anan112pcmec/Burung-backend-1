package payment_out_general

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	payment_out_constanta "github.com/anan112pcmec/Burung-backend-1/app/api/payment_out_flip"
)

func GetBalance() (*ResponseFlipGeneralWrapper, error) {
	url := "https://bigflip.id/big_sandbox_api/v2/general/balance"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	output := &ResponseFlipGeneralWrapper{}

	switch res.StatusCode {
	case 200:
		var r ResponseGetBalance
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, err
		}
		output.GetBalance = &r

	case 401:
		var r Response401Error
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, err
		}
		output.Err401 = &r

	case 422:
		var r Response422Error
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, err
		}
		output.Err422 = &r

	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	return output, nil
}

func GetListBank() (*ResponseFlipGeneralWrapper, bool) {
	url := "https://bigflip.id/big_sandbox_api/v2/general/banks"

	wrapper := &ResponseFlipGeneralWrapper{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return wrapper, false
	}

	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return wrapper, false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return wrapper, false
	}

	switch res.StatusCode {

	case 200:
		var success ResponseGetBank
		if err := json.Unmarshal(body, &success); err != nil {
			return wrapper, false
		}
		wrapper.GetBank = &success
		return wrapper, true

	case 401:
		var err401 Response401Error
		_ = json.Unmarshal(body, &err401)
		wrapper.Err401 = &err401
		return wrapper, false

	case 422:
		var err422 Response422Error
		_ = json.Unmarshal(body, &err422)
		wrapper.Err422 = &err422
		return wrapper, false

	default:
		return wrapper, false
	}
}

func GetMaintenanceInfo() (*ResponseFlipGeneralWrapper, bool) {
	url := "https://bigflip.id/big_sandbox_api/v2/general/maintenance"

	wrapper := &ResponseFlipGeneralWrapper{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Request error:", err)
		return wrapper, false
	}

	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("HTTP error:", err)
		return wrapper, false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Read body error:", err)
		return wrapper, false
	}

	switch res.StatusCode {
	case 200:
		var success ResponseMaintenanceInfo // sesuaikan struct MaintenanceInfo
		if err := json.Unmarshal(body, &success); err != nil {
			fmt.Println("Unmarshal error:", err)
			return wrapper, false
		}
		wrapper.GetMaintenance = &success
		return wrapper, true

	case 401:
		var err401 Response401Error
		_ = json.Unmarshal(body, &err401)
		wrapper.Err401 = &err401
		return wrapper, false

	case 422:
		var err422 Response422Error
		_ = json.Unmarshal(body, &err422)
		wrapper.Err422 = &err422
		return wrapper, false

	default:
		fmt.Println("Unexpected status code:", res.StatusCode)
		return wrapper, false
	}
}
