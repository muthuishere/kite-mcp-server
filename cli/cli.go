package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 1
	}

	switch args[0] {
	case "login":
		return runLogin(args[1:], stdin, stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "logout":
		return runLogout(args[1:], stdout, stderr)
	case "profile":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetUserProfile() })
	case "margins":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetUserMargins() })
	case "holdings":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetHoldings() })
	case "positions":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetPositions() })
	case "orders":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetOrders() })
	case "trades":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetTrades() })
	case "gtts":
		return runAuthedRead(args[1:], stderr, func(client *kiteconnect.Client) (any, error) { return client.GetGTTs() })
	case "order-history":
		return runOrderScoped(args[1:], stderr, func(client *kiteconnect.Client, orderID string) (any, error) { return client.GetOrderHistory(orderID) })
	case "order-trades":
		return runOrderScoped(args[1:], stderr, func(client *kiteconnect.Client, orderID string) (any, error) { return client.GetOrderTrades(orderID) })
	case "quotes":
		return runInstrumentsScoped(args[1:], stderr, func(client *kiteconnect.Client, instruments []string) (any, error) {
			return client.GetQuote(instruments...)
		})
	case "ltp":
		return runInstrumentsScoped(args[1:], stderr, func(client *kiteconnect.Client, instruments []string) (any, error) {
			return client.GetLTP(instruments...)
		})
	case "ohlc":
		return runInstrumentsScoped(args[1:], stderr, func(client *kiteconnect.Client, instruments []string) (any, error) {
			return client.GetOHLC(instruments...)
		})
	case "historical-data":
		return runHistoricalData(args[1:], stderr)
	default:
		fmt.Fprintf(stderr, "unknown cli command: %s\n\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: kite-mcp-server cli <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Auth commands:")
	fmt.Fprintln(w, "  login            Start auth flow and save access token")
	fmt.Fprintln(w, "  status           Show saved token status")
	fmt.Fprintln(w, "  logout           Delete saved token")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Read commands:")
	fmt.Fprintln(w, "  profile          Get user profile")
	fmt.Fprintln(w, "  margins          Get account margins")
	fmt.Fprintln(w, "  holdings         Get holdings")
	fmt.Fprintln(w, "  positions        Get positions")
	fmt.Fprintln(w, "  orders           Get orders")
	fmt.Fprintln(w, "  trades           Get trades")
	fmt.Fprintln(w, "  gtts             Get GTT orders")
	fmt.Fprintln(w, "  order-history    Get order history (--order-id)")
	fmt.Fprintln(w, "  order-trades     Get order trades (--order-id)")
	fmt.Fprintln(w, "  quotes           Get quotes (--instruments)")
	fmt.Fprintln(w, "  ltp              Get LTP (--instruments)")
	fmt.Fprintln(w, "  ohlc             Get OHLC (--instruments)")
	fmt.Fprintln(w, "  historical-data  Get historical data")
}

func runLogin(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	path, err := DefaultTokenPath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(stderr)
	apiKey := fs.String("api-key", os.Getenv("KITE_API_KEY"), "Kite API key")
	apiSecret := fs.String("api-secret", os.Getenv("KITE_API_SECRET"), "Kite API secret")
	tokenPath := fs.String("token-path", path, "Path to save CLI session token")
	requestToken := fs.String("request-token", "", "Kite request token (or callback URL containing request_token)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *apiKey == "" || *apiSecret == "" {
		fmt.Fprintln(stderr, "error: both --api-key and --api-secret are required (or set KITE_API_KEY/KITE_API_SECRET)")
		return 1
	}

	client := kiteconnect.New(*apiKey)
	token := strings.TrimSpace(*requestToken)
	if token == "" {
		fmt.Fprintf(stdout, "Open this URL in your browser and authorize:\n%s\n\n", client.GetLoginURL())
		fmt.Fprint(stdout, "Paste request token or callback URL: ")
		reader := bufio.NewReader(stdin)
		line, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			fmt.Fprintf(stderr, "error: failed to read request token: %v\n", readErr)
			return 1
		}
		token = strings.TrimSpace(line)
	}

	token = extractRequestToken(token)
	if token == "" {
		fmt.Fprintln(stderr, "error: request token is empty")
		return 1
	}

	userSession, err := client.GenerateSession(token, *apiSecret)
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to generate session: %v\n", err)
		return 1
	}

	client.SetAccessToken(userSession.AccessToken)
	now := time.Now().UTC()
	session := &StoredSession{
		APIKey:      *apiKey,
		AccessToken: userSession.AccessToken,
		UserID:      userSession.UserID,
		UserName:    userSession.UserName,
		LoginAt:     now,
		LastUsedAt:  now,
	}

	if err := SaveSession(*tokenPath, session); err != nil {
		fmt.Fprintf(stderr, "error: failed to save token: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Login successful for %s (%s). Token saved to %s\n", userSession.UserName, userSession.UserID, *tokenPath)
	return 0
}

func runStatus(args []string, stdout, stderr io.Writer) int {
	path, err := DefaultTokenPath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tokenPath := fs.String("token-path", path, "Path to saved CLI session token")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	session, err := LoadSession(*tokenPath)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			fmt.Fprintf(stdout, "No saved session at %s\n", *tokenPath)
			return 0
		}
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	out := map[string]any{
		"token_path":   *tokenPath,
		"api_key":      session.APIKey,
		"user_id":      session.UserID,
		"user_name":    session.UserName,
		"login_at":     session.LoginAt,
		"last_used_at": session.LastUsedAt,
	}
	if err := printJSON(stdout, out); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

func runLogout(args []string, stdout, stderr io.Writer) int {
	path, err := DefaultTokenPath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fs := flag.NewFlagSet("logout", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tokenPath := fs.String("token-path", path, "Path to saved CLI session token")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if err := DeleteSession(*tokenPath); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Logged out. Removed token file at %s\n", *tokenPath)
	return 0
}

func runAuthedRead(args []string, stderr io.Writer, fn func(client *kiteconnect.Client) (any, error)) int {
	client, _, _, err := newAuthedClient(args, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	result, err := fn(client)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if err := printJSON(os.Stdout, result); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

func runOrderScoped(args []string, stderr io.Writer, fn func(client *kiteconnect.Client, orderID string) (any, error)) int {
	path, err := DefaultTokenPath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fs := flag.NewFlagSet("order", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tokenPath := fs.String("token-path", path, "Path to saved CLI session token")
	apiKey := fs.String("api-key", "", "Kite API key override")
	orderID := fs.String("order-id", "", "Order ID")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if strings.TrimSpace(*orderID) == "" {
		fmt.Fprintln(stderr, "error: --order-id is required")
		return 1
	}

	client, err := authedClientFromSession(*tokenPath, *apiKey)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	result, err := fn(client, strings.TrimSpace(*orderID))
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if err := printJSON(os.Stdout, result); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func runInstrumentsScoped(args []string, stderr io.Writer, fn func(client *kiteconnect.Client, instruments []string) (any, error)) int {
	path, err := DefaultTokenPath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fs := flag.NewFlagSet("instruments", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tokenPath := fs.String("token-path", path, "Path to saved CLI session token")
	apiKey := fs.String("api-key", "", "Kite API key override")
	instrumentCSV := fs.String("instruments", "", "Comma-separated instruments (e.g. NSE:INFY,NSE:SBIN)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	instruments := parseCSV(*instrumentCSV)
	if len(instruments) == 0 {
		fmt.Fprintln(stderr, "error: --instruments is required")
		return 1
	}

	client, err := authedClientFromSession(*tokenPath, *apiKey)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	result, err := fn(client, instruments)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if err := printJSON(os.Stdout, result); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func runHistoricalData(args []string, stderr io.Writer) int {
	path, err := DefaultTokenPath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fs := flag.NewFlagSet("historical-data", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tokenPath := fs.String("token-path", path, "Path to saved CLI session token")
	apiKey := fs.String("api-key", "", "Kite API key override")
	instrumentToken := fs.Int("instrument-token", 0, "Instrument token")
	interval := fs.String("interval", "day", "Candle interval")
	from := fs.String("from", "", "From datetime (YYYY-MM-DD HH:MM:SS)")
	to := fs.String("to", "", "To datetime (YYYY-MM-DD HH:MM:SS)")
	continuous := fs.Bool("continuous", false, "Include continuous contracts")
	oi := fs.Bool("oi", false, "Include open interest")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *instrumentToken == 0 || strings.TrimSpace(*from) == "" || strings.TrimSpace(*to) == "" {
		fmt.Fprintln(stderr, "error: --instrument-token, --from and --to are required")
		return 1
	}

	fromTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(*from))
	if err != nil {
		fmt.Fprintf(stderr, "error: invalid --from format: %v\n", err)
		return 1
	}
	toTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(*to))
	if err != nil {
		fmt.Fprintf(stderr, "error: invalid --to format: %v\n", err)
		return 1
	}

	client, err := authedClientFromSession(*tokenPath, *apiKey)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	result, err := client.GetHistoricalData(*instrumentToken, strings.TrimSpace(*interval), fromTime, toTime, *continuous, *oi)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if err := printJSON(os.Stdout, result); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func newAuthedClient(args []string, stderr io.Writer) (*kiteconnect.Client, *StoredSession, string, error) {
	path, err := DefaultTokenPath()
	if err != nil {
		return nil, nil, "", err
	}

	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tokenPath := fs.String("token-path", path, "Path to saved CLI session token")
	apiKeyOverride := fs.String("api-key", "", "Kite API key override")
	if err := fs.Parse(args); err != nil {
		return nil, nil, "", err
	}

	session, err := LoadSession(*tokenPath)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, nil, "", fmt.Errorf("no saved token found. run 'kite-mcp-server cli login' first")
		}
		return nil, nil, "", err
	}

	apiKey := session.APIKey
	if strings.TrimSpace(*apiKeyOverride) != "" {
		apiKey = strings.TrimSpace(*apiKeyOverride)
	}

	client := kiteconnect.New(apiKey)
	client.SetAccessToken(session.AccessToken)

	session.LastUsedAt = time.Now().UTC()
	if saveErr := SaveSession(*tokenPath, session); saveErr != nil {
		return nil, nil, "", saveErr
	}

	return client, session, *tokenPath, nil
}

func authedClientFromSession(tokenPath, apiKeyOverride string) (*kiteconnect.Client, error) {
	session, err := LoadSession(tokenPath)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, fmt.Errorf("no saved token found. run 'kite-mcp-server cli login' first")
		}
		return nil, err
	}

	apiKey := session.APIKey
	if strings.TrimSpace(apiKeyOverride) != "" {
		apiKey = strings.TrimSpace(apiKeyOverride)
	}

	client := kiteconnect.New(apiKey)
	client.SetAccessToken(session.AccessToken)
	session.LastUsedAt = time.Now().UTC()
	if err := SaveSession(tokenPath, session); err != nil {
		return nil, err
	}
	return client, nil
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func extractRequestToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "request_token=") {
		u, err := url.Parse(trimmed)
		if err == nil {
			if token := strings.TrimSpace(u.Query().Get("request_token")); token != "" {
				return token
			}
		}

		segments := strings.Split(trimmed, "request_token=")
		if len(segments) > 1 {
			raw := segments[len(segments)-1]
			if idx := strings.Index(raw, "&"); idx >= 0 {
				raw = raw[:idx]
			}
			return strings.TrimSpace(raw)
		}
	}
	return trimmed
}
