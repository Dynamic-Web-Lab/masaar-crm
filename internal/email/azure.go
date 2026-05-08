package email

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/maidulcu/masaar-crm/internal/domain"
)

type AzureService struct {
	endpoint    string
	key         []byte
	fromAddress string
	client      *http.Client
}

func NewAzureService(endpoint, key, fromAddress string) *AzureService {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		decoded = []byte(key)
	}
	return &AzureService{
		endpoint:    endpoint,
		key:         decoded,
		fromAddress: fromAddress,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type azureEmailContent struct {
	Subject   string `json:"subject"`
	PlainText string `json:"plainText,omitempty"`
	HTML      string `json:"html,omitempty"`
}

type azureEmailAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"displayName,omitempty"`
}

type azureRecipients struct {
	To []azureEmailAddress `json:"to"`
}

type azureEmailPayload struct {
	Sender     string             `json:"sender"`
	Content    azureEmailContent  `json:"content"`
	Recipients azureRecipients   `json:"recipients"`
}

func (s *AzureService) IsConfigured() bool {
	return s.endpoint != "" && len(s.key) > 0 && s.fromAddress != ""
}

func (s *AzureService) Send(email *domain.EmailHistory) error {
	if !s.IsConfigured() {
		return fmt.Errorf("azure email service not configured")
	}

	payload := azureEmailPayload{
		Sender: s.fromAddress,
		Content: azureEmailContent{
			Subject:   email.Subject,
			PlainText: email.Body,
			HTML:      email.HTMLBody,
		},
		Recipients: azureRecipients{
			To: []azureEmailAddress{{
				Address:     email.ToEmail,
			}},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("azure marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/emails:send?api-version=2024-07-01-preview", s.endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("azure create request: %w", err)
	}

	date := time.Now().UTC().Format(time.RFC1123)
	contentHash := sha256Hex(body)
	host := req.URL.Host

	sig := s.sign(date, host, contentHash)
	auth := fmt.Sprintf("HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=%s", sig)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ms-date", date)
	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("Authorization", auth)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("azure send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("azure email API error: %s - %s", resp.Status, string(respBody))
}

func (s *AzureService) sign(date, host, contentHash string) string {
	stringToSign := fmt.Sprintf("x-ms-date:%s\nhost:%s\nx-ms-content-sha256:%s", date, host, contentHash)
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
