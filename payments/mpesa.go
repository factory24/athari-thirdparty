package payments

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/factory24/athari-thirdparty/pkg/data/domains"

	"log"
	"net/http"

	"github.com/motemen/go-loghttp"
)

type mpesaClient struct {
	http *http.Client
}

type MpesaClient interface {
	Register(paymentMethod *domains.PaymentMethodDomain) (interface{}, error)
}
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

func (client *mpesaClient) getAccessToken(paymentMethod *domains.PaymentMethodDomain) (string, error) {
	config, err := getConfigValues(paymentMethod, "consumerKey", "baseUrl", "consumerSecret")
	if err != nil {
		return "", err
	}

	consumerKey := config["consumerKey"]
	mpesaBaseURL := config["baseUrl"]
	consumerSecret := config["consumerSecret"]

	auth := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))

	url := mpesaBaseURL + "/oauth/v1/generate?grant_type=client_credentials"

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	resp := new(AccessTokenResponse)
	if err := json.NewDecoder(response.Body).Decode(resp); err != nil {
		return "", err
	}
	log.Printf("Raw API response: %+v\n", resp) // Log the raw JSON
	return resp.AccessToken, nil
}

func (client *mpesaClient) Register(paymentMethod *domains.PaymentMethodDomain) (interface{}, error) {
	config, err := getConfigValues(paymentMethod, "businessShortCode", "baseUrl")
	if err != nil {
		return "", err
	}
	businessShortCode := config["businessShortCode"]
	baseUrl := config["baseUrl"]

	accessToken, err := client.getAccessToken(paymentMethod)
	if err != nil {
		return nil, err
	}

	url := baseUrl + "/mpesa/c2b/v1/registerurl"
	payload := map[string]interface{}{
		"ShortCode":       businessShortCode,
		"ResponseType":    "Completed",
		"ConfirmationURL": "https://api.1flow.org/v1/billing/webhooks/payment/confirmation",
		"ValidationURL":   "https://api.1flow.org/v1/billing/webhooks/payment/validation",
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	log.Println("Register payload :::::: |", payload)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var apiResponse map[string]interface{} // Use a generic map to store the JSON response
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	log.Printf("Raw API response: %+v\n", apiResponse) // Log the raw JSON

	if response.StatusCode != http.StatusOK {
		// Attempt to extract a more specific error message if possible.
		if errorMsg, ok := apiResponse["errorMessage"].(string); ok {
			return nil, fmt.Errorf("MPESA API error (status %d): %s", response.StatusCode, errorMsg)
		} else if errorData, ok := apiResponse["data"].(map[string]interface{}); ok {
			if errorMsg, ok := errorData["errorMessage"].(string); ok {
				return nil, fmt.Errorf("MPESA API error (status %d): %s", response.StatusCode, errorMsg)
			}
		}

		return nil, fmt.Errorf("MPESA API error (status %d)", response.StatusCode) // Generic error if no specific message found
	}

	return apiResponse, nil
}

func NewMpesaClient() MpesaClient {
	return &mpesaClient{
		http: &http.Client{
			Transport: &loghttp.Transport{},
		},
	}
}

func getConfigValues(paymentMethod *domains.PaymentMethodDomain, keys ...string) (map[string]string, error) {
	configValues := make(map[string]string)

	for _, key := range keys {
		value, err := paymentMethod.GetRequiredConfiguration(key)
		if err != nil || value == nil {
			return nil, fmt.Errorf("missing or invalid %s configuration: %w", key, err)
		}
		log.Printf("Using %s: %s", key, value.Value)
		configValues[key] = value.Value
	}

	return configValues, nil
}
