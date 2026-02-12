package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

type request struct {
	MethodName               string      `json:"methodName,omitempty"`
	ResponseTimeoutInSeconds int         `json:"responseTimeoutInSeconds,omitempty"`
	Payload                  interface{} `json:"payload,omitempty"`
}

func (r request) String() string {
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

type Response struct {
	Status  int         `json:"status,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

func (r Response) String() string {
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

type DirectMethodClient interface {
	// Invoke : Parameters deviceId, methodName,
	Invoke(string, string, int, interface{}) (*Response, error)
}

type azureDirectMethodClient struct {
	client *http.Client
}

func (directMethod *azureDirectMethodClient) Invoke(deviceId string, methodName string, timeoutInSeconds int, payload interface{}) (*Response, error) {
	req := &request{
		MethodName:               strcase.ToSnake(methodName),
		ResponseTimeoutInSeconds: timeoutInSeconds,
		Payload:                  payload,
	}

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s.azure-devices.net/twins/%s/methods?api-version=2021-04-12", strings.ToLower(os.Getenv("AZ.IOT_HUB.NAME")), deviceId)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBytes))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", os.Getenv("AZ.IOT_HUB.SAS_TOKEN"))
	if err != nil {
		return nil, err
	}

	log.Println("Request :::: |", req)
	log.Println("Request to send :::: |", request)

	response, err := directMethod.client.Do(request)
	if err != nil {
		return nil, err
	}

	log.Println("DirectMethodClient:Invoke:Status code ::::: |", response.StatusCode)

	resp := new(Response)
	if err := json.NewDecoder(response.Body).Decode(resp); err != nil {
		return nil, err
	}

	log.Println("DirectMethodClient:Invoke:Status code ::::: |", resp.Payload)

	/*
		404: Indicates that either device ID is invalid, or that the device was not online
		upon invocation of a direct method and for connectTimeoutInSeconds
		thereafter (use accompanied error message to understand the root cause);
	*/
	if response.StatusCode == 404 { //

	}

	/*
		504 indicates gateway timeout caused by device not responding
		to a direct method call within responseTimeoutInSeconds.
	*/
	if response.StatusCode == 504 {

	}

	/*
		200 indicates successful execution of direct method;
	*/
	log.Println("Actual response from azure direct method :::::: |", resp)

	return resp, nil
}

func NewAzureDirectMethodClient() DirectMethodClient {
	return &azureDirectMethodClient{client: http.DefaultClient}
}
