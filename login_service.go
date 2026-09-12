package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PaymentEvent struct {
	ID          string    `json:"id"`
	AmountCents int64     `json:"amount_cents"`
	Status      string    `json:"status"`
	At          time.Time `json:"at"`
}
type AuditNotification struct {
	Kind  string    `json:"kind"`
	Phone string    `json:"phone"`
	At    time.Time `json:"at"`
}
type LoginService struct {
	client  *InfraiClient
	mu      sync.Mutex
	events  []PaymentEvent
	notices []AuditNotification
}

func (s *LoginService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		Phone       string `json:"phone"`
		Code        string `json:"code"`
		PaymentID   string `json:"payment_id"`
		AmountCents int64  `json:"amount_cents"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	switch r.URL.Path {
	case "/login/request":
		if strings.TrimSpace(in.Phone) == "" {
			http.Error(w, "phone is required", http.StatusBadRequest)
			return
		}
		if err := s.client.RequestOTP(in.Phone); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		s.recordNotice("otp_requested", in.Phone)
		writeJSON(w, map[string]string{"status": "code_sent"})
	case "/login/verify":
		if strings.TrimSpace(in.Phone) == "" {
			http.Error(w, "phone is required", http.StatusBadRequest)
			return
		}
		if err := s.client.VerifyOTP(in.Phone, in.Code); err != nil {
			http.Error(w, "code rejected", http.StatusUnauthorized)
			return
		}
		s.recordNotice("login_verified", in.Phone)
		writeJSON(w, map[string]string{"status": "verified"})
	case "/payments/events":
		if in.PaymentID == "" || in.AmountCents <= 0 {
			http.Error(w, "payment_id and amount_cents are required", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.events = append(s.events, PaymentEvent{ID: in.PaymentID, AmountCents: in.AmountCents, Status: "authorized", At: time.Now().UTC()})
		s.mu.Unlock()
		writeJSON(w, map[string]string{"status": "authorized"})
	default:
		http.NotFound(w, r)
	}
}
func (s *LoginService) recordNotice(kind, phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notices = append(s.notices, AuditNotification{Kind: kind, Phone: phone, At: time.Now().UTC()})
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
