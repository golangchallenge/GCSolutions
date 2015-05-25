package instagram

import (
	"bytes"
	"testing"
)

func Test_sig(t *testing.T) {
	// Example from https://instagram.com/developer/secure-api-requests/
	got := sig(
		"6dc1787668c64c939929c17683d7cb74",
		"/users/self",
		map[string][]string{
			"access_token": []string{"fb2e77d.47a0479900504cb3ab4a1f626d174d2d"},
		},
	)
	want := "cbf5a1f41db44412506cb6563a3218b50f45a710c7a8a65a3e9b18315bb338bf"
	if got != want {
		t.Fatalf("Sig() got %s, want %s", got, want)
	}
}

func Test_Client_Popular(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode: skipping live API calls")
	}
	c := NewClient()
	p, err := c.Popular()
	if err != nil {
		t.Fatalf("Popular() failed: %s", err)
	}
	urls := []string{}
	for _, m := range p.Media {
		if m.IsPhoto() {
			urls = append(urls, m.StandardImage().URL)
		}
	}
	if len(urls) == 0 {
		t.Fatalf("Want some urls, got none")
	}
	var buf bytes.Buffer
	b, err := buf.ReadFrom(p.Media[0].StandardImage())
	if err != nil {
		t.Fatalf("Error reading image: %s", err)
	}
	if b == 0 {
		t.Fatalf("Rep#Read() got 0 bytes, want some bytes")
	}
}
