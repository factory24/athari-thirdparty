package payments

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/factory24/athari-thirdparty/pkg/data/domains"
	"github.com/factory24/athari-thirdparty/pkg/data/dtos"
	"github.com/motemen/go-loghttp"
)

type PayFastClient interface {
	RedirectPayment(paymentMethod *domains.PaymentMethodDomain, dto *dtos.PayFastDto) (redirectUrl string, formData map[string]string, err error)
	ValidateITN(paymentMethod *domains.PaymentMethodDomain, itnData url.Values) (bool, error)
}

type payfastClient struct {
	http *http.Client
}

func (client *payfastClient) generateSignature(data map[string]string, passPhrase string) string {

	// Collect keys
	keys := make([]string, 0, len(data))
	for k, v := range data {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build payload
	var payloadBuilder strings.Builder
	for _, key := range keys {
		payloadBuilder.WriteString(fmt.Sprintf("%s=%s&", key, url.QueryEscape(data[key])))
	}
	payload := strings.TrimSuffix(payloadBuilder.String(), "&")

	// Append passphrase if provided
	if passPhrase != "" {
		payload += "&passphrase=" + url.QueryEscape(passPhrase)
	}

	// Hash with MD5
	hasher := md5.New()
	hasher.Write([]byte(payload))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (client *payfastClient) RedirectPayment(paymentMethod *domains.PaymentMethodDomain, dto *dtos.PayFastDto) (redirectUrl string, formData map[string]string, err error) {

	url := os.Getenv("PAYFAST_URL")
	if url == "" {
		url = "https://sandbox.payfast.co.za/eng/process"
	}

	var passPhraseValue string
	passPhrase, err := paymentMethod.GetRequiredConfiguration("passPhrase")
	if err == nil && passPhrase != nil {
		passPhraseValue = passPhrase.Value
	}

	merchantID, err := paymentMethod.GetRequiredConfiguration("merchant_id")
	if err != nil {
		return "", nil, fmt.Errorf("merchant_id not found in payment method configuration")
	}

	merchantKey, err := paymentMethod.GetRequiredConfiguration("merchant_key")
	if err != nil {
		return "", nil, fmt.Errorf("merchant_key not found in payment method configuration")
	}

	payload := map[string]string{
		"merchant_id":  merchantID.Value,
		"merchant_key": merchantKey.Value,
		"return_url":   dto.ReturnURL,
		"cancel_url":   dto.CancelURL,
		"notify_url":   dto.NotifyURL,
		"m_payment_id": strconv.Itoa(dto.MPaymentID),
		"amount":       fmt.Sprintf("%.2f", float64(dto.Amount)),
		"item_name":    dto.ItemName,
	}

	signature := client.generateSignature(payload, passPhraseValue)
	payload["signature"] = signature

	return url, payload, nil

}

func (client *payfastClient) ValidateITN(paymentMethod *domains.PaymentMethodDomain, itnData url.Values) (bool, error) {
	validationUrl := os.Getenv("PAYFAST_VALIDATE_URL")
	if validationUrl == "" {
		validationUrl = "https://sandbox.payfast.co.za/eng/query/validate"
	}

	log.Println("Validating ITN by posting back to PayFast...")
	resp, err := client.http.PostForm(validationUrl, itnData)
	if err != nil {
		return false, fmt.Errorf("failed to post back to payfast for validation: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read validation response body: %w", err)
	}

	responseBody := strings.TrimSpace(string(bodyBytes))
	if responseBody != "VALID" {
		return false, fmt.Errorf("payfast validation check failed, status received: '%s'", responseBody)
	}
	log.Println("PayFast ITN authenticity confirmed.")

	dataForSignature := make(map[string]string)
	for key, values := range itnData {
		if key != "signature" && len(values) > 0 {
			dataForSignature[key] = values[0]
		}
	}

	var passPhraseValue string
	passPhraseConfig, err := paymentMethod.GetRequiredConfiguration("passPhrase")
	if err == nil && passPhraseConfig != nil {
		passPhraseValue = passPhraseConfig.Value
	}

	expectedSignature := client.generateSignature(dataForSignature, passPhraseValue)
	receivedSignature := itnData.Get("signature")

	log.Printf("Validating signature. Expected: %s, Received: %s", expectedSignature, receivedSignature)
	if expectedSignature != receivedSignature {
		return false, fmt.Errorf("signature mismatch")
	}
	log.Println("PayFast ITN signature validated.")

	return true, nil
}

func NewPayFastClient() PayFastClient {
	return &payfastClient{
		http: &http.Client{
			Transport: &loghttp.Transport{},
		},
	}
}
