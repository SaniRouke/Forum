// utils_test.go
package utils

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestErrorPage(t *testing.T) {
	rec := httptest.NewRecorder()
	ErrorPage(rec, http.StatusNotFound, "page not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rec.Code, http.StatusNotFound)
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

func TestNeuter1(t *testing.T) {

	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Neuter(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Neuter() = %v, want %v", got, tt.want)
			}
		})
	}
}
