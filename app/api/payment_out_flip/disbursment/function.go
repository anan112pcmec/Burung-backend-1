package payment_out_disbursment

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	payment_out_constanta "github.com/anan112pcmec/Burung-backend-1/app/api/payment_out_flip"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
)

func ReqBankAccountInquiry(data PayloadBankAccountInquiry) (*ResponseDisbursmentWrapper, bool) {

	wrapper := &ResponseDisbursmentWrapper{}
	tautan := "https://bigflip.id/big_sandbox_api/v2/disbursement/bank-account-inquiry"

	// Build x-www-form-urlencoded payload
	payload := url.Values{}
	payload.Set("account_number", data.AccountNumber)
	payload.Set("bank_code", data.BankCode)
	payload.Set("inquiry_key", data.InquiryKey)

	req, err := http.NewRequest("POST", tautan, strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, false
	}

	// Headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", payment_out_constanta.Auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, false
	}
	defer res.Body.Close()

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, false
	}

	switch res.StatusCode {
	case 200:
		var parse ResponseBankAccInquiry
		if err := json.Unmarshal(body, &parse); err != nil {
			return nil, false
		}

		wrapper.ResponseBankAccInq = &parse
		return wrapper, true
	case 401:
		var parse Response401Error
		if err := json.Unmarshal(body, &parse); err != nil {
			return nil, false
		}

		wrapper.Error401 = &parse
		return wrapper, true
	case 422:
		var parse Response422Error
		if err := json.Unmarshal(body, &parse); err != nil {
			return nil, false
		}

		wrapper.Error422 = &parse
		return wrapper, true
	default:
		return nil, false
	}
}

func ReqCreateDisbursment(data PayloadCreateDisbursment) (*ResponseDisbursmentWrapper, bool) {

	wrapper := &ResponseDisbursmentWrapper{}

	tautan := "https://bigflip.id/big_sandbox_api/v3/disbursement"
	method := "POST"

	// --- encode body ---
	formData := url.Values{}
	formData.Set("account_number", data.AccountNumber)
	formData.Set("bank_code", data.BankCode)
	formData.Set("amount", data.Amount)
	formData.Set("remark", data.Remark)
	if data.ReciepentCity != "" {
		formData.Set("recipient_city", payment_out_constanta.CityFlipJawaCode[data.ReciepentCity])
	}
	if data.BeneficiaryEmail != "" {
		formData.Set("beneficiary_email", data.BeneficiaryEmail)
	}
	bodyPayload := strings.NewReader(formData.Encode())

	// --- create request ---
	req, err := http.NewRequest(method, tautan, bodyPayload)
	if err != nil {
		return nil, false
	}

	// --- headers ---
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json; charset=UTF-8")

	// Authorization: Basic <Base64(secretKey)>
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	// idempotency key (harus unik)
	req.Header.Add("idempotency-key", helper.GenerateIdempotencyKey())

	// optional timestamp
	req.Header.Add("X-TIMESTAMP", time.Now().Format(time.RFC3339))

	// --- send request ---
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, false
	}

	switch res.StatusCode {
	case 200:
		var parse ResponseDisbursment
		if err := json.Unmarshal(body, &parse); err != nil {
			return nil, false
		}

		wrapper.ResponseDisbursment = &parse
		return wrapper, true
	case 401:
		var parse Response401Error
		if err := json.Unmarshal(body, &parse); err != nil {
			return nil, false
		}

		wrapper.Error401 = &parse
		return wrapper, true
	case 422:
		var parse Response422Error
		if err := json.Unmarshal(body, &parse); err != nil {
			return nil, false
		}

		wrapper.Error422 = &parse
		return wrapper, true
	default:
		return nil, false
	}
}

func GetDisbursmentByFilter(params map[string]string) (*ResponseDisbursmentWrapper, bool) {
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	url := fmt.Sprintf("https://bigflip.id/big_sandbox_api/v3/get-disbursement?%s", query.Encode())
	wrapper := &ResponseDisbursmentWrapper{}

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
		var parse []ResponseDisbursment
		if err := json.Unmarshal(body, &parse); err != nil {
			return wrapper, false
		}
		wrapper.ResponseAllDisbursment = &parse
		return wrapper, true

	case 401:
		var parse Response401Error
		_ = json.Unmarshal(body, &parse)
		wrapper.Error401 = &parse
		return wrapper, false

	case 422:
		var parse Response422Error
		_ = json.Unmarshal(body, &parse)
		wrapper.Error422 = &parse
		return wrapper, false

	default:
		return wrapper, false
	}
}
