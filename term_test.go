package quickbooks

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerm(t *testing.T) {
	byteValue, err := os.ReadFile("data/testing/term.json")
	require.NoError(t, err)

	var r struct {
		Term Term
		Time Date
	}
	require.NoError(t, json.Unmarshal(byteValue, &r))
	assert.Equal(t, "3", r.Term.Id)
	assert.Equal(t, "Net 30", r.Term.Name)
	assert.Equal(t, TermTypeStandard, r.Term.Type)
	assert.Equal(t, 30, r.Term.DueDays)
	assert.True(t, r.Term.Active)
}

func TestPurchase(t *testing.T) {
	byteValue, err := os.ReadFile("data/testing/purchase.json")
	require.NoError(t, err)

	var r struct {
		Purchase Purchase
		Time     Date
	}
	require.NoError(t, json.Unmarshal(byteValue, &r))
	p := r.Purchase
	assert.Equal(t, "252", p.Id)
	assert.Equal(t, "CreditCard", p.PaymentType)
	assert.Equal(t, "41", p.AccountRef.Value)
	assert.Equal(t, "Vendor", p.EntityRef.Type)
	require.Len(t, p.Line, 1)
	d := p.Line[0].AccountBasedExpenseLineDetail
	assert.Equal(t, BillableStatusBillable, d.BillableStatus)
	assert.Equal(t, "158", d.CustomerRef.Value)
	assert.Equal(t, "76", d.AccountRef.Value)
	assert.Equal(t, "2026-09-14", p.TxnDate.Format("2006-01-02"))
}

func TestPreferences(t *testing.T) {
	byteValue, err := os.ReadFile("data/testing/preferences.json")
	require.NoError(t, err)

	var r struct {
		Preferences Preferences
		Time        Date
	}
	require.NoError(t, json.Unmarshal(byteValue, &r))
	p := r.Preferences
	assert.Equal(t, "2025-12-31", p.AccountingInfoPrefs.BookCloseDate)
	assert.True(t, p.AccountingInfoPrefs.UseAccountNumbers)
	assert.Equal(t, "Clients", p.AccountingInfoPrefs.CustomerTerminology)
	assert.Equal(t, "3", p.SalesFormsPrefs.DefaultTerms.Value)
	require.Len(t, p.SalesFormsPrefs.CustomField, 1)
	names := map[string]string{}
	enabled := 0
	for _, f := range p.SalesFormsPrefs.CustomField[0].CustomField {
		if f.StringValue != "" {
			names[f.Name] = f.StringValue
		}
		if f.BooleanValue {
			enabled++
		}
	}
	assert.Equal(t, 2, enabled)
	assert.Equal(t, "PO Number", names["SalesFormsPrefs.SalesCustomName1"])
	assert.Equal(t, "Clarity ID", names["SalesFormsPrefs.SalesCustomName2"])
	assert.Equal(t, "USD", p.CurrencyPrefs.HomeCurrency.Value)
}

func TestCustomerTermAndCustomFieldsRoundTrip(t *testing.T) {
	c := Customer{Id: "9", DisplayName: "VMware LLC", SalesTermRef: ReferenceType{Value: "3"},
		CustomField: []CustomField{{DefinitionId: "1", Name: "PO Number", Type: "StringType", StringValue: "PO-123"}}}
	b, err := json.Marshal(c)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"SalesTermRef":{"value":"3"}`)
	assert.Contains(t, string(b), `"StringValue":"PO-123"`)
	// Zero Date fields do not round-trip through the library's Date type, so read
	// back only the fields under test.
	var back struct {
		SalesTermRef ReferenceType
		CustomField  []CustomField
	}
	require.NoError(t, json.Unmarshal(b, &back))
	assert.Equal(t, "3", back.SalesTermRef.Value)
	assert.Equal(t, "PO-123", back.CustomField[0].StringValue)
}

func TestPreferencesSparseUpdatePayload(t *testing.T) {
	on, label := true, "Clients"
	patch := PreferencesPatch{AccountingInfoPrefs: &AccountingInfoPatch{ClassTrackingPerTxn: &on, ClassTrackingPerTxnLine: &on, CustomerTerminology: &label}}
	b, err := json.Marshal(preferencesUpdate{Id: "1", SyncToken: "7", PreferencesPatch: patch, Sparse: true})
	require.NoError(t, err)
	// Only what the patch names is sent: nothing else can be changed by accident.
	assert.JSONEq(t, `{"Id":"1","SyncToken":"7","sparse":true,"AccountingInfoPrefs":{"ClassTrackingPerTxn":true,"ClassTrackingPerTxnLine":true,"CustomerTerminology":"Clients"}}`, string(b))
	assert.True(t, PreferencesPatch{}.Empty())
	assert.False(t, patch.Empty())
}
