package quickbooks

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// QueryPaged is like Query but adds the ability to specify a start position and page size ot the query.
func QueryPaged[T any](c *Client, query string, startpos int, pagesize int) ([]T, error) {
	selectStatement := fmt.Sprintf("%s STARTPOSITION %d MAXRESULTS %d", query, startpos, pagesize)
	return Query[T](c, selectStatement)

}

// Query allows you to query any QuickBooks Online entity and have the result unmarshalled into
// a slice of a type you specify.
//
// Example:
//
//	type JournalEntry struct {
//		 Id     string
//		 Amount float64
//	}
//	result, err := quickbooks.Query[JournalEntry](client, "SELECT * FROM JournalEntry ORDERBY Id")
func Query[T any](c *Client, query string) ([]T, error) {

	var resp struct {
		QueryResponse map[string]json.RawMessage
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	for key, value := range resp.QueryResponse {
		switch key {
		case "startPosition", "maxResults", "totalCount": // skip these...
		default:
			var data []T
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.UseNumber()
			err := decoder.Decode(&data)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	return []T{}, nil
}

// QueryAll retrieves all pages of results for a query by automatically handling pagination.
// It will continue fetching pages until all results are retrieved.
// The pageSize parameter controls how many results are fetched per request (max 1000 per QuickBooks API).
//
// Example:
//
//	customers, err := quickbooks.QueryAll[Customer](client, "SELECT * FROM Customer", 100)
func QueryAll[T any](c *Client, query string, pageSize int) ([]T, error) {
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 1000 // QuickBooks API maximum
	}

	var allResults []T
	startPos := 1 // QuickBooks uses 1-based indexing

	for {
		// Fetch a page of results
		page, err := QueryPaged[T](c, query, startPos, pageSize)
		if err != nil {
			return allResults, err
		}

		// Add results to our collection
		allResults = append(allResults, page...)

		// Check if we got less than a full page (indicates last page)
		if len(page) < pageSize {
			break
		}

		// Move to next page
		startPos += pageSize
	}

	return allResults, nil
}
