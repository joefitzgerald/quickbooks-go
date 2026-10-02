package quickbooks

import "strconv"

// Class represents a QuickBooks Online Class used to categorize transactions and
// transaction lines (e.g. departments, locations, funds).
type Class struct {
	Id                 string        `json:"Id,omitempty"`
	Name               string        `json:",omitempty"`
	SyncToken          string        `json:",omitempty"`
	FullyQualifiedName string        `json:",omitempty"`
	Active             bool          `json:",omitempty"`
	SubClass           bool          `json:",omitempty"`
	ParentRef          ReferenceType `json:",omitempty"`
	Domain             string        `json:"domain,omitempty"`
	MetaData           MetaData      `json:",omitempty"`
}

// CreateClass creates the given Class. The company must have class tracking on.
func (c *Client) CreateClass(class *Class) (*Class, error) {
	var resp struct {
		Class Class
		Time  Date
	}
	if err := c.post("class", class, &resp, nil); err != nil {
		return nil, err
	}
	return &resp.Class, nil
}

// FindClasses gets the full list of Classes in the QuickBooks account. It
// returns an empty slice (not an error) when the company has no classes or
// class tracking is otherwise empty.
func (c *Client) FindClasses() ([]Class, error) {
	var resp struct {
		QueryResponse struct {
			Classes       []Class `json:"Class"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM Class", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return []Class{}, nil
	}

	classes := make([]Class, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += queryPageSize {
		query := "SELECT * FROM Class ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(queryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		classes = append(classes, resp.QueryResponse.Classes...)
	}

	return classes, nil
}

// QueryClasses accepts an SQL query and returns all classes found using it,
// or an empty slice when none match.
func (c *Client) QueryClasses(query string) ([]Class, error) {
	var resp struct {
		QueryResponse struct {
			Classes       []Class `json:"Class"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Classes, nil
}

// FindClassById returns a class with the given Id.
func (c *Client) FindClassById(id string) (*Class, error) {
	var resp struct {
		Class Class
		Time  Date
	}

	if err := c.get("class/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Class, nil
}
