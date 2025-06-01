package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
)

func TestCreateAPIKeyHandler(t *testing.T) {
	dataDir := t.TempDir()
	cfg := &config.Config{DataPath: dataDir, UploadPath: filepath.Join(dataDir, "upload")}
	db, err := store.NewDB(dataDir)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}

	as := store.NewAPIKeyService(db)
	rs := store.NewResourceService(cfg, db)
	restService := httpapi.NewRESTService(cfg, as, rs)

	err = rs.CreateUploadDir()
	if err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}

	// create initial trusted and untrusted API keys
	_, err = as.AddAPIKey("trusted", "trusted comment", true, nil)
	if err != nil {
		t.Fatalf("failed to add trusted API key: %v", err)
	}

	_, err = as.AddAPIKey("untrusted", "untrusted comment", false, nil)
	if err != nil {
		t.Fatalf("failed to add untrusted API key: %v", err)
	}

	tests := []struct {
		name           string
		authKey        string
		highlyTrusted  bool
		body           string
		expectedStatus int
	}{
		{
			name:           "trusted API key - valid request 1",
			authKey:        "trusted",
			highlyTrusted:  false,
			body:           `{"key":"new-key-1","comment":"test","highly_trusted":false}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "trusted API key - invalid JSON",
			authKey:        "trusted",
			body:           `not json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "untrusted API key",
			authKey:        "untrusted",
			highlyTrusted:  false,
			body:           `{"key":"new-key-2","comment":"test","highly_trusted":false}`,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing Authorization",
			authKey:        "",
			highlyTrusted:  false,
			body:           `{"key":"new-key-3","comment":"test","highly_trusted":false}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong Authorization",
			authKey:        "wrong",
			highlyTrusted:  false,
			body:           `{"key":"new-key-4","comment":"test","highly_trusted":false}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "trusted API key - reduced params 1",
			authKey:        "trusted",
			highlyTrusted:  false,
			body:           `{"key":"new-key-5","comment":"test"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "trusted API key - reduced params 2",
			authKey:        "trusted",
			highlyTrusted:  false,
			body:           `{"key":"new-key-6"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "trusted API key - valid request 2",
			authKey:        "trusted",
			highlyTrusted:  true,
			body:           `{"key":"new-key-7","comment":"test","highly_trusted":true}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "trusted API key - no comment",
			authKey:        "trusted",
			highlyTrusted:  false,
			body:           `{"key":"new-key-8","highly_trusted":false}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "trusted API key - empty JSON",
			authKey:        "trusted",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "trusted API key - invalid JSON 2",
			authKey:        "trusted",
			body:           `{"k":"new-key-9"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "trusted API key - invalid JSON 3",
			authKey:        "trusted",
			body:           `{"key":"new-key-10"`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/apikey", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			if tt.authKey != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authKey)
			}

			restService.CreateAPIKeyHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusCreated {
				// parse response
				var resp httpapi.APIKeyResponse

				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				if ok, _ := as.IsAPIKeyHighlyTrusted(resp.UUID); ok != tt.highlyTrusted {
					t.Fatalf("wrong trust level for key.")
				}
			}
		})
	}
}
