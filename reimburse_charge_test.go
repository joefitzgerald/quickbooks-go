package quickbooks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readFixture[T any](t *testing.T, file, key string) T {
	t.Helper()
	raw, err := os.ReadFile("data/testing/" + file)
	require.NoError(t, err)
	var wrapper map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &wrapper))
	var out T
	require.NoError(t, json.Unmarshal(wrapper[key], &out))
	return out
}

// Recorded from the sandbox: a charge billed on invoice 154, a refund charge (a
// negative bill line, whose Amount QuickBooks gives unsigned), and the purchase
// that made the billed charge.
func TestReimburseChargeFixtures(t *testing.T) {
	billed := readFixture[ReimburseCharge](t, "reimburse_charge.json", "ReimburseCharge")
	assert.Equal(t, "149", billed.Id)
	assert.True(t, billed.HasBeenInvoiced)
	assert.Equal(t, "154", billed.InvoiceID())
	assert.Equal(t, "70", billed.CustomerRef.Value)
	require.Len(t, billed.Line, 1)
	assert.Equal(t, DetailTypeReimburse, billed.Line[0].DetailType)
	require.NotNil(t, billed.Line[0].ReimburseLineDetail)
	assert.Equal(t, "125", billed.Line[0].ReimburseLineDetail.ItemAccountRef.Value)
	assert.Nil(t, billed.Line[0].SalesItemLineDetail)

	refund := readFixture[ReimburseCharge](t, "reimburse_charge_refund.json", "ReimburseCharge")
	assert.Equal(t, "64.99", refund.Amount.String())
	assert.Equal(t, "-64.99", refund.Line[0].Amount.String())
	assert.Equal(t, "", refund.InvoiceID())

	p := readFixture[Purchase](t, "purchase_billed.json", "Purchase")
	require.NotNil(t, p.Line[0].AccountBasedExpenseLineDetail)
	assert.Equal(t, BillableStatusHasBeenBilled, p.Line[0].AccountBasedExpenseLineDetail.BillableStatus)
	assert.Equal(t, []LinkedTxn{{TxnID: "149", TxnType: "ReimburseCharge"}}, p.Line[0].LinkedTxn)

	inv := readFixture[Invoice](t, "invoice_reimburse.json", "Invoice")
	assert.Contains(t, inv.LinkedTxn, LinkedTxn{TxnID: "149", TxnType: "ReimburseCharge"})
	var linked *Line
	for i := range inv.Line {
		if len(inv.Line[i].LinkedTxn) > 0 {
			linked = &inv.Line[i]
		}
	}
	require.NotNil(t, linked)
	assert.Equal(t, LinkedTxn{TxnID: "149", TxnType: "ReimburseCharge", TxnLineId: "1"}, linked.LinkedTxn[0])
	assert.Equal(t, "Purchase Order Number", inv.CustomField[0].Name)
}

// A line marshals only its own detail: an empty detail of another type is not sent.
func TestLineMarshalsOnlyItsDetail(t *testing.T) {
	l := Line{
		Amount:     "671.92",
		DetailType: DetailTypeSalesItem,
		LinkedTxn:  []LinkedTxn{{TxnID: "149", TxnType: "ReimburseCharge", TxnLineId: "1"}},
		SalesItemLineDetail: &SalesItemLineDetail{
			ItemRef:        ReferenceType{Value: "21"},
			ItemAccountRef: ReferenceType{Value: "125"},
			Qty:            "1",
			UnitPrice:      "671.92",
		},
	}
	raw, err := json.Marshal(l)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	assert.NotContains(t, m, "AccountBasedExpenseLineDetail")
	assert.NotContains(t, m, "DiscountLineDetail")
	assert.NotContains(t, m, "TaxLineDetail")
	assert.NotContains(t, m, "ReimburseLineDetail")
	assert.Equal(t, "1", m["LinkedTxn"].([]any)[0].(map[string]any)["TxnLineId"])
	assert.NotContains(t, m["SalesItemLineDetail"], "MarkupInfo")
}

type recorded struct {
	path  string
	query url.Values
	body  map[string]any
}

func invoiceServer(t *testing.T, got *[]recorded) *Client {
	t.Helper()
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/file" {
			w.Header().Set("Content-Type", "image/png")
			assert.Empty(t, r.Header.Get("Authorization"))
			_, _ = w.Write([]byte("receipt-bytes"))
			return
		}
		if r.URL.Path == "/v3/company/1/download/7" {
			_, _ = w.Write([]byte(srvURL + "/file\n"))
			return
		}
		rec := recorded{path: r.URL.Path, query: r.URL.Query()}
		if r.Method == http.MethodPost {
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(raw, &rec.body))
		}
		*got = append(*got, rec)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"Invoice": map[string]any{"Id": "154", "SyncToken": "3", "PrivateNote": "Voided"}})
	}))
	t.Cleanup(srv.Close)
	srvURL = srv.URL
	endpoint, err := url.Parse(srv.URL + "/v3/company/1/")
	require.NoError(t, err)
	return &Client{Client: srv.Client(), endpoint: endpoint, minorVersion: "75"}
}

func TestInvoiceWrites(t *testing.T) {
	var got []recorded
	c := invoiceServer(t, &got)

	_, err := c.CreateInvoiceWithOptions(&Invoice{DocNumber: "355", CustomerRef: ReferenceType{Value: "70"}}, InvoiceOptions{RequestID: "inv-1:0:abc", CustomFields: true})
	require.NoError(t, err)
	assert.Equal(t, "inv-1:0:abc", got[0].query.Get("requestid"))
	assert.Equal(t, "enhancedAllCustomFields", got[0].query.Get("include"))
	assert.Equal(t, "355", got[0].body["DocNumber"])

	_, err = c.ReplaceInvoice(&Invoice{Id: "154", SyncToken: "2", CustomerRef: ReferenceType{Value: "70"}, LinkedTxn: []LinkedTxn{}}, InvoiceOptions{})
	require.NoError(t, err)
	assert.Equal(t, false, got[1].body["sparse"])
	assert.Equal(t, "2", got[1].body["SyncToken"])
	assert.Equal(t, []any{}, got[1].body["LinkedTxn"])
	assert.Empty(t, got[1].query.Get("requestid"))

	_, err = c.ReplaceInvoice(&Invoice{Id: "154"}, InvoiceOptions{})
	assert.Error(t, err, "a full update needs the SyncToken it was read at")

	voided, err := c.VoidInvoiceWithSyncToken("154", "3")
	require.NoError(t, err)
	assert.Equal(t, "Voided", voided.PrivateNote)
	assert.Equal(t, "void", got[2].query.Get("operation"))
	assert.Equal(t, map[string]any{"Id": "154", "SyncToken": "3"}, got[2].body)

	_, err = c.FindInvoiceByIdWithOptions("154", InvoiceOptions{CustomFields: true})
	require.NoError(t, err)
	assert.Equal(t, "/v3/company/1/invoice/154", got[3].path)
	assert.Equal(t, "enhancedAllCustomFields", got[3].query.Get("include"))
}

func TestOpenAttachable(t *testing.T) {
	var got []recorded
	c := invoiceServer(t, &got)
	body, contentType, err := c.OpenAttachable("7")
	require.NoError(t, err)
	defer body.Close()
	b, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, "receipt-bytes", string(b))
	assert.Equal(t, "image/png", contentType)
}
