package quickbooks

import "encoding/json"

// ReimburseCharge is a billable expense: QuickBooks creates one for each line of a
// Purchase or Bill marked Billable to a customer. It is read-only through the API.
//
// The charge does not name its source; the source line names the charge in its
// LinkedTxn (TxnType ReimburseCharge). Amount is unsigned: the signed amount (a
// refund is negative) is the sum of the lines. HasBeenInvoiced
// turns true, and LinkedTxn names the invoice, when an invoice lists the charge
// in its own LinkedTxn.
type ReimburseCharge struct {
	Id              string        `json:"Id,omitempty"`
	SyncToken       string        `json:",omitempty"`
	MetaData        MetaData      `json:",omitempty"`
	TxnDate         Date          `json:",omitempty"`
	CustomerRef     ReferenceType `json:",omitempty"`
	HasBeenInvoiced bool
	Amount          json.Number   `json:",omitempty"`
	PrivateNote     string        `json:",omitempty"`
	CurrencyRef     ReferenceType `json:",omitempty"`
	LinkedTxn       []LinkedTxn   `json:",omitempty"`
	Line            []Line        `json:",omitempty"`
	Domain          string        `json:"domain,omitempty"`
}

// InvoiceID is the id of the invoice that billed the charge, or "".
func (r ReimburseCharge) InvoiceID() string {
	for _, lt := range r.LinkedTxn {
		if lt.TxnType == "Invoice" {
			return lt.TxnID
		}
	}
	return ""
}

// FindReimburseChargeById reads one charge.
func (c *Client) FindReimburseChargeById(id string) (*ReimburseCharge, error) {
	var resp struct {
		ReimburseCharge ReimburseCharge
		Time            Date
	}
	if err := c.get("reimbursecharge/"+id, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.ReimburseCharge, nil
}
