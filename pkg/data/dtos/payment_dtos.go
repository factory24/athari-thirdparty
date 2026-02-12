package dtos

import (
	"encoding/json"
	"fmt"

	"github.com/iancoleman/strcase"
)

type PaymentTypeDto struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"required"`
	RoleId      string `json:"roleId"`
	AccountType string `json:"accountType"`
	IsActive    bool   `json:"isActive"`
}

type SubAccountDto struct {
	BusinessName        string  `json:"businessName,omitempty"`
	SettlementBank      int     `json:"settlementBank,omitempty"`
	AccountNumber       string  `json:"accountNumber,omitempty"`
	PercentageCharge    float64 `json:"percentageCharge"`
	Currency            string  `json:"currency,omitempty"`
	Description         string  `json:"description,omitempty"`
	PrimaryContactName  string  `json:"primaryContactName,omitempty"`
	PrimaryContactEmail string  `json:"primaryContactEmail,omitempty"`
	PrimaryContactPhone string  `json:"primaryContactPhone,omitempty"`
	SettlementSchedule  string  `json:"settlement_schedule"`
}

type PaystackSubAccountDto struct {
	BusinessName        string  `json:"business_name," validate:"required"`
	SettlementBank      int     `json:"settlement_bank," validate:"required"`
	AccountNumber       string  `json:"account_number," validate:"required"`
	PercentageCharge    float64 `json:"percentage_charge," validate:"required"`
	Currency            string  `json:"currency" validate:"required"`
	BranchCode          string  `json:"branch_code"`
	Description         string  `json:"description"`
	PrimaryContactName  string  `json:"primary_contact_name"`
	PrimaryContactEmail string  `json:"primary_contact_email"`
	PrimaryContactPhone string  `json:"primary_contact_phone"`
	SettlementSchedule  string  `json:"settlement_schedule"`
	Metadata            string  `json:"metadata"`
	Active              bool    `json:"active,omitempty"`
}

type PaystackAccountResolvedInformationDto struct {
	AccountNumber string `json:"accountNumber,omitempty" validate:"required"`
	BankCode      string `json:"bankCode,omitempty" validate:"required"`
	Currency      string `json:"currency"`
}

type PayFastDto struct {
	MerchantID  int    `json:"merchantId" validate:"required"`
	MerchantKey string `json:"merchantKey" validate:"required"`
	ReturnURL   string `json:"returnUrl"`
	CancelURL   string `json:"cancelUrl"`
	NotifyURL   string `json:"notifyUrl"`
	MPaymentID  int    `json:"mPaymentId" validate:"required"`
	Amount      int    `json:"amount" validate:"required"`
	ItemName    string `json:"itemName" validate:"required"`
}

func (d PaystackAccountResolvedInformationDto) ToMap() map[string]string {
	data, err := json.Marshal(d)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil
	}

	stringMap := make(map[string]string)
	for k, v := range result {
		stringMap[strcase.ToSnake(k)] = fmt.Sprintf("%v", v)
	}

	return stringMap
}

func (d PaystackSubAccountDto) String() string {
	jb, _ := json.Marshal(d)
	return string(jb)
}

func (d PaymentTypeDto) String() string {
	jb, _ := json.Marshal(d)
	return string(jb)
}

func (d SubAccountDto) String() string {
	jb, _ := json.Marshal(d)
	return string(jb)
}

type PaymentGatewayDto struct {
	Name                 string                    `json:"name" validate:"required"`
	DisplayName          string                    `json:"displayName" validate:"required"`
	Description          string                    `json:"description"`
	Enabled              bool                      `json:"enabled"`
	PaymentGatewayFields []*PaymentGatewayFieldDto `json:"fields"`
}

type GetCodeDto struct {
	Id     string `json:"id" validate:"required"`
	Entity string `json:"entity" validate:"required"`
}

type PaymentGatewayFieldDto struct {
	Label    string `json:"label,omitempty"`
	Name     string `json:"name,omitempty"`
	Required bool   `json:"required"`
}

type PaymentMethodDto struct {
	IsEnabled        bool                             `json:"isEnabled"`
	IsType           string                           `json:"isType" validate:"oneof=default subaccount"`
	PaymentGatewayId string                           `json:"paymentGatewayId,omitempty" validate:"required"`
	Configurations   []*PaymentMethodConfigurationDto `json:"configurations,omitempty"`
	SubAccount       *SubAccountDto                   `json:"subAccount,omitempty"`
	Name             string                           `json:"name" `
}

type PaymentMethodConfigurationDto struct {
	Key   string `json:"key,omitempty" validate:"required"`
	Value string `json:"value,omitempty" validate:"required"`
}

type PaymentMethodIdsDto struct {
	Ids []string `json:"ids" validate:"required"`
}

type CountryGatewayMappingDto struct {
	PaymentGatewayID string   `json:"paymentGatewayId" validate:"required"`
	Countries        []string `json:"countries" validate:"required"`
}

type CountryBatchDto struct {
	CountryCodes []string `json:"codes,omitempty" validator:"required"`
}
