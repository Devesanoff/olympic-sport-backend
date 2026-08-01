package pdf

import (
	"bytes"
	"fmt"
	"image/png"
	"strconv"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

// Badge dimensions (CR80 standard vertical)
const (
	BadgeWidth  = 54.0 // mm
	BadgeHeight = 86.0 // mm
)

// BadgeBuilder handles constructing a multi-page PDF of badges.
type BadgeBuilder struct {
	pdf *gofpdf.Fpdf
}

// NewBadgeBuilder initializes a new PDF document.
func NewBadgeBuilder() *BadgeBuilder {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: BadgeWidth, Ht: BadgeHeight},
	})
	return &BadgeBuilder{pdf: pdf}
}

// AddBadgePage generates a single badge page.
func (b *BadgeBuilder) AddBadgePage(fullName, categoryName, colorHex, qrToken string) error {
	b.pdf.AddPage()

	// 1. Parse Color Hex to RGB
	r, g, bl := parseHexColor(colorHex)

	// 2. Draw Top Color Header
	b.pdf.SetFillColor(r, g, bl)
	b.pdf.Rect(0, 0, BadgeWidth, 15, "F")

	b.pdf.SetFont("Arial", "B", 12)
	b.pdf.SetTextColor(255, 255, 255)
	b.pdf.SetY(5)
	b.pdf.CellFormat(BadgeWidth, 5, categoryName, "", 0, "C", false, 0, "")

	// 3. Draw Photo Placeholder
	b.pdf.SetFillColor(200, 200, 200) // Gray background
	b.pdf.Rect(17, 20, 20, 25, "F")
	b.pdf.SetFont("Arial", "", 8)
	b.pdf.SetTextColor(100, 100, 100)
	b.pdf.SetXY(17, 30)
	b.pdf.CellFormat(20, 5, "PHOTO", "", 0, "C", false, 0, "")

	// 4. Draw Participant Name
	b.pdf.SetFont("Arial", "B", 10)
	b.pdf.SetTextColor(0, 0, 0)
	b.pdf.SetXY(0, 50)
	b.pdf.CellFormat(BadgeWidth, 5, fullName, "", 0, "C", false, 0, "")

	// 5. Generate and Embed QR Code
	qrCode, err := qrcode.New(qrToken, qrcode.Medium)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}
	qrImage := qrCode.Image(256)

	var buf bytes.Buffer
	if err := png.Encode(&buf, qrImage); err != nil {
		return fmt.Errorf("failed to encode QR png: %w", err)
	}

	opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	// Register image dynamically with a unique name based on token length/hash to avoid collisions if multiple badges share token
	// Or simply use a standard name since we overwrite or it's per page. Actually, token should be unique.
	imgName := "qr_" + qrToken
	b.pdf.RegisterImageOptionsReader(imgName, opt, &buf)

	b.pdf.ImageOptions(imgName, 17, 60, 20, 20, false, opt, 0, "")

	return nil
}

// Output returns the PDF as a byte slice.
func (b *BadgeBuilder) Output() ([]byte, error) {
	var buf bytes.Buffer
	if err := b.pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// parseHexColor converts a hex string (e.g., #FF5733 or FF5733) to r, g, b ints.
func parseHexColor(s string) (int, int, int) {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	if len(s) == 6 {
		c, err := strconv.ParseUint(s, 16, 32)
		if err == nil {
			return int(c >> 16), int((c >> 8) & 0xFF), int(c & 0xFF)
		}
	}
	// Default to dark gray if invalid
	return 50, 50, 50
}
