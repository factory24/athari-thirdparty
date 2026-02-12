package domains

import (
	"encoding/json"
	"fmt"
	"time"
)

type PaystackResponse[T any] struct {
	Status  bool   `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type BankDomain struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Code        string     `json:"code"`
	LongCode    string     `json:"longcode"`
	Gateway     *string    `json:"gateway,omitempty"` // Optional field using pointer
	PayWithBank bool       `json:"pay_with_bank"`
	Active      bool       `json:"active"`
	IsDeleted   bool       `json:"is_deleted"`
	Country     string     `json:"country"`
	Currency    string     `json:"currency"`
	Type        string     `json:"type"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"` // Optional field using pointer
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"` // Optional fields using pointer
}

type BranchDomain struct {
	ID         string `json:"id"`
	BranchCode string `json:"branch_code"`
	BranchName string `json:"branch_name"`
	BankID     int    `json:"bank"` // Assuming bank is an ID
	Currency   string `json:"currency"`
	IsDeleted  bool   `json:"is_deleted"`
}

type PaymentTypeDomain struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type"`
	RoleId string `json:"roleId"`
}

type PaymentSubAccountDomain struct {
	UserID              string  `json:"userId,omitempty"`
	Code                string  `json:"code,omitempty"`
	UserType            string  `json:"userType,omitempty"`
	BusinessName        string  `json:"businessName,omitempty"`
	SettlementBank      int     `json:"settlementBank,omitempty"`
	AccountNumber       string  `json:"accountNumber,omitempty"`
	PercentageCharge    float64 `json:"percentageCharge,omitempty"`
	Description         string  `json:"description,omitempty"`
	PrimaryContactName  string  `json:"primaryContactName,omitempty"`
	PrimaryContactEmail string  `json:"primaryContactEmail,omitempty"`
	PrimaryContactPhone string  `json:"primaryContactPhone,omitempty"`
	SettlementSchedule  string  `json:"settlement_schedule,omitempty"`
}

type SettingDomain struct {
	SettingID   string `json:"settingId,omitempty"`
	Name        string `json:"name,omitempty"`
	Code        string `json:"code,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
	IsEnabled   bool   `json:"isEnabled,omitempty"`
}

func (d SettingDomain) String() string {
	jb, _ := json.Marshal(d)
	return string(jb)
}

type PaystackSubAccountDomain struct {
	ID                  int        `json:"id,omitempty"`
	Active              bool       `json:"active"`
	IsVerified          bool       `json:"is_verified"`
	CreatedAt           *time.Time `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt"`
	Code                string     `json:"subaccount_code,omitempty"`
	BusinessName        string     `json:"business_name,omitempty"`
	SettlementBank      string     `json:"settlement_bank,omitempty"`
	AccountNumber       string     `json:"account_number,omitempty"`
	PercentageCharge    float64    `json:"percentage_charge,omitempty"`
	Description         string     `json:"description,omitempty"`
	PrimaryContactName  string     `json:"primary_contact_name,omitempty"`
	PrimaryContactEmail string     `json:"primary_contact_email,omitempty"`
	PrimaryContactPhone string     `json:"primary_contact_phone,omitempty"`
	SettlementSchedule  string     `json:"settlement_schedule,omitempty"`
}

func (d PaystackSubAccountDomain) String() string {
	jb, _ := json.Marshal(d)
	return string(jb)
}

func (d PaymentSubAccountDomain) String() string {
	jb, _ := json.Marshal(d)
	return string(jb)
}

type PaystackAccountResolvedInformation struct {
	AccountName   string `json:"account_name,omitempty"`
	AccountNumber string `json:"account_number,omitempty"`
	BankId        int    `json:"bank_id,omitempty"`
}

type PaymentAccountInformationDomain struct {
	AccountName   string `json:"accountName,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	BankId        int    `json:"bankId,omitempty"`
}

type PaymentGatewayDomain struct {
	ID                   string                 `json:"id,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	Code                 string                 `json:"code,omitempty"`
	DisplayName          string                 `json:"displayName,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Enabled              bool                   `json:"enabled,omitempty"`
	Logo                 string                 `json:"logo,omitempty"`
	PaymentGatewayFields []*PaymentGatewayField `json:"fields,omitempty"`
}

type PaymentGatewayField struct {
	Label    string `json:"label,omitempty"`
	Name     string `json:"name,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type PaymentGatewayCountriesDomain struct {
	PaymentGateway *PaymentGatewayDomain `json:"paymentGateway"`
	Countries      []*CountryDomain      `json:"countries,omitempty"`
}

type PaymentMethodConfigurationDomain struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type RolePermissionAccessLevelsDomain struct {
	Role        *RoleDomain               `json:"role"`
	Permissions []*ModulePermissionDomain `json:"permissions"`
}
type ModuleDomain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PermissionDomain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ModulePermissionDomain struct {
	Module       *ModuleDomain     `json:"module"`
	Permission   *PermissionDomain `json:"permission"`
	ReadAccess   bool              `json:"readAccess"`
	CreateAccess bool              `json:"createAccess"`
	UpdateAccess bool              `json:"updateAccess"`
}
type RoleDomain struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	BusinessId  string      `json:"businessId"`
	CreatedBy   *UserDomain `json:"createdBy"`
	RoleType    string      `json:"roleType"`
}
type UserDomain struct {
	ID           string `json:"id,omitempty"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	Username     string `json:"username,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	AccountType  string `json:"accountType,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Country      string `json:"country,omitempty"`
	Medium       string `json:"medium,omitempty"`
	OrgId        string `json:"orgId,omitempty"`
	IsStaff      bool   `json:"isStaff,omitempty"`
	Role         string `json:"role,omitempty"`
	RoleId       string `json:"roleId,omitempty"`
	BusinessId   string `json:"businessId,omitempty"`
}

type ZoneAttributeDomain struct {
	Name  string `json:"name,omitempty"`
	Code  string `json:"code"`
	Value string `json:"value"`
}

type ZoneMappedAttributeDomain struct {
	Zone      *ZoneDomain          `json:"zone"`
	Attribute *ZoneAttributeDomain `json:"settings"`
}

type ZoneDomain struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Location       string      `json:"location"`
	FacilityType   string      `json:"facilityType"`
	Area           string      `json:"area"`
	Latitude       float64     `json:"latitude"`
	Longitude      float64     `json:"longitude"`
	FlowboxCount   int         `json:"flowboxCount,omitempty"`
	FlowmeterCount int         `json:"flowmeterCount,omitempty"`
	Currency       string      `json:"currency"`
	CardsCount     int         `json:"cardsCount,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	CreatedBy      *UserDomain `json:"createdBy,omitempty"`
}

type PaymentMethodDomain struct {
	ID                string                              `json:"id,omitempty"`
	PaymentGateway    *PaymentGatewayDomain               `json:"paymentGateway,omitempty"`
	IsEnabled         bool                                `json:"isEnabled"`
	IsType            string                              `json:"isType"`
	Configurations    []*PaymentMethodConfigurationDomain `json:"configurations,omitempty"`
	PaymentSubAccount *PaymentSubAccountDomain            `json:"subAccount,omitempty"`
	Name              string                              `json:"name"`
}

type PayFastDomain struct {
	MerchantID    string  `json:"merchantId"`
	MPaymentID    string  `json:"mPaymentId"`
	PfPaymentID   string  `json:"pfPaymentId"`
	PaymentStatus string  `json:"paymentStatus"`
	ItemName      string  `json:"itemName"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
}

func (d *PaymentMethodDomain) String() string {
	marshal, _ := json.Marshal(d)
	return string(marshal)
}

func (gateway *PaymentGatewayDomain) String() string {
	marshal, _ := json.Marshal(gateway)
	return string(marshal)
}

func (d *PaymentMethodDomain) GetRequiredConfiguration(key string) (*PaymentMethodConfigurationDomain, error) {
	for _, config := range d.Configurations {
		if config.Key == key {
			return config, nil
		}
	}

	return nil, fmt.Errorf("failed to get required configuration with given key: %s not found", key)
}

type GatewayConfigurationDomain struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type CountryDomain struct {
	ID             uint      `json:"id,omitempty"`
	Name           string    `json:"name,omitempty"`
	Cca2           string    `json:"cca2,omitempty"`
	Cca3           string    `json:"cca3,omitempty"`
	Ccn3           string    `json:"ccn3,omitempty"`
	ContinentID    uint      `json:"-"`
	SubContinentID uint      `json:"-"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}
