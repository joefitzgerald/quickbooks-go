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
	VendorAndPurchases  VendorAndPurchasesPrefs `json:"VendorAndPurchasesPrefs,omitempty"`
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

// VendorAndPurchasesPrefs holds the expense settings: tracking expenses by customer and
// making them billable, without which QuickBooks makes no billable charges.
type VendorAndPurchasesPrefs struct {
	TrackingByCustomer      bool `json:",omitempty"`
	BillableExpenseTracking bool `json:",omitempty"`
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
	AccountingInfoPrefs     *AccountingInfoPatch     `json:",omitempty"`
	SalesFormsPrefs         *SalesFormsPatch         `json:",omitempty"`
	ProductAndServicesPrefs *ProductAndServicesPatch `json:",omitempty"`
	VendorAndPurchasesPrefs *VendorAndPurchasesPatch `json:",omitempty"`
}

// ProductAndServicesPatch is the changeable part of ProductAndServicesPrefs.
type ProductAndServicesPatch struct {
	ForSales                 *bool `json:",omitempty"`
	ForPurchase              *bool `json:",omitempty"`
	QuantityWithPriceAndRate *bool `json:",omitempty"`
	QuantityOnHand           *bool `json:",omitempty"`
}

// VendorAndPurchasesPatch is the changeable part of VendorAndPurchasesPrefs.
type VendorAndPurchasesPatch struct {
	TrackingByCustomer      *bool `json:",omitempty"`
	BillableExpenseTracking *bool `json:",omitempty"`
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
	AllowDeposit     *bool `json:",omitempty"`
	AllowDiscount    *bool `json:",omitempty"`
	AllowEstimates   *bool `json:",omitempty"`
	AllowServiceDate *bool `json:",omitempty"`
	AllowShipping    *bool `json:",omitempty"`
}

// Empty reports whether the patch changes nothing.
func (p PreferencesPatch) Empty() bool {
	return p.AccountingInfoPrefs == nil && p.SalesFormsPrefs == nil && p.ProductAndServicesPrefs == nil && p.VendorAndPurchasesPrefs == nil
}

// Complete fills every setting the patch leaves unset with its current value, in every
// section modelled here. QuickBooks' "sparse" Preferences update turns off settings of
// the sections a request leaves out (a sandbox lost custom transaction numbers, item
// quantities and billable expense tracking that way, 2026-10-04), so UpdatePreferences
// always sends them all.
func (p PreferencesPatch) Complete(cur *Preferences) PreferencesPatch {
	b := func(dst **bool, v bool) {
		if *dst == nil {
			*dst = &v
		}
	}
	out := p
	ai := AccountingInfoPatch{}
	if p.AccountingInfoPrefs != nil {
		ai = *p.AccountingInfoPrefs
	}
	b(&ai.UseAccountNumbers, cur.AccountingInfoPrefs.UseAccountNumbers)
	b(&ai.ClassTrackingPerTxn, cur.AccountingInfoPrefs.ClassTrackingPerTxn)
	b(&ai.ClassTrackingPerTxnLine, cur.AccountingInfoPrefs.ClassTrackingPerTxnLine)
	b(&ai.TrackDepartments, cur.AccountingInfoPrefs.TrackDepartments)
	if ai.CustomerTerminology == nil && cur.AccountingInfoPrefs.CustomerTerminology != "" {
		t := cur.AccountingInfoPrefs.CustomerTerminology
		ai.CustomerTerminology = &t
	}
	out.AccountingInfoPrefs = &ai
	sf := SalesFormsPatch{}
	if p.SalesFormsPrefs != nil {
		sf = *p.SalesFormsPrefs
	}
	b(&sf.CustomTxnNumbers, cur.SalesFormsPrefs.CustomTxnNumbers)
	b(&sf.AllowDeposit, cur.SalesFormsPrefs.AllowDeposit)
	b(&sf.AllowDiscount, cur.SalesFormsPrefs.AllowDiscount)
	b(&sf.AllowEstimates, cur.SalesFormsPrefs.AllowEstimates)
	b(&sf.AllowServiceDate, cur.SalesFormsPrefs.AllowServiceDate)
	b(&sf.AllowShipping, cur.SalesFormsPrefs.AllowShipping)
	out.SalesFormsPrefs = &sf
	ps := ProductAndServicesPatch{}
	if p.ProductAndServicesPrefs != nil {
		ps = *p.ProductAndServicesPrefs
	}
	b(&ps.ForSales, cur.ProductAndServices.ForSales)
	b(&ps.ForPurchase, cur.ProductAndServices.ForPurchase)
	b(&ps.QuantityWithPriceAndRate, cur.ProductAndServices.QuantityWithPriceAndRate)
	b(&ps.QuantityOnHand, cur.ProductAndServices.QuantityOnHand)
	out.ProductAndServicesPrefs = &ps
	vp := VendorAndPurchasesPatch{}
	if p.VendorAndPurchasesPrefs != nil {
		vp = *p.VendorAndPurchasesPrefs
	}
	b(&vp.TrackingByCustomer, cur.VendorAndPurchases.TrackingByCustomer)
	b(&vp.BillableExpenseTracking, cur.VendorAndPurchases.BillableExpenseTracking)
	out.VendorAndPurchasesPrefs = &vp
	return out
}

// preferencesUpdate is the sparse update payload.
type preferencesUpdate struct {
	Id        string
	SyncToken string
	PreferencesPatch
	Sparse bool `json:"sparse"`
}

// UpdatePreferences sparsely updates the company preferences: the settings named in
// the patch change, and every other setting modelled here is sent at its current value
// (see Complete). The current SyncToken is read first: QuickBooks
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
	payload := preferencesUpdate{Id: current.Id, SyncToken: current.SyncToken, PreferencesPatch: patch.Complete(current), Sparse: true}
	if err := c.post("preferences", payload, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Preferences, nil
}
