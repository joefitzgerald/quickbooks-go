package quickbooks

// Preferences is the company-wide preference singleton. Only the parts Flux
// and its sandbox tooling read are modelled; unknown sections are ignored.
type Preferences struct {
	Id                  string                  `json:"Id,omitempty"`
	SyncToken           string                  `json:",omitempty"`
	MetaData            MetaData                `json:",omitempty"`
	AccountingInfoPrefs AccountingInfoPrefs     `json:",omitempty"`
	SalesFormsPrefs     SalesFormsPrefs         `json:",omitempty"`
	ProductAndServices  ProductAndServicesPrefs `json:"ProductAndServicesPrefs,omitempty"`
	CurrencyPrefs       CurrencyPrefs           `json:",omitempty"`
	OtherPrefs          OtherPrefs              `json:",omitempty"`
}

// AccountingInfoPrefs holds the books-closing date and account-number settings.
type AccountingInfoPrefs struct {
	BookCloseDate             string `json:",omitempty"`
	UseAccountNumbers         bool   `json:",omitempty"`
	ClassTrackingPerTxn       bool   `json:",omitempty"`
	ClassTrackingPerTxnLine   bool   `json:",omitempty"`
	TrackDepartments          bool   `json:",omitempty"`
	DepartmentTerminology     string `json:",omitempty"`
	CustomerTerminology       string `json:",omitempty"`
	FirstMonthOfFiscalYear    string `json:",omitempty"`
	TaxYearMonth              string `json:",omitempty"`
	FirstMonthOfIncomeTaxYear string `json:",omitempty"`
}

// SalesFormsPrefs holds sales form defaults, including the custom fields
// printed on invoices (PO number, project ids and the like).
type SalesFormsPrefs struct {
	CustomField               []CustomFieldGroup `json:",omitempty"`
	CustomTxnNumbers          bool               `json:",omitempty"`
	AllowDeposit              bool               `json:",omitempty"`
	AllowDiscount             bool               `json:",omitempty"`
	AllowEstimates            bool               `json:",omitempty"`
	AllowServiceDate          bool               `json:",omitempty"`
	AllowShipping             bool               `json:",omitempty"`
	DefaultTerms              ReferenceType      `json:",omitempty"`
	DefaultCustomerMessage    string             `json:",omitempty"`
	EmailCopyToCompany        bool               `json:",omitempty"`
	ETransactionEnabledStatus string             `json:",omitempty"`
	IPNSupportEnabled         bool               `json:",omitempty"`
	UsingProgressInvoicing    bool               `json:",omitempty"`
	UsingPriceLevels          bool               `json:",omitempty"`
	DefaultDiscountAccount    string             `json:",omitempty"`
	DefaultShippingAccount    string             `json:",omitempty"`
}

// CustomFieldGroup is how Preferences reports custom field definitions: a group
// of name/value pairs such as SalesFormsPrefs.UseSalesCustom1 = true and
// SalesFormsPrefs.SalesCustomName1 = "PO Number".
type CustomFieldGroup struct {
	CustomField []CustomField `json:",omitempty"`
}

// ProductAndServicesPrefs holds inventory and service item settings.
type ProductAndServicesPrefs struct {
	ForSales                 bool `json:",omitempty"`
	ForPurchase              bool `json:",omitempty"`
	QuantityWithPriceAndRate bool `json:",omitempty"`
	QuantityOnHand           bool `json:",omitempty"`
}

// CurrencyPrefs holds the home currency and multicurrency flag.
type CurrencyPrefs struct {
	HomeCurrency         ReferenceType `json:",omitempty"`
	MultiCurrencyEnabled bool          `json:",omitempty"`
}

// OtherPrefs is the free-form name/value bag Intuit keeps the rest in.
type OtherPrefs struct {
	NameValue []NameValue `json:",omitempty"`
}

// NameValue is one entry of a name/value bag.
type NameValue struct {
	Name  string `json:",omitempty"`
	Value string `json:",omitempty"`
}

// FindPreferences returns the company preferences singleton.
func (c *Client) FindPreferences() (*Preferences, error) {
	prefs, err := Query[Preferences](c, "SELECT * FROM Preferences")
	if err != nil {
		return nil, err
	}
	if len(prefs) == 0 {
		return &Preferences{}, nil
	}
	return &prefs[0], nil
}
