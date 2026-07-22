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
	"github.com/dynamicweblab/masaar-crm/internal/whatsapp"
	"github.com/dynamicweblab/masaar-crm/internal/ws"
)

type PaymentReminderService struct {
	paymentRepo      *repo.PaymentRepo
	reminderRepo     *repo.PaymentReminderRepo
	leaseRepo        *repo.LeaseRepo
	tenantRepo       *repo.TenantRepo
	contactRepo      *repo.ContactRepo
	emailService     *email.Service
	whatsappSender   *whatsapp.Sender
	hub              *ws.Hub
}

func NewPaymentReminderService(
	paymentRepo *repo.PaymentRepo,
	reminderRepo *repo.PaymentReminderRepo,
	leaseRepo *repo.LeaseRepo,
	tenantRepo *repo.TenantRepo,
	contactRepo *repo.ContactRepo,
	emailService *email.Service,
	whatsappSender *whatsapp.Sender,
	hub *ws.Hub,
) *PaymentReminderService {
	return &PaymentReminderService{
		paymentRepo:    paymentRepo,
		reminderRepo:   reminderRepo,
		leaseRepo:      leaseRepo,
		tenantRepo:     tenantRepo,
		contactRepo:    contactRepo,
		emailService:   emailService,
		whatsappSender: whatsappSender,
		hub:            hub,
	}
}

func (s *PaymentReminderService) GenerateReminders(ctx context.Context, companyID uuid.UUID) error {
	today := time.Now()

	reminderDates := []struct {
		reminderType domain.ReminderType
		daysOffset   int
	}{
		{domain.Reminder30Days, -30},
		{domain.Reminder15Days, -15},
		{domain.Reminder7Days, -7},
	}

	for _, rd := range reminderDates {
		targetDate := today.AddDate(0, 0, rd.daysOffset)
		payments, err := s.paymentRepo.GetByDueDate(ctx, companyID, targetDate)
		if err != nil {
			log.Printf("Error fetching payments due on %v: %v", targetDate, err)
			continue
		}

		for _, p := range payments {
			if p.Status != domain.PaymentPending && p.Status != domain.PaymentOverdue {
				continue
			}

			for _, method := range []domain.DeliveryMethod{
				domain.DeliveryWhatsApp,
				domain.DeliveryEmail,
			} {
				existing, err := s.reminderRepo.GetByPaymentAndType(ctx, p.ID, rd.reminderType, method)
				if err == nil && existing != nil {
					continue
				}

				reminder := &domain.PaymentReminder{
					CompanyID:      companyID,
					PaymentID:      p.ID,
					ReminderType:   rd.reminderType,
					ReminderDate:   targetDate,
					DeliveryMethod: method,
					DeliveryStatus: domain.DeliveryPending,
				}

				if err := s.reminderRepo.Create(ctx, reminder); err != nil {
					log.Printf("Error creating reminder for payment %s: %v", p.ID, err)
				}
			}
		}
	}

	lateReminders := []struct {
		reminderType domain.ReminderType
		daysOffset   int
	}{
		{domain.Reminder1DayLate, 1},
		{domain.Reminder5DaysLate, 5},
		{domain.Reminder10DaysLate, 10},
	}

	for _, lr := range lateReminders {
		targetDate := today.AddDate(0, 0, -lr.daysOffset)
		overdue, err := s.paymentRepo.GetOverdueByDate(ctx, companyID, targetDate)
		if err != nil {
			log.Printf("Error fetching overdue payments since %v: %v", targetDate, err)
			continue
		}

		for _, p := range overdue {
			for _, method := range []domain.DeliveryMethod{
				domain.DeliveryWhatsApp,
				domain.DeliveryEmail,
			} {
				existing, err := s.reminderRepo.GetByPaymentAndType(ctx, p.ID, lr.reminderType, method)
				if err == nil && existing != nil {
					continue
				}

				reminder := &domain.PaymentReminder{
					CompanyID:      companyID,
					PaymentID:      p.ID,
					ReminderType:   lr.reminderType,
					ReminderDate:   targetDate,
					DeliveryMethod: method,
					DeliveryStatus: domain.DeliveryPending,
				}

				if err := s.reminderRepo.Create(ctx, reminder); err != nil {
					log.Printf("Error creating late reminder for payment %s: %v", p.ID, err)
				}
			}
		}
	}

	return nil
}

func (s *PaymentReminderService) SendPendingReminders(ctx context.Context, companyID uuid.UUID) error {
	reminders, err := s.reminderRepo.GetPending(ctx, companyID)
	if err != nil {
		return fmt.Errorf("get pending reminders: %w", err)
	}

	for _, reminder := range reminders {
		payment, err := s.paymentRepo.GetByID(ctx, reminder.PaymentID)
		if err != nil {
			log.Printf("Error fetching payment %s: %v", reminder.PaymentID, err)
			s.reminderRepo.MarkFailed(ctx, reminder.ID, "payment not found")
			continue
		}

		lease, err := s.leaseRepo.GetByID(ctx, payment.LeaseID)
		if err != nil {
			log.Printf("Error fetching lease %s: %v", payment.LeaseID, err)
			s.reminderRepo.MarkFailed(ctx, reminder.ID, "lease not found")
			continue
		}

		tenant, err := s.tenantRepo.GetByID(ctx, lease.TenantID)
		if err != nil {
			log.Printf("Error fetching tenant %s: %v", lease.TenantID, err)
			s.reminderRepo.MarkFailed(ctx, reminder.ID, "tenant not found")
			continue
		}

		switch reminder.DeliveryMethod {
		case domain.DeliveryWhatsApp:
			if s.whatsappSender != nil && tenant.PhoneWA != "" {
				msg := s.buildReminderMessage(payment, lease, tenant, reminder.ReminderType)
				_, err := s.whatsappSender.SendMessage(ctx, tenant.PhoneWA, msg)
				if err != nil {
					s.reminderRepo.MarkFailed(ctx, reminder.ID, err.Error())
					continue
				}
			}
		case domain.DeliveryEmail:
			if tenant.Email != "" {
				subject := s.buildReminderSubject(reminder.ReminderType)
				body := s.buildReminderEmailBody(payment, lease, tenant, reminder.ReminderType)
				emailHist := &domain.EmailHistory{
					ToEmail: tenant.Email,
					Subject: subject,
					Body:    body,
					Status:  domain.EmailPending,
				}
				if err := s.emailService.Send(emailHist); err != nil {
					s.reminderRepo.MarkFailed(ctx, reminder.ID, err.Error())
					continue
				}
			}
		}

		if err := s.reminderRepo.MarkSent(ctx, reminder.ID); err != nil {
			log.Printf("Error marking reminder as sent: %v", err)
		}

		s.hub.SendToUser(companyID.String(), ws.Event{
			Type: "payment_reminder_sent",
			Payload: map[string]string{
				"payment_id": payment.ID.String(),
				"status":     string(payment.Status),
			},
		})
	}

	return nil
}

func (s *PaymentReminderService) buildReminderMessage(payment *domain.Payment, lease *domain.Lease, tenant *domain.Tenant, reminderType domain.ReminderType) string {
	switch reminderType {
	case domain.Reminder30Days:
		return fmt.Sprintf("تذكير: الدفع لعقد الإيجار مستحق خلال 30 يوم. المبلغ: %s %.2f\n\nReminder: Lease rent payment due in 30 days. Amount: %s %.2f", payment.Currency, payment.Amount, payment.Currency, payment.Amount)
	case domain.Reminder15Days:
		return fmt.Sprintf("تذكير: الدفع لعقد الإيجار مستحق خلال 15 يوم. المبلغ: %s %.2f\n\nReminder: Lease rent payment due in 15 days. Amount: %s %.2f", payment.Currency, payment.Amount, payment.Currency, payment.Amount)
	case domain.Reminder7Days:
		return fmt.Sprintf("تذكير عاجل: الدفع لعقد الإيجار مستحق خلال 7 أيام. المبلغ: %s %.2f\n\nUrgent Reminder: Lease rent payment due in 7 days. Amount: %s %.2f", payment.Currency, payment.Amount, payment.Currency, payment.Amount)
	case domain.Reminder1DayLate:
		return fmt.Sprintf("⚠️ تنبيه: دفعة الإيجار متأخرة بـ 1 يوم. المبلغ المستحق: %s %.2f\n\n⚠️ Alert: Lease payment is 1 day overdue. Amount due: %s %.2f", payment.Currency, payment.Amount+payment.LateFeeAmount, payment.Currency, payment.Amount+payment.LateFeeAmount)
	case domain.Reminder5DaysLate:
		return fmt.Sprintf("🔴 تنبيه عاجل: دفعة الإيجار متأخرة بـ 5 أيام. يرجى الدفع فوراً. المبلغ المستحق: %s %.2f\n\n🔴 Urgent Alert: Lease payment is 5 days overdue. Please pay immediately. Amount due: %s %.2f", payment.Currency, payment.Amount+payment.LateFeeAmount, payment.Currency, payment.Amount+payment.LateFeeAmount)
	case domain.Reminder10DaysLate:
		return fmt.Sprintf("🔴🔴 تنبيه عاجل جداً: دفعة الإيجار متأخرة بـ 10 أيام. اتصل بنا حالاً. المبلغ المستحق: %s %.2f\n\n🔴🔴 Critical Alert: Lease payment is 10 days overdue. Contact us immediately. Amount due: %s %.2f", payment.Currency, payment.Amount+payment.LateFeeAmount, payment.Currency, payment.Amount+payment.LateFeeAmount)
	}
	return "Payment reminder"
}

func (s *PaymentReminderService) buildReminderSubject(reminderType domain.ReminderType) string {
	switch reminderType {
	case domain.Reminder30Days:
		return "تذكير الدفع: 30 يوم متبقي | Payment Reminder: 30 Days Left"
	case domain.Reminder15Days:
		return "تذكير الدفع: 15 يوم متبقي | Payment Reminder: 15 Days Left"
	case domain.Reminder7Days:
		return "تذكير عاجل: 7 أيام متبقية | Urgent Reminder: 7 Days Left"
	case domain.Reminder1DayLate:
		return "⚠️ تنبيه: الدفع متأخر | ⚠️ Alert: Payment Overdue"
	case domain.Reminder5DaysLate:
		return "🔴 تنبيه عاجل: 5 أيام متأخر | 🔴 Urgent: 5 Days Overdue"
	case domain.Reminder10DaysLate:
		return "🔴🔴 تنبيه حرج: 10 أيام متأخرة | 🔴🔴 Critical: 10 Days Overdue"
	}
	return "Payment Reminder"
}

func (s *PaymentReminderService) buildReminderEmailBody(payment *domain.Payment, lease *domain.Lease, tenant *domain.Tenant, reminderType domain.ReminderType) string {
	header := "تذكير الدفع / Payment Reminder\n\n"
	details := fmt.Sprintf("المستأجر / Tenant: %s\nالعقار / Property: %s\nالمبلغ / Amount: %s %.2f\nتاريخ الاستحقاق / Due Date: %s\n",
		tenant.FullNameEN, "Property", payment.Currency, payment.Amount, payment.DueDate.Format("2006-01-02"))

	switch reminderType {
	case domain.Reminder30Days:
		return header + "تذكير: دفعة الإيجار مستحقة خلال 30 يوم\nReminder: Lease payment is due in 30 days\n\n" + details
	case domain.Reminder7Days:
		return header + "تذكير عاجل: دفعة الإيجار مستحقة خلال 7 أيام\nUrgent Reminder: Lease payment is due in 7 days\n\n" + details
	case domain.Reminder1DayLate:
		return header + "تنبيه: دفعة الإيجار متأخرة بـ 1 يوم\nAlert: Lease payment is 1 day overdue\n\n" + details
	case domain.Reminder5DaysLate:
		return header + "تنبيه عاجل: دفعة الإيجار متأخرة بـ 5 أيام. يرجى الدفع فوراً\nUrgent Alert: Lease payment is 5 days overdue. Please pay immediately\n\n" + details
	}
	return header + details
}
