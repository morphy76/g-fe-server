package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeadersSuite(t *testing.T) {
	t.Log("Test Headers Suite")

	t.Run("Test JSONResponse", func(t *testing.T) {
		t.Log("Test Headers JSONResponse")

		nextInvoked := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextInvoked = true
		})
		mid := JSONResponse(next)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		mid.ServeHTTP(w, r)

		if !nextInvoked {
			t.Fatalf("Expected next handler to be invoked")
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("Expected Content-Type: application/json, got %s", w.Header().Get("Content-Type"))
		}
	})
}
