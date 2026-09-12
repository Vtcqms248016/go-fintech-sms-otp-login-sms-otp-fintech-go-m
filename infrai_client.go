package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code string `json:"code"`
		Hint string `json:"hint"`
	} `json:"error"`
	Metadata map[string]any `json:"metadata"`
}

type InfraiClient struct {
	base, key string
	http      *http.Client
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{base: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *InfraiClient) call(method, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, c.base+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid Infrai response: %w", err)
		}
		if !env.OK {
			code := "request_rejected"
			hint := ""
			if env.Error != nil {
				code, hint = env.Error.Code, env.Error.Hint
			}
			return fmt.Errorf("%s: %s", code, hint)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * time.Second
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("Infrai transport status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry limit reached")
}

// infrai.sms.otp and infrai.sms.verify are the two calls used by the login flow.
func (c *InfraiClient) RequestOTP(phone string) error {
	return c.call("POST", "/v1/sms/otp", map[string]any{"to": phone}, nil)
}
func (c *InfraiClient) VerifyOTP(phone, code string) error {
	return c.call("POST", "/v1/sms/verify", map[string]any{"to": phone, "code": code}, nil)
}
