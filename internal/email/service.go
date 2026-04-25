package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/maidulcu/masaar-crm/internal/domain"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type Service struct {
	cfg *Config
}

func NewService(cfg *Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) IsConfigured() bool {
	return s.cfg != nil && s.cfg.SMTPHost != "" && s.cfg.SMTPUser != ""
}

func (s *Service) Send(email *domain.EmailHistory) error {
	if !s.IsConfigured() {
		return fmt.Errorf("email service not configured")
	}

	// Set From if not already set
	if email.FromEmail == "" {
		email.FromEmail = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromEmail)
	}

	// Prepare SMTP auth
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// Build email body
	var body bytes.Buffer

	// Headers
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n",
		email.FromEmail,
		email.ToEmail,
		email.Subject,
	)

	// Content-Type
	if email.HTMLBody != "" {
		headers += "MIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n"
	} else {
		headers += "MIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n"
	}

	body.WriteString(headers + "\r\n")

	// Body
	if email.HTMLBody != "" {
		body.WriteString(email.HTMLBody)
	} else {
		body.WriteString(email.Body)
	}

	// Send via SMTP
	err := smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{email.ToEmail}, body.Bytes())
	return err
}

// ─── Email Templates ──────────────────────────────────────────────────────────

type InvoiceData struct {
	InvoiceNo   string
	DealTitle   string
	Amount      float64
	VATAmount   float64
	Total       float64
	IssuedDate  string
	DueDate     string
	ContactName string
}

func (s *Service) RenderInvoiceTemplate(data InvoiceData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; color: #333; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { border-bottom: 2px solid #0066cc; padding-bottom: 20px; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #0066cc; }
        .invoice-details { margin-bottom: 30px; }
        .invoice-details p { margin: 5px 0; }
        .items { width: 100%; border-collapse: collapse; margin-bottom: 30px; }
        .items th { background-color: #f5f5f5; padding: 10px; text-align: left; font-weight: bold; border-bottom: 1px solid #ddd; }
        .items td { padding: 10px; border-bottom: 1px solid #ddd; }
        .total-row { font-weight: bold; background-color: #f9f9f9; }
        .vat-row { font-weight: bold; color: #0066cc; }
        .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">Masaar CRM</div>
        </div>

        <div class="invoice-details">
            <p><strong>Invoice #:</strong> {{.InvoiceNo}}</p>
            <p><strong>Date:</strong> {{.IssuedDate}}</p>
            <p><strong>Due Date:</strong> {{.DueDate}}</p>
            <p><strong>Client:</strong> {{.ContactName}}</p>
        </div>

        <table class="items">
            <tr>
                <th>Description</th>
                <th style="text-align: right;">Amount</th>
            </tr>
            <tr>
                <td>{{.DealTitle}}</td>
                <td style="text-align: right;">AED {{printf "%.2f" .Amount}}</td>
            </tr>
            <tr class="vat-row">
                <td>VAT (5%)</td>
                <td style="text-align: right;">AED {{printf "%.2f" .VATAmount}}</td>
            </tr>
            <tr class="total-row">
                <td>Total</td>
                <td style="text-align: right;">AED {{printf "%.2f" .Total}}</td>
            </tr>
        </table>

        <div class="footer">
            <p>Thank you for your business. Please retain this invoice for your records.</p>
            <p>&copy; {{.IssuedDate}} Masaar CRM. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("invoice").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
