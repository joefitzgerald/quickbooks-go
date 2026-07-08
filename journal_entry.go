package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

// PostingType values for a JournalEntryLineDetail.
const (
	PostingTypeDebit  = "Debit"
	PostingTypeCredit = "Credit"
)

// JournalEntryLineDetail holds the account, posting type, and optional class
// for a single line of a JournalEntry.
type JournalEntryLineDetail struct {
	PostingType   string         `json:",omitempty"` // "Debit" | "Credit"
	AccountRef    ReferenceType  `json:",omitempty"` // Value = QBO account Id
	ClassRef      ReferenceType  `json:",omitempty"`
	DepartmentRef ReferenceType  `json:",omitempty"`
	Entity        *ReferenceType `json:",omitempty"`
	TaxCodeRef    ReferenceType  `json:",omitempty"`
}

// JournalEntryLine is a single debit or credit line of a JournalEntry.
type JournalEntryLine struct {
	Id                     string                 `json:"Id,omitempty"`
	Description            string                 `json:",omitempty"`
	Amount                 json.Number            `json:",omitempty"`
	DetailType             string                 `json:",omitempty"` // always "JournalEntryLineDetail"
	JournalEntryLineDetail JournalEntryLineDetail `json:",omitempty"`
}

// JournalEntry represents a QuickBooks general journal entry.
type JournalEntry struct {
	SyncToken   string             `json:",omitempty"`
	Domain      string             `json:"domain,omitempty"`
	TxnDate     Date               `json:",omitempty"`
	DocNumber   string             `json:",omitempty"`
	PrivateNote string             `json:",omitempty"`
	Adjustment  bool               `json:",omitempty"`
	Line        []JournalEntryLine `json:",omitempty"`
	Id          string             `json:",omitempty"`
	MetaData    MetaData           `json:",omitempty"`
}

// CreateJournalEntry creates the given journal entry within QuickBooks.
func (c *Client) CreateJournalEntry(entry *JournalEntry) (*JournalEntry, error) {
	var resp struct {
		JournalEntry JournalEntry
		Time         Date
	}

	if err := c.post("journalentry", entry, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.JournalEntry, nil
}

// DeleteJournalEntry deletes the given journal entry. Requires Id and SyncToken.
func (c *Client) DeleteJournalEntry(entry *JournalEntry) error {
	if entry.Id == "" || entry.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post("journalentry", entry, nil, map[string]string{"operation": "delete"})
}

// FindJournalEntryById returns a journal entry with a given Id.
func (c *Client) FindJournalEntryById(id string) (*JournalEntry, error) {
	var resp struct {
		JournalEntry JournalEntry
		Time         Date
	}

	if err := c.get("journalentry/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.JournalEntry, nil
}

// QueryJournalEntries accepts an SQL query and returns all journal entries found using it.
func (c *Client) QueryJournalEntries(query string) ([]JournalEntry, error) {
	var resp struct {
		QueryResponse struct {
			JournalEntries []JournalEntry `json:"JournalEntry"`
			StartPosition  int
			MaxResults     int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.JournalEntries, nil
}

// FindJournalEntries gets the full list of journal entries in the QuickBooks account.
func (c *Client) FindJournalEntries() ([]JournalEntry, error) {
	var resp struct {
		QueryResponse struct {
			JournalEntries []JournalEntry `json:"JournalEntry"`
			MaxResults     int
			StartPosition  int
			TotalCount     int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM JournalEntry", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no journal entries could be found")
	}

	entries := make([]JournalEntry, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += queryPageSize {
		query := "SELECT * FROM JournalEntry ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(queryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		entries = append(entries, resp.QueryResponse.JournalEntries...)
	}

	return entries, nil
}

// UpdateJournalEntry performs a sparse update of an existing journal entry.
// It fetches the current SyncToken so callers only need to supply Id plus the
// fields they wish to change.
func (c *Client) UpdateJournalEntry(entry *JournalEntry) (*JournalEntry, error) {
	if entry.Id == "" {
		return nil, errors.New("missing journal entry id")
	}

	existing, err := c.FindJournalEntryById(entry.Id)
	if err != nil {
		return nil, err
	}

	entry.SyncToken = existing.SyncToken

	payload := struct {
		*JournalEntry
		Sparse bool `json:"sparse"`
	}{
		JournalEntry: entry,
		Sparse:       true,
	}

	var resp struct {
		JournalEntry JournalEntry
		Time         Date
	}

	if err = c.post("journalentry", payload, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.JournalEntry, nil
}
