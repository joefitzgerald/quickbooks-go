package quickbooks

import "testing"

func TestEscapeSQLString(t *testing.T) {
	cases := map[string]string{
		"BA2026-01":  "BA2026-01",
		"O'Brien":    `O\'Brien`,
		`back\slash`: `back\\slash`,
		`both'\`:     `both\'\\`,
		"":           "",
	}
	for in, want := range cases {
		if got := EscapeSQLString(in); got != want {
			t.Errorf("EscapeSQLString(%q) = %q, want %q", in, got, want)
		}
	}
}
