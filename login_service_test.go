package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPaymentDecision(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		status int
		result string
	}{
		{"authorizes positive payment", `{"payment_id":"p-1","amount_cents":2500}`, 200, "authorized"},
		{"rejects missing amount", `{"phone":"+15550001","payment_id":"p-2","amount_cents":0}`, 400, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &LoginService{client: &InfraiClient{}}
			r := httptest.NewRequest("POST", "/payments/events", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status=%d want %d", w.Code, tc.status)
			}
			if tc.result != "" && !strings.Contains(w.Body.String(), tc.result) {
				t.Fatalf("body=%s", w.Body.String())
			}
		})
	}
}
