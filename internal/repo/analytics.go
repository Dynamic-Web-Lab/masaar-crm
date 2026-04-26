package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maidulcu/masaar-crm/internal/domain"
)

type AnalyticsRepository struct {
	conn *pgxpool.Pool
}

func NewAnalyticsRepository(conn *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{conn: conn}
}

func (r *AnalyticsRepository) GetTenantAnalytics(ctx context.Context, companyID uuid.UUID) (*domain.TenantAnalytics, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM tenants WHERE company_id = $1 AND status = 'active') as active_tenants,
			(SELECT COUNT(*) FROM tenants WHERE company_id = $1 AND status != 'active') as inactive_tenants,
			(SELECT COUNT(*) FROM tenants WHERE company_id = $1) as total_tenants,
			(SELECT COUNT(*) FROM rental_properties rp WHERE rp.company_id = $1 AND rp.occupancy_status = 'vacant') as vacant_units,
			(SELECT COUNT(*) FROM rental_properties rp WHERE rp.company_id = $1 AND rp.occupancy_status = 'occupied') as occupied_units,
			(SELECT COUNT(*) FROM leases l WHERE l.company_id = $1 AND l.status = 'active') as active_leases,
			(SELECT COUNT(*) FROM payments WHERE company_id = $1 AND status = 'overdue') as overdue_payments,
			COALESCE((SELECT SUM(amount) FROM payments WHERE company_id = $1 AND status = 'overdue'), 0) as overdue_amount,
			COALESCE((SELECT AVG(monthly_rent) FROM leases WHERE company_id = $1 AND status = 'active'), 0) as avg_rent,
			COALESCE((SELECT SUM(monthly_rent) FROM leases WHERE company_id = $1 AND status = 'active'), 0) as total_monthly_revenue,
			(SELECT COUNT(*) FROM leases WHERE company_id = $1 AND end_date BETWEEN NOW() AND NOW() + INTERVAL '60 days' AND status = 'active') as upcoming_renewals
	`

	var activeTenants, inactiveTenants, totalTenants, vacantUnits, occupiedUnits, activeLeases, overduePayments, upcomingRenewals int
	var overdueAmount, avgRent, totalRevenue float64

	err := r.conn.QueryRow(ctx, query, companyID).Scan(
		&activeTenants,
		&inactiveTenants,
		&totalTenants,
		&vacantUnits,
		&occupiedUnits,
		&activeLeases,
		&overduePayments,
		&overdueAmount,
		&avgRent,
		&totalRevenue,
		&upcomingRenewals,
	)
	if err != nil {
		return nil, err
	}

	totalUnits := vacantUnits + occupiedUnits
	occupancyRate := 0.0
	if totalUnits > 0 {
		occupancyRate = float64(occupiedUnits) / float64(totalUnits) * 100
	}

	// Calculate collection rate: received payments / total due payments
	collectionRateQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'received' THEN amount ELSE 0 END), 0) as collected,
			COALESCE(SUM(amount), 0) as total_due
		FROM payments
		WHERE company_id = $1 AND due_date <= NOW()
	`

	var collected, totalDue float64
	err = r.conn.QueryRow(ctx, collectionRateQuery, companyID).Scan(&collected, &totalDue)
	if err != nil {
		return nil, err
	}

	collectionRate := 0.0
	if totalDue > 0 {
		collectionRate = (collected / totalDue) * 100
	}

	// Calculate tenant churn: moved out last 12 months / average active tenants
	churnQuery := `
		SELECT COUNT(*) FROM leases
		WHERE company_id = $1
		AND status = 'terminated'
		AND termination_date >= NOW() - INTERVAL '12 months'
	`
	var churned int
	err = r.conn.QueryRow(ctx, churnQuery, companyID).Scan(&churned)
	if err != nil {
		return nil, err
	}

	churnRate := 0.0
	if activeTenants > 0 {
		churnRate = (float64(churned) / float64(activeTenants)) * 100
	}

	return &domain.TenantAnalytics{
		TotalTenants:        totalTenants,
		ActiveTenants:       activeTenants,
		InactiveTenants:     inactiveTenants,
		VacantUnits:         vacantUnits,
		OccupiedUnits:       occupiedUnits,
		OccupancyRate:       occupancyRate,
		AverageRentPerUnit:  avgRent,
		TotalMonthlyRevenue: totalRevenue,
		CollectionRate:      collectionRate,
		OverduePayments:     overduePayments,
		OverdueDuesAmount:   overdueAmount,
		UpcomingRenewals:    upcomingRenewals,
		TenantChurnRate:     churnRate,
	}, nil
}

func (r *AnalyticsRepository) GetPropertyAnalytics(ctx context.Context, companyID uuid.UUID, propertyID uuid.UUID) (*domain.PropertyAnalytics, error) {
	query := `
		SELECT
			rp.id,
			rp.name,
			rp.property_type,
			rp.area,
			rp.units_count,
			COALESCE((SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND status = 'active'), 0) as occupied,
			(rp.units_count - COALESCE((SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND status = 'active'), 0)) as vacant,
			COALESCE((SELECT SUM(l.monthly_rent) FROM leases l WHERE l.property_id = rp.id AND l.status = 'active'), 0) as monthly_revenue,
			(SELECT COUNT(*) FROM maintenance_tasks WHERE property_id = rp.id AND status IN ('pending', 'scheduled')) as pending_maintenance,
			(SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND status = 'active') as active_leases,
			(SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND end_date BETWEEN NOW() AND NOW() + INTERVAL '60 days' AND status = 'active') as expiring_leases
		FROM rental_properties rp
		WHERE rp.company_id = $1 AND rp.id = $2
	`

	var propertyAnalytics domain.PropertyAnalytics
	var occupied, vacant, activeLs, expiringLs, pendingMaint int
	var monthlyRev float64

	err := r.conn.QueryRow(ctx, query, companyID, propertyID).Scan(
		&propertyAnalytics.PropertyID,
		&propertyAnalytics.PropertyName,
		&propertyAnalytics.PropertyType,
		&propertyAnalytics.Area,
		&propertyAnalytics.TotalUnits,
		&occupied,
		&vacant,
		&monthlyRev,
		&pendingMaint,
		&activeLs,
		&expiringLs,
	)
	if err != nil {
		return nil, err
	}

	propertyAnalytics.OccupiedUnits = occupied
	propertyAnalytics.VacantUnits = vacant
	propertyAnalytics.MonthlyRevenue = monthlyRev
	propertyAnalytics.MaintenanceNeeded = pendingMaint
	propertyAnalytics.ActiveLeases = activeLs
	propertyAnalytics.ExpiringLeases = expiringLs

	if propertyAnalytics.TotalUnits > 0 {
		propertyAnalytics.OccupancyRate = float64(occupied) / float64(propertyAnalytics.TotalUnits) * 100
	}

	// Get operating expenses for this property (last 30 days)
	expenseQuery := `
		SELECT COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE company_id = $1 AND property_id = $2 AND expense_date >= NOW() - INTERVAL '30 days'
	`
	var expenses float64
	err = r.conn.QueryRow(ctx, expenseQuery, companyID, propertyID).Scan(&expenses)
	if err == nil {
		propertyAnalytics.OperatingExpenses = expenses
		propertyAnalytics.NetOperatingIncome = monthlyRev - expenses
	}

	return &propertyAnalytics, nil
}

func (r *AnalyticsRepository) ListPropertiesAnalytics(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]domain.PropertyAnalytics, int, error) {
	countQuery := `SELECT COUNT(*) FROM rental_properties WHERE company_id = $1`
	var total int
	err := r.conn.QueryRow(ctx, countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			rp.id,
			rp.name,
			rp.property_type,
			rp.area,
			rp.units_count,
			COALESCE((SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND status = 'active'), 0) as occupied,
			(rp.units_count - COALESCE((SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND status = 'active'), 0)) as vacant,
			COALESCE((SELECT SUM(l.monthly_rent) FROM leases l WHERE l.property_id = rp.id AND l.status = 'active'), 0) as monthly_revenue,
			(SELECT COUNT(*) FROM maintenance_tasks WHERE property_id = rp.id AND status IN ('pending', 'scheduled')) as pending_maintenance,
			(SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND status = 'active') as active_leases,
			(SELECT COUNT(*) FROM leases WHERE property_id = rp.id AND end_date BETWEEN NOW() AND NOW() + INTERVAL '60 days' AND status = 'active') as expiring_leases
		FROM rental_properties rp
		WHERE rp.company_id = $1
		ORDER BY rp.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.conn.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var properties []domain.PropertyAnalytics
	for rows.Next() {
		var p domain.PropertyAnalytics
		var occupied, vacant, activeLs, expiringLs, pendingMaint int
		var monthlyRev float64

		err := rows.Scan(
			&p.PropertyID,
			&p.PropertyName,
			&p.PropertyType,
			&p.Area,
			&p.TotalUnits,
			&occupied,
			&vacant,
			&monthlyRev,
			&pendingMaint,
			&activeLs,
			&expiringLs,
		)
		if err != nil {
			return nil, 0, err
		}

		p.OccupiedUnits = occupied
		p.VacantUnits = vacant
		p.MonthlyRevenue = monthlyRev
		p.MaintenanceNeeded = pendingMaint
		p.ActiveLeases = activeLs
		p.ExpiringLeases = expiringLs

		if p.TotalUnits > 0 {
			p.OccupancyRate = float64(occupied) / float64(p.TotalUnits) * 100
		}

		properties = append(properties, p)
	}

	return properties, total, rows.Err()
}

func (r *AnalyticsRepository) GetTenantPerformance(ctx context.Context, companyID uuid.UUID, tenantID uuid.UUID) (*domain.TenantPerformanceMetrics, error) {
	query := `
		SELECT
			t.id,
			COALESCE(t.full_name_en, t.full_name_ar) as name,
			t.status,
			(SELECT COUNT(*) FROM leases WHERE tenant_id = t.id) as rental_history,
			(SELECT COUNT(*) FROM leases WHERE tenant_id = t.id AND status = 'terminated') as terminated_count
		FROM tenants t
		WHERE t.company_id = $1 AND t.id = $2
	`

	var tenantID2 uuid.UUID
	var name, status string
	var rentalHistory, terminatedCount int

	err := r.conn.QueryRow(ctx, query, companyID, tenantID).Scan(
		&tenantID2,
		&name,
		&status,
		&rentalHistory,
		&terminatedCount,
	)
	if err != nil {
		return nil, err
	}

	metrics := &domain.TenantPerformanceMetrics{
		TenantID:      tenantID,
		TenantName:    name,
		RentalHistory: rentalHistory,
		Status:        status,
	}

	// Calculate average stay
	avgStayQuery := `
		SELECT COALESCE(AVG(EXTRACT(DAY FROM (end_date - start_date))), 0)
		FROM leases
		WHERE tenant_id = $1 AND status IN ('terminated', 'expired')
	`
	err = r.conn.QueryRow(ctx, avgStayQuery, tenantID).Scan(&metrics.AverageStay)
	if err != nil {
		return nil, err
	}

	// Calculate payment on-time rate
	paymentQuery := `
		SELECT
			COALESCE(COUNT(*) FILTER (WHERE status = 'received' AND paid_date <= due_date), 0) as on_time,
			COALESCE(COUNT(*), 0) as total
		FROM payments p
		JOIN leases l ON p.lease_id = l.id
		WHERE l.tenant_id = $1 AND p.status = 'received'
	`
	var onTime, totalPayments int
	err = r.conn.QueryRow(ctx, paymentQuery, tenantID).Scan(&onTime, &totalPayments)
	if err != nil {
		return nil, err
	}

	if totalPayments > 0 {
		metrics.PaymentOnTimeRate = (float64(onTime) / float64(totalPayments)) * 100
	}

	// Calculate risk score (0-100)
	// Factors: late payments (40%), disputes (30%), tenure (30%)
	disputeQuery := `SELECT COUNT(*) FROM leads WHERE contact_id = (SELECT id FROM contacts WHERE email = (SELECT email FROM tenants WHERE id = $1)) AND stage = 'lost'`
	var disputes int
	err = r.conn.QueryRow(ctx, disputeQuery, tenantID).Scan(&disputes)
	if err == nil {
		metrics.DisputeCount = disputes
	}

	latePaymentScore := 0
	if totalPayments > 0 {
		latePayments := totalPayments - onTime
		latePaymentScore = int((float64(latePayments) / float64(totalPayments)) * 40)
	}

	disputeScore := disputes * 10
	if disputeScore > 30 {
		disputeScore = 30
	}

	tenureScore := 0
	if rentalHistory > 5 {
		tenureScore = 0
	} else if rentalHistory > 2 {
		tenureScore = 15
	} else {
		tenureScore = 30
	}

	metrics.RiskScore = latePaymentScore + disputeScore + tenureScore

	return metrics, nil
}

func (r *AnalyticsRepository) ListTenantsPerformance(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]domain.TenantPerformanceMetrics, int, error) {
	countQuery := `SELECT COUNT(*) FROM tenants WHERE company_id = $1`
	var total int
	err := r.conn.QueryRow(ctx, countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			t.id,
			COALESCE(t.full_name_en, t.full_name_ar) as name,
			t.status,
			(SELECT COUNT(*) FROM leases WHERE tenant_id = t.id) as rental_history,
			(SELECT COUNT(*) FROM leases WHERE tenant_id = t.id AND status = 'terminated') as terminated_count
		FROM tenants t
		WHERE t.company_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.conn.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var metrics []domain.TenantPerformanceMetrics
	for rows.Next() {
		var m domain.TenantPerformanceMetrics
		var status string
		var terminatedCount int

		err := rows.Scan(
			&m.TenantID,
			&m.TenantName,
			&status,
			&m.RentalHistory,
			&terminatedCount,
		)
		if err != nil {
			return nil, 0, err
		}

		m.Status = status
		metrics = append(metrics, m)
	}

	return metrics, total, rows.Err()
}

func (r *AnalyticsRepository) GetFinancialAnalytics(ctx context.Context, companyID uuid.UUID, startDate, endDate time.Time) (*domain.FinancialAnalytics, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'received' THEN amount ELSE 0 END), 0) as rent_collected,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0) as rent_pending,
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN amount ELSE 0 END), 0) as rent_overdue,
			COALESCE(SUM(amount) FILTER (WHERE category_type = 'utilities'), 0) as utilities_expense,
			COALESCE(SUM(amount) FILTER (WHERE category_type IN ('plumbing', 'electrical', 'hvac', 'flooring', 'painting', 'structural')), 0) as maintenance_expense,
			COALESCE(SUM(amount) FILTER (WHERE category_type NOT IN ('utilities', 'plumbing', 'electrical', 'hvac', 'flooring', 'painting', 'structural')), 0) as other_expense
		FROM (
			SELECT amount, NULL::VARCHAR as category_type, status FROM payments WHERE company_id = $1 AND due_date BETWEEN $2 AND $3
			UNION ALL
			SELECT amount, category_type, NULL::VARCHAR as status FROM expenses WHERE company_id = $1 AND expense_date BETWEEN $2 AND $3
		) combined_data
	`

	var rentCollected, rentPending, rentOverdue, utilitiesExp, maintenanceExp, otherExp float64

	err := r.conn.QueryRow(ctx, query, companyID, startDate, endDate).Scan(
		&rentCollected,
		&rentPending,
		&rentOverdue,
		&utilitiesExp,
		&maintenanceExp,
		&otherExp,
	)
	if err != nil {
		return nil, err
	}

	totalExpenses := utilitiesExp + maintenanceExp + otherExp
	totalRevenue := rentCollected
	netProfit := totalRevenue - totalExpenses

	profitMargin := 0.0
	if totalRevenue > 0 {
		profitMargin = (netProfit / totalRevenue) * 100
	}

	return &domain.FinancialAnalytics{
		Period:              startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		TotalRevenue:        totalRevenue,
		TotalExpenses:       totalExpenses,
		NetProfit:           netProfit,
		ProfitMargin:        profitMargin,
		RentCollected:       rentCollected,
		RentPending:         rentPending,
		UtilitiesExpense:    utilitiesExp,
		MaintenanceExpense:  maintenanceExp,
		OtherExpenses:       otherExp,
	}, nil
}

func (r *AnalyticsRepository) GetMaintenanceAnalytics(ctx context.Context, companyID uuid.UUID) (*domain.MaintenanceAnalytics, error) {
	query := `
		SELECT
			COUNT(*) as total_tasks,
			COUNT(*) FILTER (WHERE status = 'completed') as completed_tasks,
			COUNT(*) FILTER (WHERE status IN ('pending', 'scheduled', 'in_progress')) as pending_tasks,
			COUNT(*) FILTER (WHERE priority = 'high' OR priority = 'urgent') as high_priority,
			COALESCE(AVG(EXTRACT(DAY FROM (completion_date - created_at))) FILTER (WHERE status = 'completed'), 0) as avg_completion_days
		FROM maintenance_tasks
		WHERE company_id = $1
	`

	var analytics domain.MaintenanceAnalytics
	var completedTasks, pendingTasks, avgDays int
	var highPriority int

	err := r.conn.QueryRow(ctx, query, companyID).Scan(
		&analytics.TotalTasks,
		&completedTasks,
		&pendingTasks,
		&avgDays,
		&highPriority,
	)
	if err != nil {
		return nil, err
	}

	analytics.CompletedTasks = completedTasks
	analytics.PendingTasks = pendingTasks
	analytics.HighPriorityTasks = highPriority
	analytics.AvgCompletionDays = float64(avgDays)

	if analytics.TotalTasks > 0 {
		analytics.CompletionRate = (float64(completedTasks) / float64(analytics.TotalTasks)) * 100
	}

	return &analytics, nil
}
