package cli

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadDeleteSession(t *testing.T) {
	t.Parallel()

	tokenPath := filepath.Join(t.TempDir(), "state", "session.json")
	now := time.Now().UTC().Round(time.Second)
	expected := &StoredSession{
		APIKey:      "api-key",
		AccessToken: "access-token",
		UserID:      "AB1234",
		UserName:    "Test User",
		LoginAt:     now,
		LastUsedAt:  now,
	}

	if err := SaveSession(tokenPath, expected); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := LoadSession(tokenPath)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	if loaded.APIKey != expected.APIKey || loaded.AccessToken != expected.AccessToken || loaded.UserID != expected.UserID {
		t.Fatalf("loaded session mismatch: got %+v want %+v", loaded, expected)
	}

	if err := DeleteSession(tokenPath); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, err = LoadSession(tokenPath)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound after delete, got %v", err)
	}
}
