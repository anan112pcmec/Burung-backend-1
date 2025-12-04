package payment_out_general

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	payment_out_constanta "github.com/anan112pcmec/Burung-backend-1/app/api/payment_out_flip"

func GetBalance[T Response401Error | ResponseGetBalance]() T {
	var result T

	url := "https://bigflip.id/big_sandbox_api/v2/general/balance"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result
	}

	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result
	}

	if res.StatusCode != 200 {
		return result
	}

	return result
}

func GetListBank[T ResponseGetBank | Response401Error | Response422Error]() T {
	var result T
	url := "https://bigflip.id/big_sandbox_api/v2/general/banks"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result
	}

	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	res, err := client.Do(req)
	if err != nil {
		return result
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result
	}

	switch res.StatusCode {
	case 200:
		if err := json.Unmarshal(body, &result); err != nil {
			return result
		}
		return result

	case 401:

		_ = json.Unmarshal(body, &result)
		return result

	case 422:
		_ = json.Unmarshal(body, &result)
		return result

	default:
		return result
	}
}

func GetMaintenanceInfo[T ResponseMaintenanceInfo | Response401Error]() T {
	var result T
	url := "https://bigflip.id/big_sandbox_api/v2/general/maintenance"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return result
	}
	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return result
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result
	}

	switch res.StatusCode {
	case 200:
		if err := json.Unmarshal(body, &result); err != nil {
			return result
		}
		return result

	case 401:

		_ = json.Unmarshal(body, &result)
		return result

	case 422:
		_ = json.Unmarshal(body, &result)
		return result

	default:
		return result
	}

}
