package quickbooks

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Purchase represents a QuickBooks Online Purchase: an expense paid by cash,
// check or credit card. Lines marked BillableStatus "Billable" with a
// CustomerRef are the billable expenses that can be added to that customer's
// invoice.
type Purchase struct {
	Id            string        `json:"Id,omitempty"`
	SyncToken     string        `json:",omitempty"`
	MetaData      MetaData      `json:",omitempty"`
	DocNumber     string        `json:",omitempty"`
	TxnDate       Date          `json:",omitempty"`
	PaymentType   string        `json:",omitempty"` // Cash | Check | CreditCard
	AccountRef    ReferenceType `json:",omitempty"` // the bank / credit card account paid from
	EntityRef     ReferenceType `json:",omitempty"` // payee (vendor, customer or employee)
	DepartmentRef ReferenceType `json:",omitempty"`
	PrivateNote   string        `json:",omitempty"`
	Credit        bool          `json:",omitempty"`
	TotalAmt      json.Number   `json:",omitempty"`
	CurrencyRef   ReferenceType `json:",omitempty"`
	ExchangeRate  json.Number   `json:",omitempty"`
	LinkedTxn     []LinkedTxn   `json:",omitempty"`
	Line          []Line        `json:",omitempty"`
	Domain        string        `json:"domain,omitempty"`
}

// Billable statuses of an expense line.
const (
	BillableStatusBilled      = "Billed"
	BillableStatusBillable    = "Billable"
	BillableStatusNotBillable = "NotBillable"
)

// CreatePurchase creates the given Purchase on the QuickBooks server.
func (c *Client) CreatePurchase(purchase *Purchase) (*Purchase, error) {
	var resp struct {
		Purchase Purchase
		Time     Date
	}
	if err := c.post("purchase", purchase, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Purchase, nil
}

// FindPurchases gets the full list of Purchases in the QuickBooks account.
func (c *Client) FindPurchases() ([]Purchase, error) {
	return QueryAll[Purchase](c, "SELECT * FROM Purchase", 1000)
}

// FindPurchaseById returns a purchase with a given Id.
func (c *Client) FindPurchaseById(id string) (*Purchase, error) {
	var r struct {
		Purchase Purchase
		Time     Date
	}
	if err := c.get("purchase/"+id, &r, nil); err != nil {
		return nil, err
	}
	return &r.Purchase, nil
}

// QueryPurchases accepts an SQL query and returns all purchases found using it.
func (c *Client) QueryPurchases(query string) ([]Purchase, error) {
	return QueryAll[Purchase](c, query, 1000)
}

// UpdatePurchase performs a sparse update of an existing purchase, fetching the
// current SyncToken first.
func (c *Client) UpdatePurchase(purchase *Purchase) (*Purchase, error) {
	if purchase.Id == "" {
		return nil, errors.New("missing purchase id")
	}
	existing, err := c.FindPurchaseById(purchase.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing purchase: %v", err)
	}
	purchase.SyncToken = existing.SyncToken
	payload := struct {
		*Purchase
		Sparse bool `json:"sparse"`
	}{Purchase: purchase, Sparse: true}
	var resp struct {
		Purchase Purchase
		Time     Date
	}
	if err := c.post("purchase", payload, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Purchase, nil
}
