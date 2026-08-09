package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/healthz", nil)

	Handler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
	}

	body := recorder.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}
