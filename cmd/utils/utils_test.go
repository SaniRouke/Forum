// utils_test.go
package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestErrorPage(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	ErrorPage(rec, http.StatusNotFound, "page not found")

	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	expected := "page not found"
	if !strings.Contains(rec.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rec.Body.String(), expected)
	}
}

func TestNeuter(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := Neuter(nextHandler)

	tests := []struct {
		url          string
		expectedCode int
	}{
		{"/", http.StatusOK},
		{"/static/", http.StatusNotFound},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.url, nil)
		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)

		if status := rr.Code; status != test.expectedCode {
			t.Errorf("handler returned wrong status code: got %v want %v", status, test.expectedCode)
		}
	}
}
