package cli

import (
"encoding/json"
"errors"
"fmt"
"os"
"path/filepath"
"time"
)

const defaultTokenFileName = "session.json"

var ErrSessionNotFound = errors.New("no saved session found")

type StoredSession struct {
APIKey      string    `json:"api_key"`
AccessToken string    `json:"access_token"`
UserID      string    `json:"user_id,omitempty"`
UserName    string    `json:"user_name,omitempty"`
LoginAt     time.Time `json:"login_at"`
LastUsedAt  time.Time `json:"last_used_at"`
}

func DefaultTokenPath() (string, error) {
if explicit := os.Getenv("KITE_TOKEN_PATH"); explicit != "" {
return explicit, nil
}

home, err := os.UserHomeDir()
if err != nil {
return "", fmt.Errorf("failed to resolve home directory: %w", err)
}

return filepath.Join(home, ".kite-mcp", defaultTokenFileName), nil
}

func SaveSession(path string, session *StoredSession) error {
if session == nil {
return errors.New("session is nil")
}
if session.APIKey == "" {
return errors.New("api key is required")
}
if session.AccessToken == "" {
return errors.New("access token is required")
}

dir := filepath.Dir(path)
if err := os.MkdirAll(dir, 0o700); err != nil {
return fmt.Errorf("failed to create session directory: %w", err)
}

payload, err := json.MarshalIndent(session, "", "  ")
if err != nil {
return fmt.Errorf("failed to encode session: %w", err)
}

tmpFile := path + ".tmp"
if err := os.WriteFile(tmpFile, payload, 0o600); err != nil {
return fmt.Errorf("failed to write session temp file: %w", err)
}

if err := os.Rename(tmpFile, path); err != nil {
_ = os.Remove(tmpFile)
return fmt.Errorf("failed to finalize session file: %w", err)
}

return nil
}

func LoadSession(path string) (*StoredSession, error) {
payload, err := os.ReadFile(path)
if err != nil {
if errors.Is(err, os.ErrNotExist) {
return nil, ErrSessionNotFound
}
return nil, fmt.Errorf("failed to read session file: %w", err)
}

var session StoredSession
if err := json.Unmarshal(payload, &session); err != nil {
return nil, fmt.Errorf("failed to decode session file: %w", err)
}

if session.APIKey == "" || session.AccessToken == "" {
return nil, errors.New("session file is invalid")
}

return &session, nil
}

func DeleteSession(path string) error {
if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
return fmt.Errorf("failed to delete session file: %w", err)
}
return nil
}
