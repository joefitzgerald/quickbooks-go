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

// PreferencesPatch names the settings a sparse update changes. Every field is a
// pointer: nil leaves the setting as it is, so only what is named is sent.
type PreferencesPatch struct {
	AccountingInfoPrefs *AccountingInfoPatch `json:",omitempty"`
	SalesFormsPrefs     *SalesFormsPatch     `json:",omitempty"`
}

// AccountingInfoPatch is the changeable part of AccountingInfoPrefs.
type AccountingInfoPatch struct {
	UseAccountNumbers       *bool   `json:",omitempty"`
	ClassTrackingPerTxn     *bool   `json:",omitempty"`
	ClassTrackingPerTxnLine *bool   `json:",omitempty"`
	TrackDepartments        *bool   `json:",omitempty"`
	CustomerTerminology     *string `json:",omitempty"`
}

// SalesFormsPatch is the changeable part of SalesFormsPrefs.
type SalesFormsPatch struct {
	CustomTxnNumbers *bool `json:",omitempty"`
}

// Empty reports whether the patch changes nothing.
func (p PreferencesPatch) Empty() bool {
	return p.AccountingInfoPrefs == nil && p.SalesFormsPrefs == nil
}

// preferencesUpdate is the sparse update payload.
type preferencesUpdate struct {
	Id        string
	SyncToken string
	PreferencesPatch
	Sparse bool `json:"sparse"`
}

// UpdatePreferences sparsely updates the company preferences: only the settings
// named in the patch are changed. The current SyncToken is read first: QuickBooks
// refuses a stale one, which is also why its own settings screen can fail on a
// sandbox company with "someone else was working on this at the same time".
func (c *Client) UpdatePreferences(patch PreferencesPatch) (*Preferences, error) {
	current, err := c.FindPreferences()
	if err != nil {
		return nil, err
	}
	var resp struct {
		Preferences Preferences
		Time        Date
	}
	payload := preferencesUpdate{Id: current.Id, SyncToken: current.SyncToken, PreferencesPatch: patch, Sparse: true}
	if err := c.post("preferences", payload, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Preferences, nil
}
