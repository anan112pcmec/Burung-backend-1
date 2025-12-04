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

func ReqBankAccountInquiry[T ResponseBankAccInquiry | Response401Error | Response422Error](data PayloadBankAccountInquiry) (T, bool) {

	var result T

	tautan := "https://bigflip.id/big_sandbox_api/v2/disbursement/bank-account-inquiry"

	// Build x-www-form-urlencoded payload
	payload := url.Values{}
	payload.Set("account_number", data.AccountNumber)
	payload.Set("bank_code", data.BankCode)
	payload.Set("inquiry_key", data.InquiryKey)

	req, err := http.NewRequest("POST", tautan, strings.NewReader(payload.Encode()))
	if err != nil {
		return result, false
	}

	// Headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", payment_out_constanta.Auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result, false
	}
	defer res.Body.Close()

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result, false
	}

	// Unmarshal ke struct generic
	if err := json.Unmarshal(body, &result); err != nil {
		return result, false
	}

	return result, true
}

func ReqCreateDisbursment[T ResponseDisbursment | Response401Error | Response422Error](data PayloadCreateDisbursment) (T, bool) {
	var result T

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
		return result, false
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
		return result, false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result, false
	}

	// --- decode ke generic T ---
	if err := json.Unmarshal(body, &result); err != nil {
		return result, false
	}

	return result, true
}

func GetDisbursmentById[T ResponseDisbursment | Response401Error | Response422Error](data PayloadGetDisburstmentById) (T, bool) {
	var result T
	id := data.Id

	// masukin ID ke query: ?id=xxxxx
	url := fmt.Sprintf("https://bigflip.id/big_sandbox_api/v3/get-disbursement?id=%s", id)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return result, false
	}
	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return result, false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return result, false
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, false
	}

	return result, true
}

func GetDisbursmentByIdempotencyKey[T ResponseDisbursment | Response401Error | Response422Error](data PayloadGetDisburstmentByIdempotencyKey) (T, bool) {
	var result T
	idempotency_key := data.IdempotencyKey

	// masukin ID ke query: ?id=xxxxx
	url := fmt.Sprintf("https://bigflip.id/big_sandbox_api/v3/get-disbursement?idempotency-key=%s", idempotency_key)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return result, false
	}
	req.Header.Add("Accept", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", payment_out_constanta.Auth)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return result, false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return result, false
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, false
	}

	return result, true
}
