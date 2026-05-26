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
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/motemen/go-loghttp"
)

const (
	paystackBaseUrl = "https://api.paystack.co"
)

type PaystackClient interface {
	CreateSubAccount(echo.Context, *domains.PaymentMethodDomain, *dtos.SubAccountDto) (*domains.PaystackSubAccountDomain, error)
	DeleteSubAccount(echo.Context, *domains.PaymentMethodDomain, string) (*domains.PaystackSubAccountDomain, error)
	GetBanks(echo.Context, *domains.PaymentMethodDomain, *domains.SettingDomain) ([]*domains.BankDomain, error)
	GetBranches(echo.Context, *domains.PaymentMethodDomain, string, string) ([]*domains.BranchDomain, error)
	GetSubAccount(echo.Context, *domains.PaymentMethodDomain, string) (*domains.PaystackSubAccountDomain, error)
	ListSubAccounts(echo.Context, *domains.PaymentMethodDomain) ([]*domains.PaystackSubAccountDomain, error)
	ResolveAccount(echo.Context, *domains.PaymentMethodDomain, *dtos.PaystackAccountResolvedInformationDto) (*domains.PaystackAccountResolvedInformation, error)
	UpdateSubAccount(echo.Context, *domains.PaymentMethodDomain, string, *dtos.PaystackSubAccountDto) (*domains.PaystackSubAccountDomain, error)
	GetSettlements(echo.Context, *domains.PaymentMethodDomain) (*domains.PagedResult, error)
}

type paystackClient struct {
	http *http.Client
}

func (client paystackClient) UpdateSubAccount(ctx echo.Context, paymentMethod *domains.PaymentMethodDomain, s string, dto *dtos.PaystackSubAccountDto) (*domains.PaystackSubAccountDomain, error) {
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

func (client paystackClient) DeleteSubAccount(_ echo.Context, paymentMethod *domains.PaymentMethodDomain, s string) (*domains.PaystackSubAccountDomain, error) {
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

func (client paystackClient) ResolveAccount(_ echo.Context, paymentMethod *domains.PaymentMethodDomain, dto *dtos.PaystackAccountResolvedInformationDto) (*domains.PaystackAccountResolvedInformation, error) {
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

func (client paystackClient) ListSubAccounts(_ echo.Context, paymentMethod *domains.PaymentMethodDomain) ([]*domains.PaystackSubAccountDomain, error) {
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

func (client paystackClient) GetSubAccount(_ echo.Context, paymentMethod *domains.PaymentMethodDomain, s string) (*domains.PaystackSubAccountDomain, error) {
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

func (client paystackClient) CreateSubAccount(_ echo.Context, paymentMethod *domains.PaymentMethodDomain, dto *dtos.SubAccountDto) (*domains.PaystackSubAccountDomain, error) {
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

func (client paystackClient) GetBanks(ctx echo.Context, paymentMethod *domains.PaymentMethodDomain, settingDomain *domains.SettingDomain) ([]*domains.BankDomain, error) {
	bankUrl := fmt.Sprintf("%s/bank", paystackBaseUrl)
	paystackBankUrl, err := url.Parse(bankUrl)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, paystackBankUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	queries := ctx.QueryParams()
	if queries.Get("currency") == "" {
		log.Printf("Currency value not passed through query param, using setting value = %s\n", settingDomain.Value)
		queries.Set("currency", settingDomain.Value)
	}

	q := paystackBankUrl.Query()
	for k, v := range queries {
		for _, val := range v {
			q.Add(k, val)
		}
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

func (client paystackClient) GetBranches(_ echo.Context, paymentMethod *domains.PaymentMethodDomain, bank string, currency string) ([]*domains.BranchDomain, error) {
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

func (client paystackClient) GetSettlements(ctx echo.Context, paymentMethod *domains.PaymentMethodDomain) (*domains.PagedResult, error) {
	settlementUrl := fmt.Sprintf("%s/settlement", paystackBaseUrl)
	req, err := http.NewRequest(http.MethodGet, settlementUrl, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if page := ctx.QueryParam("page"); page != "" {
		q.Add("page", page)
	}
	if perPage := ctx.QueryParam("perPage"); perPage != "" {
		q.Add("perPage", perPage)
	}
	if from := ctx.QueryParam("from"); from != "" {
		q.Add("from", from)
	}
	if to := ctx.QueryParam("to"); to != "" {
		q.Add("to", to)
	}
	if subaccount := ctx.QueryParam("subaccount"); subaccount != "" {
		q.Add("subaccount", subaccount)
	}
	if status := ctx.QueryParam("status"); status != "" {
		q.Add("status", status)
	}
	req.URL.RawQuery = q.Encode()

	secretKeyValue, err := paymentMethod.GetRequiredConfiguration("secretKey")
	if err != nil {
		return nil, err
	}

	if secretKeyValue.Value == "" {
		return nil, errors.New("unable to get paystack secret key")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKeyValue.Value))

	response, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	d := new(domains.PaystackResponse[[]*domains.SettlementDomain])
	if err := json.NewDecoder(response.Body).Decode(d); err != nil {
		return nil, err
	}

	if !d.Status {
		return nil, errors.New(d.Message)
	}

	pagedResult := &domains.PagedResult{
		Items: d.Data,
	}

	if d.Meta != nil {
		pagedResult.Total = int64(d.Meta.Total)
		pagedResult.Page = int64(d.Meta.Page)
		pagedResult.Size = int64(d.Meta.PerPage)
		pagedResult.TotalPages = int64(d.Meta.PageCount)
	}

	return pagedResult, nil
}

func NewPaystackClient() PaystackClient {
	return &paystackClient{
		http: &http.Client{
			Transport: &loghttp.Transport{},
		},
	}
}
