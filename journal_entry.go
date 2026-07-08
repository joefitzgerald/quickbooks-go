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

// FindJournalEntryByDocNumber returns the journal entry with the given
// DocNumber, or nil if none exists.
func (c *Client) FindJournalEntryByDocNumber(docNumber string) (*JournalEntry, error) {
	entries, err := c.QueryJournalEntries("SELECT * FROM JournalEntry WHERE DocNumber = '" + EscapeSQLString(docNumber) + "'")
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return &entries[0], nil
}

// QueryJournalEntriesByDocNumberPrefix returns all journal entries whose
// DocNumber begins with the given prefix.
func (c *Client) QueryJournalEntriesByDocNumberPrefix(prefix string) ([]JournalEntry, error) {
	return c.QueryJournalEntries("SELECT * FROM JournalEntry WHERE DocNumber LIKE '" + EscapeSQLString(prefix) + "%'")
}

// UpsertJournalEntryByDocNumber creates the entry, or updates the existing entry
// that shares its DocNumber, making posting idempotent by DocNumber. It returns
// the resulting entry and whether it was created (true) or updated (false).
// entry.DocNumber must be set.
func (c *Client) UpsertJournalEntryByDocNumber(entry *JournalEntry) (*JournalEntry, bool, error) {
	if entry.DocNumber == "" {
		return nil, false, errors.New("missing doc number")
	}

	existing, err := c.FindJournalEntryByDocNumber(entry.DocNumber)
	if err != nil {
		return nil, false, err
	}

	if existing == nil {
		created, err := c.CreateJournalEntry(entry)
		return created, true, err
	}

	entry.Id = existing.Id
	updated, err := c.UpdateJournalEntry(entry)
	return updated, false, err
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
