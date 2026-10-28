package quickbooks

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayment(t *testing.T) {
	byteValue, err := os.ReadFile("data/testing/payment.json")
	require.NoError(t, err)

	var r struct {
		Payment Payment
		Time    Date
	}
	require.NoError(t, json.Unmarshal(byteValue, &r))
	p := r.Payment
	assert.Equal(t, "152", p.Id)
	assert.Equal(t, "20", p.CustomerRef.Value)
	assert.Equal(t, "Check", p.PaymentMethodRef.Name)
	assert.Equal(t, "10264", p.PaymentRefNum)
	assert.Equal(t, "Wire for 340 and 341", p.PrivateNote)
	assert.Equal(t, 10000.0, p.TotalAmt)
	assert.Equal(t, "2026-10-20", p.TxnDate.Format("2006-01-02"))
	require.Len(t, p.Line, 2)
	assert.Equal(t, 7500.0, p.Line[0].Amount)
	require.Len(t, p.Line[0].LinkedTxn, 1)
	assert.Equal(t, "340", p.Line[0].LinkedTxn[0].TxnID)
	assert.Equal(t, "Invoice", p.Line[0].LinkedTxn[0].TxnType)
	assert.Equal(t, "341", p.Line[1].LinkedTxn[0].TxnID)
}
