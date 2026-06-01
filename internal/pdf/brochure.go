package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// BrochureData holds everything needed to render a listing brochure PDF.
type BrochureData struct {
	// Company branding
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string
	LogoURL        string
	PrimaryColor   string // hex e.g. "#1a3a5c" — parsed to RGB
	Disclaimer     string

	// Agent
	AgentName  string
	AgentPhone string
	AgentEmail string

	// Listing
	ListingID       string
	Title           string
	ReferenceNumber string
	PropertyType    string
	ListingType     string // sale | rent
	Price           float64
	Currency        string
	RentPeriod      string
	Bedrooms        int
	Bathrooms       int
	TotalSqft       float64
	ParkingSpaces   int
	Furnishing      string
	YearBuilt       int
	Area            string
	Community       string
	City            string
	Emirate         string
	Description     string
	Amenities       []string
	CoverImageURL   string
	AvailableFrom   *time.Time

	GeneratedAt time.Time
}

// GenerateBrochure renders a single-page A4 listing brochure as PDF bytes.
func GenerateBrochure(d BrochureData) ([]byte, error) {
	if d.Currency == "" {
		d.Currency = "AED"
	}
	if d.GeneratedAt.IsZero() {
		d.GeneratedAt = time.Now()
	}

	// Parse primary color (default: navy)
	pr, pg, pb := 26, 58, 92
	if len(d.PrimaryColor) == 7 && d.PrimaryColor[0] == '#' {
		fmt.Sscanf(d.PrimaryColor[1:], "%02x%02x%02x", &pr, &pg, &pb)
	}

	fpdf := gofpdf.New("P", "mm", "A4", "")
	fpdf.SetMargins(14, 14, 14)
	fpdf.SetAutoPageBreak(true, 14)
	fpdf.AddPage()

	pageW, _ := fpdf.GetPageSize()
	contentW := pageW - 28

	// ── Header bar ────────────────────────────────────────────────────────
	fpdf.SetFillColor(pr, pg, pb)
	fpdf.Rect(0, 0, pageW, 22, "F")
	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 13)
	fpdf.SetXY(14, 7)
	fpdf.CellFormat(contentW/2, 8, d.CompanyName, "", 0, "L", false, 0, "")
	fpdf.SetFont("Helvetica", "", 9)
	fpdf.SetXY(pageW/2, 7)
	fpdf.CellFormat(contentW/2, 8, d.CompanyPhone+" | "+d.CompanyEmail, "", 0, "R", false, 0, "")

	// ── Cover image area (placeholder box if no image) ────────────────────
	fpdf.SetXY(14, 26)
	imgH := 60.0
	if d.CoverImageURL != "" {
		// gofpdf can embed images from URLs via HTTP
		// We use a placeholder grey rect since gofpdf needs local file paths
		// The handler will embed the image bytes separately if needed
		fpdf.SetFillColor(220, 225, 232)
		fpdf.Rect(14, 26, contentW, imgH, "F")
		fpdf.SetTextColor(150, 155, 165)
		fpdf.SetFont("Helvetica", "I", 9)
		fpdf.SetXY(14, 26+imgH/2-4)
		fpdf.CellFormat(contentW, 8, d.CoverImageURL, "", 0, "C", false, 0, "")
	} else {
		fpdf.SetFillColor(220, 225, 232)
		fpdf.Rect(14, 26, contentW, imgH, "F")
		fpdf.SetTextColor(160, 165, 170)
		fpdf.SetFont("Helvetica", "I", 9)
		fpdf.SetXY(14, 26+imgH/2-4)
		fpdf.CellFormat(contentW, 8, "No image", "", 0, "C", false, 0, "")
	}

	y := 26 + imgH + 6

	// ── Status ribbon ─────────────────────────────────────────────────────
	accentR, accentG, accentB := 180, 140, 60
	fpdf.SetFillColor(accentR, accentG, accentB)
	fpdf.Rect(14, y, 40, 7, "F")
	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 8)
	fpdf.SetXY(14, y+0.5)
	listing := strings.ToUpper(d.ListingType)
	if d.ListingType == "rent" {
		listing = "FOR RENT"
	} else {
		listing = "FOR SALE"
	}
	fpdf.CellFormat(40, 6, listing, "", 0, "C", false, 0, "")
	y += 10

	// ── Title + price ──────────────────────────────────────────────────────
	fpdf.SetTextColor(pr, pg, pb)
	fpdf.SetFont("Helvetica", "B", 16)
	fpdf.SetXY(14, y)
	fpdf.MultiCell(contentW*0.65, 7, d.Title, "", "L", false)
	y2 := fpdf.GetY()

	// Price block (right-aligned)
	fpdf.SetFont("Helvetica", "B", 18)
	fpdf.SetTextColor(accentR, accentG, accentB)
	priceStr := fmt.Sprintf("%s %s", d.Currency, formatNum(d.Price))
	if d.RentPeriod != "" {
		priceStr += " / " + d.RentPeriod
	}
	fpdf.SetXY(pageW*0.62, y)
	fpdf.CellFormat(pageW*0.38-14, 10, priceStr, "", 0, "R", false, 0, "")

	// Reference
	fpdf.SetFont("Helvetica", "", 8)
	fpdf.SetTextColor(130, 130, 130)
	fpdf.SetXY(pageW*0.62, y+11)
	if d.ReferenceNumber != "" {
		fpdf.CellFormat(pageW*0.38-14, 5, "Ref: "+d.ReferenceNumber, "", 0, "R", false, 0, "")
	}

	y = y2 + 4

	// ── Key specs row ──────────────────────────────────────────────────────
	fpdf.SetFillColor(245, 247, 250)
	fpdf.Rect(14, y, contentW, 16, "F")
	specs := []struct{ icon, val string }{
		{"🛏", fmt.Sprintf("%d Beds", d.Bedrooms)},
		{"🚿", fmt.Sprintf("%d Baths", d.Bathrooms)},
		{"📐", fmt.Sprintf("%.0f sqft", d.TotalSqft)},
		{"🚗", fmt.Sprintf("%d Parking", d.ParkingSpaces)},
		{"🏗", fmt.Sprintf("%s", d.PropertyType)},
	}
	specW := contentW / float64(len(specs))
	for i, sp := range specs {
		fpdf.SetFont("Helvetica", "B", 8)
		fpdf.SetTextColor(pr, pg, pb)
		fpdf.SetXY(14+float64(i)*specW, y+1)
		fpdf.CellFormat(specW, 6, sp.val, "", 0, "C", false, 0, "")
		fpdf.SetFont("Helvetica", "", 7)
		fpdf.SetTextColor(120, 120, 120)
		fpdf.SetXY(14+float64(i)*specW, y+8)
		fpdf.CellFormat(specW, 5, sp.icon, "", 0, "C", false, 0, "")
	}
	y += 20

	// ── Location ───────────────────────────────────────────────────────────
	location := strings.Join(filterEmpty([]string{d.Area, d.Community, d.City, d.Emirate}), ", ")
	if location != "" {
		fpdf.SetFont("Helvetica", "", 9)
		fpdf.SetTextColor(80, 80, 80)
		fpdf.SetXY(14, y)
		fpdf.CellFormat(contentW, 6, "📍  "+location, "", 1, "L", false, 0, "")
		y += 8
	}

	// ── Two-column layout: description + details ──────────────────────────
	colW := (contentW - 6) / 2

	// Description (left)
	if d.Description != "" {
		fpdf.SetFont("Helvetica", "B", 9)
		fpdf.SetTextColor(pr, pg, pb)
		fpdf.SetXY(14, y)
		fpdf.CellFormat(colW, 6, "Description", "", 1, "L", false, 0, "")

		fpdf.SetFont("Helvetica", "", 8)
		fpdf.SetTextColor(60, 60, 60)
		fpdf.SetXY(14, fpdf.GetY())

		desc := d.Description
		if len(desc) > 600 {
			desc = desc[:597] + "..."
		}
		fpdf.MultiCell(colW, 4.5, desc, "", "L", false)
	}

	// Details (right)
	detailY := y
	fpdf.SetFont("Helvetica", "B", 9)
	fpdf.SetTextColor(pr, pg, pb)
	fpdf.SetXY(14+colW+6, detailY)
	fpdf.CellFormat(colW, 6, "Property Details", "", 1, "L", false, 0, "")
	detailY += 7

	details := []struct{ k, v string }{
		{"Type", d.PropertyType},
		{"Furnishing", d.Furnishing},
		{"Year Built", ifStr(d.YearBuilt > 0, fmt.Sprintf("%d", d.YearBuilt))},
		{"Available", ifTime(d.AvailableFrom)},
	}
	for _, row := range details {
		if row.v == "" {
			continue
		}
		fpdf.SetFont("Helvetica", "", 8)
		fpdf.SetTextColor(100, 100, 100)
		fpdf.SetXY(14+colW+6, detailY)
		fpdf.CellFormat(colW*0.45, 5, row.k+":", "", 0, "L", false, 0, "")
		fpdf.SetTextColor(40, 40, 40)
		fpdf.SetFont("Helvetica", "B", 8)
		fpdf.CellFormat(colW*0.55, 5, row.v, "", 1, "L", false, 0, "")
		detailY += 5.5
	}

	// Amenities
	if len(d.Amenities) > 0 {
		bottom := fpdf.GetY()
		if detailY > bottom {
			bottom = detailY
		}
		bottom += 6

		fpdf.SetFont("Helvetica", "B", 9)
		fpdf.SetTextColor(pr, pg, pb)
		fpdf.SetXY(14, bottom)
		fpdf.CellFormat(contentW, 6, "Amenities", "", 1, "L", false, 0, "")
		bottom = fpdf.GetY()

		fpdf.SetFont("Helvetica", "", 7.5)
		fpdf.SetTextColor(60, 60, 60)
		amenityW := contentW / 3
		for i, a := range d.Amenities {
			col := i % 3
			row := i / 3
			fpdf.SetXY(14+float64(col)*amenityW, bottom+float64(row)*5)
			fpdf.CellFormat(amenityW, 5, "• "+a, "", 0, "L", false, 0, "")
		}
	}

	// ── Agent card ─────────────────────────────────────────────────────────
	pageH := 297.0
	agentY := pageH - 38
	fpdf.SetFillColor(pr, pg, pb)
	fpdf.Rect(0, agentY, pageW, 38, "F")
	fpdf.SetTextColor(255, 255, 255)
	fpdf.SetFont("Helvetica", "B", 10)
	fpdf.SetXY(14, agentY+6)
	fpdf.CellFormat(contentW/2, 7, d.AgentName, "", 1, "L", false, 0, "")
	fpdf.SetFont("Helvetica", "", 8)
	fpdf.SetXY(14, agentY+14)
	fpdf.CellFormat(contentW/2, 5, d.AgentPhone, "", 1, "L", false, 0, "")
	fpdf.SetXY(14, agentY+20)
	fpdf.CellFormat(contentW/2, 5, d.AgentEmail, "", 1, "L", false, 0, "")

	fpdf.SetFont("Helvetica", "I", 7)
	fpdf.SetTextColor(200, 210, 220)
	fpdf.SetXY(pageW/2, agentY+6)
	if d.Disclaimer != "" {
		fpdf.MultiCell(contentW/2, 4, d.Disclaimer, "", "R", false)
	}
	fpdf.SetXY(pageW/2, agentY+26)
	fpdf.CellFormat(contentW/2, 4, "Generated "+d.GeneratedAt.Format("Jan 2006"), "", 0, "R", false, 0, "")

	var buf bytes.Buffer
	if err := fpdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("brochure pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func formatNum(n float64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.2fM", n/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.0f", n)
}

func filterEmpty(ss []string) []string {
	out := ss[:0]
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func ifStr(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}

func ifTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("Jan 2006")
}
