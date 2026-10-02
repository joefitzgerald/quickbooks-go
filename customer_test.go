package quickbooks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// customerServer is a QuickBooks company with one customer. It records the bodies
// posted to it and applies Active as QuickBooks does: an inactive customer's name
// carries " (deleted)".
func customerServer(t *testing.T, posted *[]map[string]any) *Client {
	t.Helper()
	answer := func(w http.ResponseWriter, active bool, syncToken string) {
		name := "Conduent Phase 2"
		if !active {
			name += " (deleted)"
		}
		// Active is spelled out: Customer leaves a false one out when marshalled.
		_ = json.NewEncoder(w).Encode(map[string]any{"Customer": map[string]any{"Id": "58", "SyncToken": syncToken, "DisplayName": name, "Active": active}})
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			assert.Equal(t, "/v3/company/1/customer/58", r.URL.Path)
			answer(w, true, "7")
			return
		}
		assert.Equal(t, "/v3/company/1/customer", r.URL.Path)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, json.Unmarshal(raw, &body))
		*posted = append(*posted, body)
		answer(w, body["Active"] == true, "8")
	}))
	t.Cleanup(srv.Close)
	endpoint, err := url.Parse(srv.URL + "/v3/company/1/")
	require.NoError(t, err)
	return &Client{Client: srv.Client(), endpoint: endpoint, minorVersion: "75"}
}

func TestDeactivateAndActivateCustomer(t *testing.T) {
	var posted []map[string]any
	c := customerServer(t, &posted)

	gone, err := c.DeactivateCustomer("58")
	require.NoError(t, err)
	assert.False(t, gone.Active)
	assert.Equal(t, "Conduent Phase 2 (deleted)", gone.DisplayName)
	back, err := c.ActivateCustomer("58")
	require.NoError(t, err)
	assert.True(t, back.Active)
	assert.Equal(t, "Conduent Phase 2", back.DisplayName)

	// Each request is sparse, carries the current SyncToken and Active spelled out,
	// false included, and nothing else that could change the customer.
	require.Len(t, posted, 2)
	assert.Equal(t, map[string]any{"Id": "58", "SyncToken": "7", "Active": false, "sparse": true}, posted[0])
	assert.Equal(t, map[string]any{"Id": "58", "SyncToken": "7", "Active": true, "sparse": true}, posted[1])

	_, err = c.DeactivateCustomer("")
	assert.Error(t, err)
	assert.Len(t, posted, 2, "no request without an id")
}

// An ordinary sparse update cannot deactivate (false is left out), and can
// reactivate while it renames: how an unused inactive customer is reused.
func TestUpdateCustomerActiveFlag(t *testing.T) {
	var posted []map[string]any
	c := customerServer(t, &posted)

	_, err := c.UpdateCustomer(&Customer{Id: "58", DisplayName: "Renamed"})
	require.NoError(t, err)
	_, err = c.UpdateCustomer(&Customer{Id: "58", DisplayName: "[PCTE-9] New Project", Active: true})
	require.NoError(t, err)

	require.Len(t, posted, 2)
	assert.NotContains(t, posted[0], "Active")
	assert.Equal(t, true, posted[1]["Active"])
	assert.Equal(t, "[PCTE-9] New Project", posted[1]["DisplayName"])
}
