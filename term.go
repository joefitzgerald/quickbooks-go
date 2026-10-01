package quickbooks

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Term represents a QuickBooks Online sales term ("Net 30", "Due on receipt").
// Type is "STANDARD" (due a number of days after the transaction date) or
// "DATE_DRIVEN" (due on a day of the month).
type Term struct {
	Id                 string      `json:"Id,omitempty"`
	SyncToken          string      `json:",omitempty"`
	MetaData           MetaData    `json:",omitempty"`
	Name               string      `json:",omitempty"`
	Active             bool        `json:",omitempty"`
	Type               string      `json:",omitempty"`
	DueDays            int         `json:",omitempty"`
	DiscountDays       int         `json:",omitempty"`
	DiscountPercent    json.Number `json:",omitempty"`
	DayOfMonthDue      int         `json:",omitempty"`
	DueNextMonthDays   int         `json:",omitempty"`
	DiscountDayOfMonth int         `json:",omitempty"`
	Domain             string      `json:"domain,omitempty"`
}

// TermTypeStandard is a term due a number of days after the transaction date.
const TermTypeStandard = "STANDARD"

// TermTypeDateDriven is a term due on a day of the month.
const TermTypeDateDriven = "DATE_DRIVEN"

// CreateTerm creates the given Term on the QuickBooks server, returning the
// resulting Term object.
func (c *Client) CreateTerm(term *Term) (*Term, error) {
	var resp struct {
		Term Term
		Time Date
	}
	if err := c.post("term", term, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Term, nil
}

// FindTerms gets the full list of Terms in the QuickBooks account. It returns
// an empty slice (not an error) when the company has none.
func (c *Client) FindTerms() ([]Term, error) {
	return QueryAll[Term](c, "SELECT * FROM Term", 1000)
}

// FindTermById returns a term with a given Id.
func (c *Client) FindTermById(id string) (*Term, error) {
	var r struct {
		Term Term
		Time Date
	}
	if err := c.get("term/"+id, &r, nil); err != nil {
		return nil, err
	}
	return &r.Term, nil
}

// FindTermByName returns the term with the given name, or nil when none has it.
func (c *Client) FindTermByName(name string) (*Term, error) {
	terms, err := Query[Term](c, "SELECT * FROM Term WHERE Name = '"+EscapeSQLString(name)+"'")
	if err != nil {
		return nil, err
	}
	if len(terms) == 0 {
		return nil, nil
	}
	return &terms[0], nil
}

// QueryTerms accepts an SQL query and returns all terms found using it.
func (c *Client) QueryTerms(query string) ([]Term, error) {
	return QueryAll[Term](c, query, 1000)
}

// UpdateTerm performs a sparse update of an existing term. It fetches the
// current SyncToken so callers only need to supply Id plus the fields they
// wish to change.
func (c *Client) UpdateTerm(term *Term) (*Term, error) {
	if term.Id == "" {
		return nil, errors.New("missing term id")
	}
	existing, err := c.FindTermById(term.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing term: %v", err)
	}
	term.SyncToken = existing.SyncToken
	payload := struct {
		*Term
		Sparse bool `json:"sparse"`
	}{Term: term, Sparse: true}
	var resp struct {
		Term Term
		Time Date
	}
	if err := c.post("term", payload, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Term, nil
}
