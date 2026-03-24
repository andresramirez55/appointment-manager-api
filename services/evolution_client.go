package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// EvolutionWhatsAppClient implementa WhatsAppSender usando Evolution API
type EvolutionWhatsAppClient struct {
	apiURL       string
	apiKey       string
	instanceName string
	httpClient   *http.Client
}

func NewEvolutionWhatsAppClient(apiURL, apiKey, instanceName string) WhatsAppSender {
	return &EvolutionWhatsAppClient{
		apiURL:       apiURL,
		apiKey:       apiKey,
		instanceName: instanceName,
		httpClient:   &http.Client{},
	}
}

type evolutionMessageRequest struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	Options struct {
		Delay int `json:"delay"`
	} `json:"options"`
}

func (e *EvolutionWhatsAppClient) SendMessage(ctx context.Context, phone string, message string) error {
	endpoint := fmt.Sprintf("%s/message/sendText/%s", e.apiURL, url.PathEscape(e.instanceName))

	reqBody := evolutionMessageRequest{
		Number: phone,
		Text:   message,
	}
	reqBody.Options.Delay = 1200 // 1.2 segundos de delay

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", e.apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("evolution API returned status %d", resp.StatusCode)
	}

	return nil
}
