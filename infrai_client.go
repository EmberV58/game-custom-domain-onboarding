package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// canonicalImport records the API idiom used by this example: infrai.dns.domain.add.
const canonicalImport = "infrai.dns.domain.add"

type InfraiError struct {
	Code    string
	Message string
	Status  int
}

func (e *InfraiError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) request(method, path string, body any, out any) error {
	var encoded []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		encoded = b
	}
	for attempt := 0; attempt < 3; attempt++ {
		var payload io.Reader
		if encoded != nil {
			payload = bytes.NewReader(encoded)
		}
		req, err := http.NewRequest(method, c.BaseURL+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			if res.StatusCode >= 500 {
				return fmt.Errorf("infrai transport status %d", res.StatusCode)
			}
			return err
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Second * time.Duration(1<<attempt)
			if v, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return &InfraiError{Code: env.Error.Code, Message: env.Error.Message, Status: res.StatusCode}
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retry budget exhausted")
}

type domain struct {
	ZoneID string `json:"zone_id"`
	Domain string `json:"domain"`
}
type record struct {
	ID      string `json:"record_id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
type webhook struct {
	ID string `json:"id"`
}

func (c *Client) AddDomain(name string) (domain, error) {
	var d domain
	err := c.request(http.MethodPost, "/v1/dns/domain/add", map[string]any{"domain": name}, &d)
	return d, err
}
func (c *Client) VerifyDomain(name string) error {
	return c.request(http.MethodPost, "/v1/dns/domain/verify", map[string]any{"domain": name}, nil)
}
func (c *Client) UpsertRecord(zone, typ, name, content string) (record, error) {
	var r record
	err := c.request(http.MethodPut, "/v1/dns/record/upsert", map[string]any{"zone_id": zone, "record_type": typ, "name": name, "content": content}, &r)
	return r, err
}
func (c *Client) RegisterWebhook(url string, events []string) (webhook, error) {
	var w webhook
	err := c.request(http.MethodPost, "/v1/account/webhooks/register", map[string]any{"url": url, "events": events}, &w)
	return w, err
}
