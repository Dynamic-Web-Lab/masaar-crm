package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// PropertyReportData holds everything needed to render a client-facing property report.
type PropertyReportData struct {
	// Company branding
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string
	AgentName      string

	// Client brief
	ClientName   string
	Area         string
	PropertyType string
	Bedrooms     string
	BudgetMin    float64
	BudgetMax    float64
	Currency     string // default AED

	// Market data (from DLD — nil fields are skipped gracefully)
	AreaSummary    map[string]interface{}   // area summary stats
	Comparables    []map[string]interface{} // recent DLD transactions
	YieldData      []map[string]interface{} // rental yield rows
	POIs           []map[string]interface{} // nearby POIs
	AIDescription  string                   // AI-generated property description

	GeneratedAt time.Time
}

// colour constants (UAE corporate palette)
const (
	headerR, headerG, headerB = 26, 58, 92  // dark navy
	accentR, accentG, accentB = 180, 140, 60 // warm gold
	lightR, lightG, lightB    = 245, 247, 250 // light grey background
	textR, textG, textB       = 40, 40, 40   // near-black text
)

func GeneratePropertyReport(d PropertyReportData) ([]byte, error) {
	if d.Currency == "" {
		d.Currency = "AED"
	}
	if d.GeneratedAt.IsZero() {
		d.GeneratedAt = time.Now()
	}

	fpdf := gofpdf.New("P", "mm", "A4", "")
	fpdf.SetMargins(15, 15, 15)
	fpdf.SetAutoPageBreak(true, 18)
	fpdf.AddPage()

	pageW := 180.0 // usable width

	// ── Header band ─────────────────────────────────────────────────────────
	fpdf.SetFillColor(headerR, headerG, headerB)
	fpdf.Rect(0, 0, 210, 28, "F")

	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 16)
	fpdf.SetXY(15, 7)
	fpdf.CellFormat(130, 8, d.CompanyName, "", 0, "L", false, 0, "")

	fpdf.SetFont("Helvetica", "", 8)
	fpdf.SetXY(15, 16)
	fpdf.CellFormat(130, 5, "Property Market Research Report", "", 0, "L", false, 0, "")

	fpdf.SetFont("Helvetica", "", 8)
	fpdf.SetXY(145, 9)
	fpdf.CellFormat(50, 5, d.GeneratedAt.Format("02 January 2006"), "", 0, "R", false, 0, "")
	if d.AgentName != "" {
		fpdf.SetXY(145, 14)
		fpdf.CellFormat(50, 5, "Prepared by: "+d.AgentName, "", 0, "R", false, 0, "")
	}

	// Gold accent line under header
	fpdf.SetFillColor(accentR, accentG, accentB)
	fpdf.Rect(0, 28, 210, 1.5, "F")

	fpdf.SetTextColor(textR, textG, textB)
	fpdf.SetY(36)

	// ── Client brief ────────────────────────────────────────────────────────
	sectionHeader(fpdf, "CLIENT BRIEF", pageW)

	fpdf.SetFillColor(lightR, lightG, lightB)
	fpdf.RoundedRect(15, fpdf.GetY(), pageW, 28, 2, "1234", "F")
	fpdf.SetY(fpdf.GetY() + 4)

	col := pageW / 3
	briefRow(fpdf, 15, "Client", d.ClientName, col)
	briefRow(fpdf, 15+col, "Area", d.Area, col)
	briefRow(fpdf, 15+col*2, "Property Type", d.PropertyType, col)
	fpdf.Ln(10)
	briefRow(fpdf, 15, "Bedrooms", d.Bedrooms, col)
	if d.BudgetMin > 0 || d.BudgetMax > 0 {
		budget := ""
		if d.BudgetMin > 0 && d.BudgetMax > 0 {
			budget = fmt.Sprintf("%s %s – %s", d.Currency, formatAED(d.BudgetMin), formatAED(d.BudgetMax))
		} else if d.BudgetMax > 0 {
			budget = fmt.Sprintf("Up to %s %s", d.Currency, formatAED(d.BudgetMax))
		} else {
			budget = fmt.Sprintf("From %s %s", d.Currency, formatAED(d.BudgetMin))
		}
		briefRow(fpdf, 15+col, "Budget", budget, col*2)
	}
	fpdf.Ln(12)

	// ── Area market snapshot ─────────────────────────────────────────────────
	if len(d.AreaSummary) > 0 {
		sectionHeader(fpdf, "AREA MARKET SNAPSHOT — "+d.Area, pageW)
		renderKeyStats(fpdf, d.AreaSummary, pageW)
		fpdf.Ln(6)
	}

	// ── Comparable transactions ──────────────────────────────────────────────
	if len(d.Comparables) > 0 {
		sectionHeader(fpdf, "COMPARABLE TRANSACTIONS (DLD DATA)", pageW)
		renderTransactionsTable(fpdf, d.Comparables, pageW)
		fpdf.Ln(6)
	}

	// ── Rental yield ─────────────────────────────────────────────────────────
	if len(d.YieldData) > 0 {
		sectionHeader(fpdf, "RENTAL YIELD ANALYSIS", pageW)
		renderYieldTable(fpdf, d.YieldData, pageW)
		fpdf.Ln(6)
	}

	// ── Nearby amenities ─────────────────────────────────────────────────────
	if len(d.POIs) > 0 {
		sectionHeader(fpdf, "NEARBY AMENITIES", pageW)
		renderPOIs(fpdf, d.POIs, pageW)
		fpdf.Ln(6)
	}

	// ── AI property description ───────────────────────────────────────────────
	if d.AIDescription != "" {
		sectionHeader(fpdf, "PROPERTY DESCRIPTION", pageW)
		fpdf.SetFont("Helvetica", "", 9)
		fpdf.SetTextColor(textR, textG, textB)
		fpdf.MultiCell(pageW, 5, d.AIDescription, "", "L", false)
		fpdf.Ln(6)
	}

	// ── Footer ───────────────────────────────────────────────────────────────
	renderFooter(fpdf, d)

	var buf bytes.Buffer
	if err := fpdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate property report pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func sectionHeader(fpdf *gofpdf.Fpdf, title string, pageW float64) {
	fpdf.SetFillColor(headerR, headerG, headerB)
	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 9)
	fpdf.CellFormat(pageW, 7, "  "+title, "", 1, "L", true, 0, "")
	// gold underline
	fpdf.SetFillColor(accentR, accentG, accentB)
	fpdf.Rect(15, fpdf.GetY(), pageW, 0.8, "F")
	fpdf.SetTextColor(textR, textG, textB)
	fpdf.Ln(3)
}

func briefRow(fpdf *gofpdf.Fpdf, x float64, label, value string, w float64) {
	y := fpdf.GetY()
	fpdf.SetXY(x, y)
	fpdf.SetFont("Helvetica", "B", 8)
	fpdf.SetTextColor(100, 100, 120)
	fpdf.CellFormat(w, 5, label, "", 0, "L", false, 0, "")
	fpdf.SetXY(x, y+5)
	fpdf.SetFont("Helvetica", "", 9)
	fpdf.SetTextColor(textR, textG, textB)
	fpdf.CellFormat(w, 5, value, "", 0, "L", false, 0, "")
}

func renderKeyStats(fpdf *gofpdf.Fpdf, summary map[string]interface{}, pageW float64) {
	keys := []struct{ label, key string }{
		{"Avg Price (AED/sqft)", "avg_price_sqft"},
		{"Median Price (AED)", "median_price"},
		{"Transactions", "transaction_count"},
		{"YoY Change", "yoy_change"},
	}

	colW := pageW / float64(len(keys))
	startY := fpdf.GetY()

	for i, k := range keys {
		x := 15 + float64(i)*colW
		val := statValue(summary, k.key)
		if val == "" {
			continue
		}
		fpdf.SetFillColor(lightR, lightG, lightB)
		fpdf.RoundedRect(x+1, startY, colW-2, 16, 1.5, "1234", "F")
		fpdf.SetXY(x+1, startY+2)
		fpdf.SetFont("Helvetica", "B", 11)
		fpdf.SetTextColor(headerR, headerG, headerB)
		fpdf.CellFormat(colW-2, 7, val, "", 0, "C", false, 0, "")
		fpdf.SetXY(x+1, startY+9)
		fpdf.SetFont("Helvetica", "", 7)
		fpdf.SetTextColor(100, 100, 120)
		fpdf.CellFormat(colW-2, 5, k.label, "", 0, "C", false, 0, "")
	}
	fpdf.SetTextColor(textR, textG, textB)
	fpdf.SetY(startY + 20)
}

func renderTransactionsTable(fpdf *gofpdf.Fpdf, rows []map[string]interface{}, pageW float64) {
	cols := []struct {
		header string
		w      float64
		key    string
	}{
		{"Date", 28, "transaction_date"},
		{"Property", 55, "property_name"},
		{"Type", 25, "property_type"},
		{"Size (sqft)", 30, "area_sqft"},
		{"Price (AED)", 42, "price"},
	}

	// Header row
	fpdf.SetFillColor(headerR, headerG, headerB)
	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 8)
	for _, col := range cols {
		fpdf.CellFormat(col.w, 6, " "+col.header, "0", 0, "L", true, 0, "")
	}
	fpdf.Ln(-1)

	limit := len(rows)
	if limit > 10 {
		limit = 10
	}

	for i, row := range rows[:limit] {
		if fpdf.GetY() > 260 {
			fpdf.AddPage()
		}
		fill := i%2 == 0
		if fill {
			fpdf.SetFillColor(lightR, lightG, lightB)
		} else {
			fpdf.SetFillColor(255, 255, 255)
		}
		fpdf.SetTextColor(textR, textG, textB)
		fpdf.SetFont("Helvetica", "", 8)
		for _, col := range cols {
			val := safeStr(row, col.key)
			fpdf.CellFormat(col.w, 6, " "+val, "0", 0, "L", fill, 0, "")
		}
		fpdf.Ln(-1)
	}

	if len(rows) > 10 {
		fpdf.SetFont("Helvetica", "I", 7)
		fpdf.SetTextColor(120, 120, 120)
		fpdf.CellFormat(pageW, 5, fmt.Sprintf("  Showing 10 of %d transactions. Source: Dubai Land Department (DLD)", len(rows)), "", 1, "L", false, 0, "")
	} else {
		fpdf.SetFont("Helvetica", "I", 7)
		fpdf.SetTextColor(120, 120, 120)
		fpdf.CellFormat(pageW, 5, "  Source: Dubai Land Department (DLD)", "", 1, "L", false, 0, "")
	}
	fpdf.SetTextColor(textR, textG, textB)
}

func renderYieldTable(fpdf *gofpdf.Fpdf, rows []map[string]interface{}, pageW float64) {
	cols := []struct {
		header string
		w      float64
		key    string
	}{
		{"Property Type", 45, "property_type"},
		{"Bedrooms", 35, "rooms"},
		{"Avg Annual Rent (AED)", 50, "avg_annual_rent"},
		{"Avg Sale Price (AED)", 50, "avg_sale_price"},
		{"Est. Yield", 30, "yield_pct"},
	}

	fpdf.SetFillColor(headerR, headerG, headerB)
	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 8)
	for _, col := range cols {
		fpdf.CellFormat(col.w, 6, " "+col.header, "0", 0, "L", true, 0, "")
	}
	fpdf.Ln(-1)

	limit := len(rows)
	if limit > 8 {
		limit = 8
	}
	for i, row := range rows[:limit] {
		fill := i%2 == 0
		if fill {
			fpdf.SetFillColor(lightR, lightG, lightB)
		} else {
			fpdf.SetFillColor(255, 255, 255)
		}
		fpdf.SetTextColor(textR, textG, textB)
		fpdf.SetFont("Helvetica", "", 8)
		for _, col := range cols {
			val := safeStr(row, col.key)
			fpdf.CellFormat(col.w, 6, " "+val, "0", 0, "L", fill, 0, "")
		}
		fpdf.Ln(-1)
	}
	fpdf.SetFont("Helvetica", "I", 7)
	fpdf.SetTextColor(120, 120, 120)
	fpdf.CellFormat(pageW, 5, "  Source: Ejari rental registry via DLDAPI", "", 1, "L", false, 0, "")
	fpdf.SetTextColor(textR, textG, textB)
}

func renderPOIs(fpdf *gofpdf.Fpdf, pois []map[string]interface{}, pageW float64) {
	limit := len(pois)
	if limit > 6 {
		limit = 6
	}

	colW := pageW / 2
	for i, poi := range pois[:limit] {
		x := 15.0
		if i%2 == 1 {
			x = 15 + colW
		}
		y := fpdf.GetY()
		if i%2 == 0 && i > 0 {
			fpdf.Ln(8)
			y = fpdf.GetY()
		}
		if i%2 == 1 {
			fpdf.SetY(y - 8) // same row as previous
		}
		fpdf.SetXY(x, y)

		category := safeStr(poi, "category")
		name := safeStr(poi, "name")
		dist := safeStr(poi, "distance_km")

		icon := poiIcon(category)
		fpdf.SetFont("Helvetica", "B", 9)
		fpdf.SetTextColor(accentR, accentG, accentB)
		fpdf.CellFormat(8, 6, icon, "", 0, "C", false, 0, "")
		fpdf.SetFont("Helvetica", "", 9)
		fpdf.SetTextColor(textR, textG, textB)
		label := name
		if dist != "" {
			label += "  ·  " + dist + " km"
		}
		fpdf.CellFormat(colW-10, 6, label, "", 0, "L", false, 0, "")
		if i%2 == 0 {
			fpdf.Ln(-1)
		}
	}
	fpdf.Ln(4)
}

func renderFooter(fpdf *gofpdf.Fpdf, d PropertyReportData) {
	// gold line
	fpdf.SetFillColor(accentR, accentG, accentB)
	fpdf.Rect(15, 272, 180, 0.8, "F")

	fpdf.SetFont("Helvetica", "I", 7)
	fpdf.SetTextColor(130, 130, 130)
	fpdf.SetXY(15, 274)
	fpdf.MultiCell(180, 4,
		"This report is prepared for informational purposes only. Market data sourced from the Dubai Land Department "+
			"(DLD) via DLDAPI. All figures are indicative and subject to change. "+
			d.CompanyName+" does not guarantee accuracy. Generated on "+d.GeneratedAt.Format("02 Jan 2006 15:04")+".",
		"", "L", false)
}

// ── Utility ──────────────────────────────────────────────────────────────────

func safeStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func statValue(m map[string]interface{}, key string) string {
	v := safeStr(m, key)
	if v == "" || v == "0" || v == "<nil>" {
		return ""
	}
	return v
}

func formatAED(v float64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", v/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.0fK", v/1_000)
	}
	return fmt.Sprintf("%.0f", v)
}

func poiIcon(category string) string {
	switch category {
	case "metro", "transit":
		return "M"
	case "school", "education":
		return "S"
	case "mall", "retail":
		return "C"
	case "hospital", "health":
		return "H"
	case "mosque":
		return "O"
	default:
		return "*"
	}
}
