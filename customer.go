// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/guregu/null.v4"
)

// Customer represents a QuickBooks Customer object.
type Customer struct {
	Id                 string          `json:",omitempty"`
	SyncToken          string          `json:",omitempty"`
	MetaData           MetaData        `json:",omitempty"`
	CustomField        []CustomField   `json:",omitempty"`
	Title              string          `json:",omitempty"`
	GivenName          string          `json:",omitempty"`
	MiddleName         string          `json:",omitempty"`
	FamilyName         string          `json:",omitempty"`
	Suffix             string          `json:",omitempty"`
	DisplayName        string          `json:",omitempty"`
	FullyQualifiedName string          `json:",omitempty"`
	CompanyName        string          `json:",omitempty"`
	PrintOnCheckName   string          `json:",omitempty"`
	Active             bool            `json:",omitempty"`
	PrimaryPhone       TelephoneNumber `json:",omitempty"`
	AlternatePhone     TelephoneNumber `json:",omitempty"`
	Mobile             TelephoneNumber `json:",omitempty"`
	Fax                TelephoneNumber `json:",omitempty"`
	CustomerTypeRef    ReferenceType   `json:",omitempty"`
	PrimaryEmailAddr   *EmailAddress   `json:",omitempty"`
	WebAddr            *WebSiteAddress `json:",omitempty"`
	// DefaultTaxCodeRef
	Taxable              *bool            `json:",omitempty"`
	TaxExemptionReasonId *string          `json:",omitempty"`
	BillAddr             *PhysicalAddress `json:",omitempty"`
	ShipAddr             *PhysicalAddress `json:",omitempty"`
	Notes                string           `json:",omitempty"`
	Job                  null.Bool        `json:",omitempty"`
	BillWithParent       bool             `json:",omitempty"`
	ParentRef            ReferenceType    `json:",omitempty"`
	Level                int              `json:",omitempty"`
	SalesTermRef         ReferenceType    `json:",omitempty"` // default payment terms (a Term id)
	PaymentMethodRef     ReferenceType    `json:",omitempty"`
	Balance              json.Number      `json:",omitempty"`
	OpenBalanceDate      Date             `json:",omitempty"`
	BalanceWithJobs      json.Number      `json:",omitempty"`
	CurrencyRef          ReferenceType    `json:",omitempty"`
}

// GetAddress prioritizes the ship address, but falls back on bill address
func (c *Customer) GetAddress() PhysicalAddress {
	if c.ShipAddr != nil {
		return *c.ShipAddr
	}
	if c.BillAddr != nil {
		return *c.BillAddr
	}
	return PhysicalAddress{}
}

// GetWebsite de-nests the Website object
func (c *Customer) GetWebsite() string {
	if c.WebAddr != nil {
		return c.WebAddr.URI
	}
	return ""
}

// GetPrimaryEmail de-nests the PrimaryEmailAddr object
func (c *Customer) GetPrimaryEmail() string {
	if c.PrimaryEmailAddr != nil {
		return c.PrimaryEmailAddr.Address
	}
	return ""
}

// CreateCustomer creates the given Customer on the QuickBooks server,
// returning the resulting Customer object.
func (c *Client) CreateCustomer(customer *Customer) (*Customer, error) {
	var resp struct {
		Customer Customer
		Time     Date
	}

	if err := c.post("customer", customer, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Customer, nil
}

// FindCustomers gets the full list of Customers in the QuickBooks account.
func (c *Client) FindCustomers() ([]Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customers     []Customer `json:"Customer"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM Customer", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no customers could be found")
	}

	customers := make([]Customer, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += queryPageSize {
		query := "SELECT * FROM Customer ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(queryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Customers == nil {
			return nil, errors.New("no customers could be found")
		}

		customers = append(customers, resp.QueryResponse.Customers...)
	}

	return customers, nil
}

// FindCustomerById returns a customer with a given Id.
func (c *Client) FindCustomerById(id string) (*Customer, error) {
	var r struct {
		Customer Customer
		Time     Date
	}

	if err := c.get("customer/"+id, &r, nil); err != nil {
		return nil, err
	}

	return &r.Customer, nil
}

// FetchCustomerRaw returns the raw JSON response for a customer by ID.
func (c *Client) FetchCustomerRaw(id string) (string, error) {
	var raw json.RawMessage
	if err := c.get("customer/"+id, &raw, map[string]string{
		"include": "enhancedAllCustomFields",
	}); err != nil {
		return "", err
	}
	return string(raw), nil
}

// FindCustomerByIdWithCustomFields returns a customer with enhanced custom fields.
func (c *Client) FindCustomerByIdWithCustomFields(id string) (*Customer, error) {
	var r struct {
		Customer Customer
		Time     Date
	}

	if err := c.get("customer/"+id, &r, map[string]string{
		"include": "enhancedAllCustomFields",
	}); err != nil {
		return nil, err
	}

	return &r.Customer, nil
}

// FindCustomerByName gets a customer with a given name.
func (c *Client) FindCustomerByName(name string) (*Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customer   []Customer
			TotalCount int
		}
	}

	query := "SELECT * FROM Customer WHERE DisplayName = '" + strings.Replace(name, "'", "''", -1) + "'"

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if len(resp.QueryResponse.Customer) == 0 {
		return nil, errors.New("no customers could be found")
	}

	return &resp.QueryResponse.Customer[0], nil
}

// customFieldParams asks the API to include custom field values on customers
// (QuickBooks Online Advanced). Writes carry the same parameter so the
// CustomField array in the payload is applied.
var customFieldParams = map[string]string{"include": "enhancedAllCustomFields"}

// FindCustomersWithCustomFields gets every customer with their custom field
// values included.
func (c *Client) FindCustomersWithCustomFields() ([]Customer, error) {
	return QueryAllWithParams[Customer](c, "SELECT * FROM Customer", 1000, customFieldParams)
}

// UpdateCustomerWithCustomFields is UpdateCustomer with custom field values
// included in the request and the response, so a sparse update can set them.
func (c *Client) UpdateCustomerWithCustomFields(customer *Customer) (*Customer, error) {
	return c.updateCustomer(customer, customFieldParams)
}

// QueryCustomers accepts an SQL query and returns all customers found using it
func (c *Client) QueryCustomers(query string) ([]Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customers     []Customer `json:"Customer"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Customers, nil
}

// UpdateCustomer updates the given Customer on the QuickBooks server,
// returning the resulting Customer object. It's a sparse update, as not all QB
// fields are present in our Customer object.
//
// A sparse update leaves out false, so it cannot make a customer inactive: use
// DeactivateCustomer. It can make one active again (Active: true), in the same
// request as a new name, parent and details, with one catch: QuickBooks takes
// the last " (...)" of a DisplayName sent with the reactivation for its own
// " (deleted)" mark and drops it ("Bridge (Phase 2)" comes back as "Bridge"),
// so such a name has to be sent again once the customer is active. QuickBooks
// refuses every other change to an inactive customer.
func (c *Client) UpdateCustomer(customer *Customer) (*Customer, error) {
	return c.updateCustomer(customer, nil)
}

func (c *Client) updateCustomer(customer *Customer, params map[string]string) (*Customer, error) {
	if customer.Id == "" {
		return nil, errors.New("missing customer id")
	}

	existingCustomer, err := c.FindCustomerById(customer.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing customer: %v", err)
	}

	customer.SyncToken = existingCustomer.SyncToken

	payload := struct {
		*Customer
		Sparse bool `json:"sparse"`
	}{
		Customer: customer,
		Sparse:   true,
	}

	var customerData struct {
		Customer Customer
		Time     Date
	}

	if err = c.post("customer", payload, &customerData, params); err != nil {
		return nil, err
	}

	return &customerData.Customer, nil
}

// customerActiveUpdate is the sparse update that only changes Active. Customer's
// own Active field is left out of a request when false, so it cannot carry this.
type customerActiveUpdate struct {
	Id        string
	SyncToken string
	Active    bool
	Sparse    bool `json:"sparse"`
}

// DeactivateCustomer makes a customer inactive, which is how QuickBooks deletes
// one: a customer can never be removed. QuickBooks appends " (deleted)" to its
// display name, leaves it out of queries that do not ask for inactive customers
// (WHERE Active IN (true, false)), refuses any other change to it until it is
// active again, and writes off its open balance. A customer with active
// sub-customers cannot be made inactive.
func (c *Client) DeactivateCustomer(id string) (*Customer, error) {
	return c.setCustomerActive(id, false)
}

// ActivateCustomer makes an inactive customer active again. QuickBooks takes
// " (deleted)" off its display name, adding "-1" when another customer has taken
// the name meanwhile. A sub-customer cannot be made active while its parent is
// inactive. To reuse an inactive customer under a new name, parent or details,
// UpdateCustomer with Active set does it in one request (see there for how
// QuickBooks treats a new name ending in parentheses).
func (c *Client) ActivateCustomer(id string) (*Customer, error) {
	return c.setCustomerActive(id, true)
}

func (c *Client) setCustomerActive(id string, active bool) (*Customer, error) {
	if id == "" {
		return nil, errors.New("missing customer id")
	}

	existingCustomer, err := c.FindCustomerById(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing customer: %v", err)
	}

	var customerData struct {
		Customer Customer
		Time     Date
	}

	payload := customerActiveUpdate{Id: id, SyncToken: existingCustomer.SyncToken, Active: active, Sparse: true}
	if err = c.post("customer", payload, &customerData, nil); err != nil {
		return nil, err
	}

	return &customerData.Customer, nil
}
