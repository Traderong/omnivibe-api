package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const brevoEmailEndpoint = "https://api.brevo.com/v3/smtp/email"

func sendBrevoEmail(to, subject, body, from, apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("Brevo API key is missing")
	}
	if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
		return fmt.Errorf("email sender or recipient is missing")
	}

	payload, err := json.Marshal(map[string]interface{}{
		"sender":      map[string]string{"email": from},
		"to":          []map[string]string{{"email": to}},
		"subject":     subject,
		"textContent": body,
	})
	if err != nil {
		return fmt.Errorf("Brevo request encoding failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		brevoEmailEndpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("Brevo request creation failed: %w", err)
	}
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Brevo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("Brevo returned HTTP %d: %s",
			resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	return nil
}
