package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wails_app/internal/config"
)

const (
	defaultModel = "gemini-1.5-flash"
	apiBaseURL   = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
)

type Manager struct {
	ctx    context.Context
	config *config.Configuration
}

func NewManager(ctx context.Context, cfg *config.Configuration) *Manager {
	return &Manager{
		ctx:    ctx,
		config: cfg,
	}
}

// Request structs for Gemini API
type GenerateContentRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// GenerateContent calls the Gemini API with a prompt and optional image data
func (m *Manager) GenerateContent(prompt string, imageData []byte) (string, error) {
	apiKey := m.config.AISettings.GeminiApiKey
	if apiKey == "" {
		return "", fmt.Errorf("Gemini API Key is not configured")
	}

	model := m.config.AISettings.GeminiModel
	if model == "" {
		model = defaultModel
	}

	url := fmt.Sprintf(apiBaseURL, model) + "?key=" + apiKey

	// Build request body
	parts := []Part{
		{Text: prompt},
	}

	if len(imageData) > 0 {
		// Detect MIME type or assume jpeg/png based on needs.
		// For simplicity, let's assume valid image data is passed.
		// Detailed MIME type detection can be added if needed.
		// Here we default to image/jpeg for now or try to sniff.
		mimeType := http.DetectContentType(imageData)

		parts = append(parts, Part{
			InlineData: &InlineData{
				MimeType: mimeType,
				Data:     base64.StdEncoding.EncodeToString(imageData),
			},
		})
	}

	reqBody := GenerateContentRequest{
		Contents: []Content{
			{Parts: parts},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(m.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}
