package quickbooks

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournalEntry(t *testing.T) {
	jsonFile, err := os.Open("data/testing/journalentry.json")
	require.NoError(t, err)
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	require.NoError(t, err)

	var r struct {
		JournalEntry JournalEntry
		Time         Date
	}
	err = json.Unmarshal(byteValue, &r)
	require.NoError(t, err)

	je := r.JournalEntry
	assert.Equal(t, "147", je.Id)
	assert.Equal(t, "0", je.SyncToken)
	assert.Equal(t, "BA2026-01", je.DocNumber)
	assert.Equal(t, "Bonus accrual - January 2026", je.PrivateNote)
	assert.True(t, je.Adjustment)
	require.Len(t, je.Line, 2)

	debit := je.Line[0]
	assert.Equal(t, "JournalEntryLineDetail", debit.DetailType)
	assert.Equal(t, json.Number("134212.58"), debit.Amount)
	assert.Equal(t, PostingTypeDebit, debit.JournalEntryLineDetail.PostingType)
	assert.Equal(t, "92", debit.JournalEntryLineDetail.AccountRef.Value)
	assert.Equal(t, "6560 Bonus Expense", debit.JournalEntryLineDetail.AccountRef.Name)
	assert.Equal(t, "Accrual Adjustment", debit.JournalEntryLineDetail.ClassRef.Name)

	credit := je.Line[1]
	assert.Equal(t, PostingTypeCredit, credit.JournalEntryLineDetail.PostingType)
	assert.Equal(t, "93", credit.JournalEntryLineDetail.AccountRef.Value)
}

// TestJournalEntryMarshalRoundTrip ensures a JournalEntry built in Go marshals
// to the shape QBO expects (posting type + account ref per line).
func TestJournalEntryMarshalRoundTrip(t *testing.T) {
	entry := &JournalEntry{
		DocNumber:   "BA2026-01",
		PrivateNote: "Bonus accrual - January 2026",
		Line: []JournalEntryLine{
			{
				Description: "Bonus accrual - January 2026",
				Amount:      json.Number("134212.58"),
				DetailType:  "JournalEntryLineDetail",
				JournalEntryLineDetail: JournalEntryLineDetail{
					PostingType: PostingTypeDebit,
					AccountRef:  ReferenceType{Value: "92"},
					ClassRef:    ReferenceType{Value: "500000000000123"},
				},
			},
			{
				Description: "Bonus accrual - January 2026",
				Amount:      json.Number("134212.58"),
				DetailType:  "JournalEntryLineDetail",
				JournalEntryLineDetail: JournalEntryLineDetail{
					PostingType: PostingTypeCredit,
					AccountRef:  ReferenceType{Value: "93"},
					ClassRef:    ReferenceType{Value: "500000000000123"},
				},
			},
		},
	}

	b, err := json.Marshal(entry)
	require.NoError(t, err)

	// Decode into a generic map to assert the wire shape QBO expects without
	// depending on the lib's Date (un)marshal quirks.
	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, "BA2026-01", got["DocNumber"])

	lines, ok := got["Line"].([]interface{})
	require.True(t, ok)
	require.Len(t, lines, 2)

	debit := lines[0].(map[string]interface{})
	assert.Equal(t, "JournalEntryLineDetail", debit["DetailType"])
	detail := debit["JournalEntryLineDetail"].(map[string]interface{})
	assert.Equal(t, PostingTypeDebit, detail["PostingType"])
	assert.Equal(t, "92", detail["AccountRef"].(map[string]interface{})["value"])

	credit := lines[1].(map[string]interface{})
	creditDetail := credit["JournalEntryLineDetail"].(map[string]interface{})
	assert.Equal(t, PostingTypeCredit, creditDetail["PostingType"])
	assert.Equal(t, "93", creditDetail["AccountRef"].(map[string]interface{})["value"])
}
