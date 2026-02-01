package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Get_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer test-token")
		}

		json.NewEncoder(w).Encode(APIResponse{
			Success: true,
			Data:    json.RawMessage(`{"id":"123"}`),
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !resp.Success {
		t.Error("Response.Success = false, want true")
	}
}

func TestClient_Post_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("body[name] = %q, want %q", body["name"], "test")
		}

		json.NewEncoder(w).Encode(APIResponse{
			Success: true,
			Data:    json.RawMessage(`{"created":true}`),
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	resp, err := client.Post("/create", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}

	if !resp.Success {
		t.Error("Response.Success = false, want true")
	}
}

func TestClient_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "VALIDATION_ERROR",
				Message: "email is required",
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	_, err := client.Post("/login", map[string]string{})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}

	expected := "API error: email is required"
	if err.Error() != expected {
		t.Errorf("Error = %q, want %q", err.Error(), expected)
	}
}

func TestClient_NoToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Authorization should be empty, got %q", r.Header.Get("Authorization"))
		}
		json.NewEncoder(w).Encode(APIResponse{Success: true})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.Get("/public")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %s, want DELETE", r.Method)
		}
		json.NewEncoder(w).Encode(APIResponse{Success: true})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	_, err := client.Delete("/item/123")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestClient_ResponseWithMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(APIResponse{
			Success: true,
			Data:    json.RawMessage(`[{"id":"1"},{"id":"2"}]`),
			Meta:    &APIMeta{Total: 50, Limit: 20, Offset: 0},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	resp, err := client.Get("/items")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if resp.Meta == nil {
		t.Fatal("Meta is nil, want non-nil")
	}
	if resp.Meta.Total != 50 {
		t.Errorf("Meta.Total = %d, want 50", resp.Meta.Total)
	}
}

func TestNewClientFromConfig_NoToken(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Save config without token
	cfg := &CLIConfig{APIBaseURL: "http://localhost:8080"}
	SaveConfig(cfg)

	_, err := NewClientFromConfig()
	if err == nil {
		t.Fatal("Expected error when no token")
	}
}
