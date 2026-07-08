package quickbooks

import "strings"

// EscapeSQLString escapes a value for safe interpolation into a QuickBooks
// query string literal (the value between single quotes). Backslashes and
// single quotes are backslash-escaped per Intuit's query syntax.
func EscapeSQLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}
