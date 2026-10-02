package quickbooks

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionListReport(t *testing.T) {
	body, err := os.ReadFile("data/testing/transaction_list_report.json")
	require.NoError(t, err)
	var asked url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/company/1/reports/TransactionList", r.URL.Path)
		asked = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		if asked.Get("customer") == "59" {
			// A customer nothing names: the report comes back without rows.
			_, _ = w.Write([]byte(`{"Header":{"ReportName":"TransactionList"},"Columns":{"Column":[]},"Rows":{}}`))
			return
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	endpoint, err := url.Parse(srv.URL + "/v3/company/1/")
	require.NoError(t, err)
	c := &Client{Client: srv.Client(), endpoint: endpoint, minorVersion: "75"}

	used, err := c.GetTransactionListReport("2000-01-01", "2099-12-31", map[string]string{"customer": "58"})
	require.NoError(t, err)
	assert.Equal(t, "58", asked.Get("customer"))
	assert.Equal(t, "2000-01-01", asked.Get("start_date"))
	assert.Equal(t, "2099-12-31", asked.Get("end_date"))
	assert.Equal(t, "TransactionList", used.Header.ReportName)
	assert.Equal(t, 2, used.TransactionCount(), "rows inside a section count; the section and its total do not")

	unused, err := c.GetTransactionListReport("2000-01-01", "2099-12-31", map[string]string{"customer": "59"})
	require.NoError(t, err)
	assert.Zero(t, unused.TransactionCount())
}
