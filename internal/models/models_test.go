package models

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing/wrong auth header: %q", r.Header.Get("Authorization"))
		}
		w.Write([]byte(`{"data":[{"id":"m1","context_length":131072},{"id":"m2"}]}`))
	}))
	defer srv.Close()

	got, err := Fetch(srv.URL, "tok")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 models, got %d", len(got))
	}
	if got[0].ID != "m1" || got[0].ContextLength != 131072 {
		t.Errorf("model 0 wrong: %+v", got[0])
	}
	if got[1].ContextLength != 0 {
		t.Errorf("model 1 should have 0 ctx, got %d", got[1].ContextLength)
	}
}

func TestFetchNoAuthHeaderWhenTokenEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("should not send auth header when token empty")
		}
		w.Write([]byte(`{"data":[{"id":"m1"}]}`))
	}))
	defer srv.Close()
	if _, err := Fetch(srv.URL, ""); err != nil {
		t.Fatalf("fetch: %v", err)
	}
}

func TestFetchBadJSONErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	if _, err := Fetch(srv.URL, ""); err == nil {
		t.Error("expected error on bad json")
	}
}

func TestFetchNon200Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()
	if _, err := Fetch(srv.URL, ""); err == nil {
		t.Error("expected error on 500")
	}
}
