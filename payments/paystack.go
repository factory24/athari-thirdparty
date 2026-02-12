package payments

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/factory24/athari-thirdparty/pkg/data/domains"
	"github.com/factory24/athari-thirdparty/pkg/data/dtos"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"github.com/motemen/go-loghttp"
)

const (
	paystackBaseUrl = "https://api.paystack.co"
)

type PaystackClient interface {
	CreateSubAccount(*fiber.Ctx, *domains.PaymentMethodDomain, *dtos.SubAccountDto) (*domains.PaystackSubAccountDomain, error)
	DeleteSubAccount(*fiber.Ctx, *domains.PaymentMethodDomain, string) (*domains.PaystackSubAccountDomain, error)
	GetBanks(*fiber.Ctx, *domains.PaymentMethodDomain, *domains.SettingDomain) ([]*domains.BankDomain, error)
	GetBranches(*fiber.Ctx, *domains.PaymentMethodDomain, string, string) ([]*domains.BranchDomain, error)
	GetSubAccount(*fiber.Ctx, *domains.PaymentMethodDomain, string) (*domains.PaystackSubAccountDomain, error)
	ListSubAccounts(*fiber.Ctx, *domains.PaymentMethodDomain) ([]*domains.PaystackSubAccountDomain, error)
	ResolveAccount(*fiber.Ctx, *domains.PaymentMethodDomain, *dtos.PaystackAccountResolvedInformationDto) (*domains.PaystackAccountResolvedInformation, error)
	UpdateSubAccount(*fiber.Ctx, *domains.PaymentMethodDomain, string, *dtos.PaystackSubAccountDto) (*domains.PaystackSubAccountDomain, error)
}

type paystackClient struct {
	http *http.Client
}

func (client paystackClient) UpdateSubAccount(ctx *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, s string, dto *dtos.PaystackSubAccountDto) (*domains.PaystackSubAccountDomain, error) {
	jb, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	log.Println("Creating subaccount on paystack :::::: | ", dto)
	subaccountUrl := fmt.Sprintf("%s/subaccount/%s", paystackBaseUrl, s)
	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request, err := http.NewRequest(http.MethodPut, subaccountUrl, bytes.NewReader(jb))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	accountDomain := new(domains.PaystackResponse[*domains.PaystackSubAccountDomain])
	if err := json.NewDecoder(response.Body).Decode(accountDomain); err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(accountDomain.Message)
	}

	if !accountDomain.Status {
		return nil, errors.New(accountDomain.Message)
	}

	return accountDomain.Data, nil
}

func (client paystackClient) DeleteSubAccount(_ *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, s string) (*domains.PaystackSubAccountDomain, error) {
	bankAccountUrl := fmt.Sprintf("%s/subaccount", paystackBaseUrl)
	resolveAccountUrl, err := url.Parse(bankAccountUrl)
	if err != nil {
		return nil, err
	}

	jb, err := json.Marshal(map[string]any{
		"active": false,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPut, resolveAccountUrl.String(), bytes.NewReader(jb))
	if err != nil {
		return nil, err
	}

	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	d := new(domains.PaystackResponse[*domains.PaystackSubAccountDomain])
	if err := json.NewDecoder(response.Body).Decode(d); err != nil {
		return nil, err
	}

	return d.Data, nil
}

func (client paystackClient) ResolveAccount(_ *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, dto *dtos.PaystackAccountResolvedInformationDto) (*domains.PaystackAccountResolvedInformation, error) {
	resolveUrl := fmt.Sprintf("%s/bank/resolve", paystackBaseUrl)
	resolveAccountUrl, err := url.Parse(resolveUrl)
	if err != nil {
		return nil, err
	}

	resolveBankUrl, err := url.Parse(resolveUrl)
	if err != nil {
		return nil, err
	}

	q := resolveBankUrl.Query()

	for key, value := range dto.ToMap() {
		q.Add(key, value)
	}

	log.Println("Resolve bank URL :::::: |", resolveBankUrl.String())
	request, err := http.NewRequest(http.MethodGet, resolveAccountUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	request.URL.RawQuery = q.Encode()

	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	d := new(domains.PaystackResponse[*domains.PaystackAccountResolvedInformation])
	if err := json.NewDecoder(response.Body).Decode(d); err != nil {
		return nil, err
	}

	return d.Data, nil
}

func (client paystackClient) ListSubAccounts(_ *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain) ([]*domains.PaystackSubAccountDomain, error) {
	subaccountUrl := fmt.Sprintf("%s/subaccount", paystackBaseUrl)
	request, err := http.NewRequest(http.MethodGet, subaccountUrl, nil)
	if err != nil {
		return nil, err
	}
	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	d := new(domains.PaystackResponse[[]*domains.PaystackSubAccountDomain])
	if err := json.NewDecoder(response.Body).Decode(d); err != nil {
		return nil, err
	}

	return d.Data, nil
}

func (client paystackClient) GetSubAccount(_ *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, s string) (*domains.PaystackSubAccountDomain, error) {
	subaccountUrl := fmt.Sprintf("%s/subaccount/%s", paystackBaseUrl, s)
	request, err := http.NewRequest(http.MethodGet, subaccountUrl, nil)
	if err != nil {
		return nil, err
	}
	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	d := new(domains.PaystackResponse[*domains.PaystackSubAccountDomain])
	if err := json.NewDecoder(response.Body).Decode(d); err != nil {
		return nil, err
	}

	return d.Data, nil
}

func (client paystackClient) CreateSubAccount(_ *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, dto *dtos.SubAccountDto) (*domains.PaystackSubAccountDomain, error) {
	d := new(dtos.PaystackSubAccountDto)
	if err := copier.Copy(d, dto); err != nil {
		return nil, err
	}

	jb, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	log.Println("Creating subaccount on paystack :::::: | ", d)
	subaccountUrl := fmt.Sprintf("%s/subaccount", paystackBaseUrl)
	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request, err := http.NewRequest(http.MethodPost, subaccountUrl, bytes.NewReader(jb))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	accountDomain := new(domains.PaystackResponse[*domains.PaystackSubAccountDomain])
	if err := json.NewDecoder(response.Body).Decode(accountDomain); err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusCreated {
		return nil, errors.New(accountDomain.Message)
	}

	if !accountDomain.Status {
		return nil, errors.New(accountDomain.Message)
	}

	return accountDomain.Data, nil
}

func (client paystackClient) GetBanks(ctx *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, settingDomain *domains.SettingDomain) ([]*domains.BankDomain, error) {
	bankUrl := fmt.Sprintf("%s/bank", paystackBaseUrl)
	paystackBankUrl, err := url.Parse(bankUrl)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, paystackBankUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	queries := ctx.Queries()
	if _, ok := queries["currency"]; !ok {
		log.Printf("Currency value not passed through query param, using setting value = %s\n", settingDomain.Value)
		queries["currency"] = settingDomain.Value
	}

	q := paystackBankUrl.Query()
	for key, value := range queries {
		q.Add(key, value)
	}
	request.URL.RawQuery = q.Encode()

	log.Println("Query params :::::: |", paystackBankUrl.Query())
	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))

	log.Println("Calling paystack bank url :::::: |", paystackBankUrl.String())
	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	apiResponse := new(domains.PaystackResponse[[]*domains.BankDomain])
	if err := json.NewDecoder(response.Body).Decode(apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

func (client paystackClient) GetBranches(_ *fiber.Ctx, paymentMethod *domains.PaymentMethodDomain, bank string, currency string) ([]*domains.BranchDomain, error) {
	branchUrl := fmt.Sprintf("%s/bank/branches?currency=%s&bank=%s", paystackBaseUrl, currency, bank)
	request, err := http.NewRequest(http.MethodGet, branchUrl, nil)
	if err != nil {
		return nil, err
	}
	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))

	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	apiResponse := new(domains.PaystackResponse[[]*domains.BranchDomain])
	if err := json.NewDecoder(response.Body).Decode(apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

func NewPaystackClient() PaystackClient {
	return &paystackClient{
		http: &http.Client{
			Transport: &loghttp.Transport{},
		},
	}
}
