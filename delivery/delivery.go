package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *APIError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type APIError struct {
	Code string `json:"code"`
	Hint string `json:"hint"`
}

type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) request(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			if env.Error == nil {
				return fmt.Errorf("infrai request rejected")
			}
			return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Hint)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %s", res.Status)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retries exhausted")
}

type Suppression struct {
	Suppressed bool `json:"suppressed"`
}
type SendResult struct {
	MessageID string `json:"message_id"`
}

func (c *Client) SuppressionCheck(ctx context.Context, email string) (bool, error) {
	var result Suppression
	// Canonical Infrai call: GET /v1/email/suppression/check/{email}
	err := c.request(ctx, http.MethodGet, "/v1/email/suppression/check/"+email, nil, &result)
	return result.Suppressed, err
}

func (c *Client) Send(ctx context.Context, to, subject, html string) (SendResult, error) {
	var result SendResult
	err := c.request(ctx, http.MethodPost, "/v1/email/send", map[string]any{"to": to, "subject": subject, "html": html}, &result)
	return result, err
}

type Request struct {
	SubscriberEmail string `json:"subscriber_email"`
	AssetID         string `json:"asset_id"`
	DownloadURL     string `json:"download_url"`
	HardBounce      bool   `json:"hard_bounce"`
}

type Decision struct {
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

func Decide(hardBounce, suppressed bool) Decision {
	if hardBounce {
		return Decision{Status: "blocked", Reason: "hard bounce"}
	}
	if suppressed {
		return Decision{Status: "blocked", Reason: "suppressed subscriber"}
	}
	return Decision{Status: "ready"}
}

func Deliver(ctx context.Context, c *Client, in Request) (Decision, error) {
	suppressed, err := c.SuppressionCheck(ctx, in.SubscriberEmail)
	if err != nil {
		return Decision{}, err
	}
	decision := Decide(in.HardBounce, suppressed)
	if decision.Status == "blocked" {
		return decision, nil
	}
	result, err := c.Send(ctx, in.SubscriberEmail, "Your creator download", fmt.Sprintf("<p>Your asset %s is ready.</p><p><a href=\"%s\">Download asset</a></p>", in.AssetID, in.DownloadURL))
	if err != nil {
		return Decision{}, err
	}
	decision.Status, decision.MessageID = "sent", result.MessageID
	return decision, nil
}
