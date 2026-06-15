package cli

import "testing"

func TestExtractRequestToken(t *testing.T) {
t.Parallel()

tests := []struct {
name  string
input string
want  string
}{
{name: "raw token", input: "abc123", want: "abc123"},
{name: "callback url", input: "http://localhost/callback?request_token=abc123&status=success", want: "abc123"},
{name: "callback string", input: "request_token=abc123&status=success", want: "abc123"},
{name: "empty", input: "", want: ""},
}

for _, tt := range tests {
tt := tt
t.Run(tt.name, func(t *testing.T) {
t.Parallel()
if got := extractRequestToken(tt.input); got != tt.want {
t.Fatalf("extractRequestToken(%q) = %q, want %q", tt.input, got, tt.want)
}
})
}
}

func TestParseCSV(t *testing.T) {
t.Parallel()
out := parseCSV("NSE:INFY, NSE:SBIN, ,NSE:TCS")
if len(out) != 3 {
t.Fatalf("expected 3 values, got %d", len(out))
}
}
