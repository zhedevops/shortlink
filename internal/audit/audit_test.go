package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileSink_Consume(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "audit.json")

	sink := NewFileSink(path)

	event := AuditEvent{
		UserID: "123",
		Action: ActionShorten,
		URL:    "http://test.ru/test",
	}

	sink.Consume(event)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var parsed AuditEvent
	err = json.Unmarshal(bytes.TrimSpace(data), &parsed)
	require.NoError(t, err)

	require.Equal(t, event.UserID, parsed.UserID)
}

func TestRemoteSink_Consume(t *testing.T) {
	var received AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_ = r.Body.Close()
		}()
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sink := NewRemoteSink(server.URL)

	event := AuditEvent{
		UserID: "1",
		Action: ActionFollow,
		URL:    "http://test.ru/abc",
	}

	sink.Consume(event)

	require.Equal(t, event.UserID, received.UserID)
}
