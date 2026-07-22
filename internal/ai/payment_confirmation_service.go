package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/dynamicweblab/masaar-crm/internal/domain"
	"github.com/dynamicweblab/masaar-crm/internal/email"
	"github.com/dynamicweblab/masaar-crm/internal/repo"
)

type PaymentConfirmationService struct {
	paymentRepo        *repo.PaymentRepo
	confirmationRepo   *repo.PaymentConfirmationRepo
	leaseRepo          *repo.LeaseRepo
	tenantRepo         *repo.TenantRepo
	propertyRepo       *repo.RentalPropertyRepo
	companySettingsRepo *repo.CompanySettingsRepo
	emailService       *email.Service
}

func NewPaymentConfirmationService(
	paymentRepo *repo.PaymentRepo,
	confirmationRepo *repo.PaymentConfirmationRepo,
	leaseRepo *repo.LeaseRepo,
	tenantRepo *repo.TenantRepo,
	propertyRepo *repo.RentalPropertyRepo,
	companySettingsRepo *repo.CompanySettingsRepo,
	emailService *email.Service,
) *PaymentConfirmationService {
	return &PaymentConfirmationService{
		paymentRepo:         paymentRepo,
		confirmationRepo:    confirmationRepo,
		leaseRepo:           leaseRepo,
		tenantRepo:          tenantRepo,
		propertyRepo:        propertyRepo,
		companySettingsRepo: companySettingsRepo,
		emailService:        emailService,
	}
}

func (s *PaymentConfirmationService) CreateAndSendConfirmation(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("get payment: %w", err)
	}

	lease, err := s.leaseRepo.GetByID(ctx, payment.LeaseID)
	if err != nil {
		return fmt.Errorf("get lease: %w", err)
	}

	tenant, err := s.tenantRepo.GetByID(ctx, lease.TenantID)
	if err != nil {
		return fmt.Errorf("get tenant: %w", err)
	}

	property, err := s.propertyRepo.GetByID(ctx, lease.PropertyID)
	if err != nil {
		return fmt.Errorf("get property: %w", err)
	}

	companySettings, err := s.companySettingsRepo.Get(ctx)
	if err != nil {
		log.Printf("could not load company settings: %v", err)
		companySettings = &domain.CompanySettings{}
	}

	confirmationNum := fmt.Sprintf("CONF-%s-%d", payment.ID.String()[:8], time.Now().Unix())

	confirmation := &domain.PaymentConfirmation{
		CompanyID:          payment.CompanyID,
		PaymentID:          payment.ID,
		ConfirmationNumber: confirmationNum,
		TenantEmail:        tenant.Email,
		TenantPhone:        &tenant.PhoneWA,
		DeliveryMethod:     domain.ConfDeliveryEmail,
		DataClassification: domain.ClassConfidential,
		RetentionUntil:     s.getRetentionDate(),
	}

	if err := s.confirmationRepo.Create(ctx, confirmation); err != nil {
		return fmt.Errorf("create confirmation: %w", err)
	}

	subject := s.buildConfirmationSubject(payment)
	body := s.buildConfirmationEmail(payment, lease, tenant, property, companySettings, confirmationNum)

	emailHist := &domain.EmailHistory{
		ToEmail:   tenant.Email,
		Subject:   subject,
		Body:      body,
		Status:    domain.EmailPending,
		RelatedTo: "payment_confirmation",
		Metadata: map[string]any{
			"payment_id":        payment.ID.String(),
			"confirmation_num":  confirmationNum,
			"data_classification": "confidential",
		},
	}

	if err := s.emailService.Send(emailHist); err != nil {
		s.confirmationRepo.MarkFailed(ctx, confirmation.ID)
		return fmt.Errorf("send confirmation email: %w", err)
	}

	if err := s.confirmationRepo.MarkSent(ctx, confirmation.ID, nil); err != nil {
		log.Printf("error marking confirmation as sent: %v", err)
	}

	log.Printf("Payment confirmation sent for payment %s to %s", payment.ID, tenant.Email)
	return nil
}

func (s *PaymentConfirmationService) SendPendingConfirmations(ctx context.Context, companyID uuid.UUID) error {
	confirmations, err := s.confirmationRepo.GetPending(ctx, companyID)
	if err != nil {
		return fmt.Errorf("get pending confirmations: %w", err)
	}

	for _, conf := range confirmations {
		payment, err := s.paymentRepo.GetByID(ctx, conf.PaymentID)
		if err != nil {
			log.Printf("error fetching payment %s: %v", conf.PaymentID, err)
			s.confirmationRepo.MarkFailed(ctx, conf.ID)
			continue
		}

		subject := s.buildConfirmationSubject(payment)
		lease, _ := s.leaseRepo.GetByID(ctx, payment.LeaseID)
		tenant, _ := s.tenantRepo.GetByID(ctx, lease.TenantID)
		property, _ := s.propertyRepo.GetByID(ctx, lease.PropertyID)
		companySettings, _ := s.companySettingsRepo.Get(ctx)

		body := s.buildConfirmationEmail(payment, lease, tenant, property, companySettings, conf.ConfirmationNumber)

		emailHist := &domain.EmailHistory{
			ToEmail:   conf.TenantEmail,
			Subject:   subject,
			Body:      body,
			Status:    domain.EmailPending,
			RelatedTo: "payment_confirmation",
		}

		if err := s.emailService.Send(emailHist); err != nil {
			log.Printf("error sending confirmation for %s: %v", conf.ID, err)
			s.confirmationRepo.MarkFailed(ctx, conf.ID)
			continue
		}

		s.confirmationRepo.MarkSent(ctx, conf.ID, nil)
	}

	return nil
}

func (s *PaymentConfirmationService) buildConfirmationSubject(payment *domain.Payment) string {
	return fmt.Sprintf("Payment Confirmation Receipt - Ref: %s | إيصال تأكيد الدفع", payment.PaymentReference)
}

func (s *PaymentConfirmationService) buildConfirmationEmail(
	payment *domain.Payment,
	lease *domain.Lease,
	tenant *domain.Tenant,
	property *domain.RentalProperty,
	companySettings *domain.CompanySettings,
	confirmationNum string,
) string {
	return fmt.Sprintf(`
<html dir="rtl" lang="ar">
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { text-align: center; border-bottom: 2px solid #0066cc; padding-bottom: 15px; margin-bottom: 20px; }
		.h1 { color: #0066cc; font-size: 24px; margin: 0; }
		.section { margin: 20px 0; }
		.label { font-weight: bold; color: #0066cc; display: inline-block; width: 150px; }
		.value { display: inline-block; }
		.amount { font-size: 18px; font-weight: bold; color: #28a745; }
		.footer { text-align: center; margin-top: 30px; padding-top: 15px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
		.compliance { background: #f0f0f0; padding: 10px; border-radius: 5px; font-size: 12px; margin-top: 20px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>إيصال تأكيد الدفع / Payment Confirmation Receipt</h1>
		</div>

		<div class="section">
			<p><strong>تاريخ / Date:</strong> %s</p>
			<p><strong>رقم التأكيد / Confirmation #:</strong> %s</p>
		</div>

		<div class="section">
			<h2 style="color: #0066cc;">معلومات الدفع / Payment Details</h2>
			<p>
				<span class="label">المبلغ / Amount:</span>
				<span class="amount">%s %.2f</span>
			</p>
			<p>
				<span class="label">تاريخ الدفع / Payment Date:</span>
				<span class="value">%s</span>
			</p>
			<p>
				<span class="label">طريقة الدفع / Method:</span>
				<span class="value">%s</span>
			</p>
			<p>
				<span class="label">الرقم المرجعي / Reference:</span>
				<span class="value">%s</span>
			</p>
		</div>

		<div class="section">
			<h2 style="color: #0066cc;">معلومات العقار / Property Information</h2>
			<p>
				<span class="label">المستأجر / Tenant:</span>
				<span class="value">%s</span>
			</p>
			<p>
				<span class="label">العقار / Property:</span>
				<span class="value">%s</span>
			</p>
			<p>
				<span class="label">الموقع / Location:</span>
				<span class="value">%s, %s</span>
			</p>
		</div>

		<div class="compliance">
			<p><strong>🔒 Data Privacy Notice / إشعار خصوصية البيانات</strong></p>
			<p>This confirmation contains confidential information classified as "Confidential" under UAE Data Protection Law (PDPL). It is intended for the exclusive use of the recipient. Unauthorized access or disclosure is prohibited. / يحتوي هذا الإيصال على معلومات سرية مصنفة كـ "سرية" بموجب قانون حماية البيانات الشخصية الإماراتي. وهو مخصص للاستخدام الحصري من قبل المتلقي فقط.</p>
			<p>Data Retention: This record will be retained for 7 years in accordance with UAE accounting and tax requirements.</p>
		</div>

		<div class="footer">
			<p>شكراً لثقتكم بنا / Thank you for your business</p>
			<p>هذا الإيصال تم إنشاؤه بواسطة نظام إدارة الممتلكات / Generated by Property Management System</p>
		</div>
	</div>
</body>
</html>
`,
		time.Now().Format("2006-01-02"),
		confirmationNum,
		payment.Currency, payment.Amount,
		time.Now().Format("2006-01-02"),
		string(payment.PaymentMethod),
		payment.PaymentReference,
		tenant.FullNameEN,
		property.Name,
		property.City, property.Emirate,
	)
}

func (s *PaymentConfirmationService) getRetentionDate() *time.Time {
	retention := time.Now().AddDate(7, 0, 0)
	return &retention
}
