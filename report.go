package quickbooks

import (
	"encoding/json"
	"fmt"
)

type ProfitAndLossReport struct {
	Header  ReportHeader  `json:"Header"`
	Columns ReportColumns `json:"Columns"`
	Rows    ReportRows    `json:"Rows"`
}

type ReportHeader struct {
	Time             string              `json:"Time"`
	ReportName       string              `json:"ReportName"`
	ReportBasis      string              `json:"ReportBasis"`
	StartPeriod      string              `json:"StartPeriod"`
	EndPeriod        string              `json:"EndPeriod"`
	SummarizeColumns string              `json:"SummarizeColumnsBy,omitempty"`
	Currency         string              `json:"Currency"`
	Options          []map[string]string `json:"Options,omitempty"`
}

type ReportColumns struct {
	Column []ReportColumn `json:"Column"`
}

type ReportColumn struct {
	ColTitle string           `json:"ColTitle,omitempty"`
	ColType  string           `json:"ColType,omitempty"`
	MetaData []ReportMetaData `json:"MetaData,omitempty"`
}

type ReportMetaData struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type ReportRows struct {
	Row []ReportRow `json:"Row"`
}

type ReportRow struct {
	ColData []ReportColData `json:"ColData,omitempty"`
	Summary *ReportSummary  `json:"Summary,omitempty"`
	Type    string          `json:"type,omitempty"`
	Group   string          `json:"group,omitempty"`
	Rows    *ReportRows     `json:"Rows,omitempty"`
}

type ReportColData struct {
	Value string `json:"value"`
	ID    string `json:"id,omitempty"`
}

type ReportSummary struct {
	ColData []ReportColData `json:"ColData"`
}

func (c *Client) GetProfitAndLossReport(startDate, endDate string, accountingMethod string) (*ProfitAndLossReport, error) {
	if accountingMethod == "" {
		accountingMethod = "Cash"
	}

	queryParams := map[string]string{
		"start_date":        startDate,
		"end_date":          endDate,
		"accounting_method": accountingMethod,
	}

	var report ProfitAndLossReport

	endpoint := "reports/ProfitAndLoss"
	if err := c.get(endpoint, &report, queryParams); err != nil {
		return nil, fmt.Errorf("failed to get profit and loss report: %w", err)
	}

	return &report, nil
}

func (r *ProfitAndLossReport) GetNetIncome() (float64, error) {
	for _, row := range r.Rows.Row {
		if row.Group == "NetIncome" && row.Summary != nil {
			for _, colData := range row.Summary.ColData {
				if colData.Value != "" && colData.Value != "Net Income" {
					var netIncome float64
					if err := json.Unmarshal([]byte(colData.Value), &netIncome); err != nil {
						return 0, fmt.Errorf("failed to parse net income value %s: %w", colData.Value, err)
					}
					return netIncome, nil
				}
			}
		}

		if row.Rows != nil {
			for _, subRow := range row.Rows.Row {
				if subRow.Group == "NetIncome" && subRow.Summary != nil {
					for _, colData := range subRow.Summary.ColData {
						if colData.Value != "" && colData.Value != "Net Income" {
							var netIncome float64
							if err := json.Unmarshal([]byte(colData.Value), &netIncome); err != nil {
								return 0, fmt.Errorf("failed to parse net income value %s: %w", colData.Value, err)
							}
							return netIncome, nil
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("net income not found in report")
}
