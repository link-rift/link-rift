package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRespondSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"message": "hello"}
	RespondSuccess(c, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Error != nil {
		t.Error("expected error to be nil")
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := NotFound("link")
	RespondError(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil {
		t.Fatal("expected error body")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestRespondErrorValidation(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := Validation("email", "invalid email format")
	RespondError(c, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

func TestRespondErrorGeneric(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, errors.New("something broke"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRespondList(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	items := []string{"a", "b", "c"}
	RespondList(c, items, 100, 10, 0)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta")
	}
	if resp.Meta.Total != 100 {
		t.Errorf("expected total=100, got %d", resp.Meta.Total)
	}
	if resp.Meta.Limit != 10 {
		t.Errorf("expected limit=10, got %d", resp.Meta.Limit)
	}
	if resp.Meta.Offset != 0 {
		t.Errorf("expected offset=0, got %d", resp.Meta.Offset)
	}
}

func TestMapToHTTPStatus(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{NotFound("x"), http.StatusNotFound},
		{AlreadyExists("x"), http.StatusConflict},
		{Unauthorized("x"), http.StatusUnauthorized},
		{Forbidden("x"), http.StatusForbidden},
		{Validation("x", "y"), http.StatusBadRequest},
		{RateLimited(), http.StatusTooManyRequests},
		{Wrap(errors.New("x"), "y"), http.StatusInternalServerError},
		{errors.New("generic"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		got := MapToHTTPStatus(tt.err)
		if got != tt.status {
			t.Errorf("MapToHTTPStatus(%v) = %d, want %d", tt.err, got, tt.status)
		}
	}
}
