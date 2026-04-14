package services

import (
	"fmt"
	"time"

	"io"
	"net/http"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type PDFService struct {
	minio *MinioService
}

func NewPDFService(minio *MinioService) *PDFService {
	return &PDFService{minio: minio}
}

// GenerateInvoice creates a high-density, professional PDF invoice
func (s *PDFService) GenerateInvoice(data map[string]interface{}) ([]byte, error) {
	orderNumber, _ := data["order_number"].(string)
	operatorName, _ := data["operator_name"].(string)
	buyerName, _ := data["buyer_name"].(string)
	sellerName, _ := data["seller_name"].(string)
	buyerLogo, _ := data["buyer_logo"].(string)
	sellerLogo, _ := data["seller_logo"].(string)
	currencyCode, _ := data["currency"].(string)
	if currencyCode == "" { currencyCode = "RWF" }

	items, _ := data["items"].([]interface{})

	m := maroto.New()

	// 0. Download Logos if available
	var bLogoBytes, sLogoBytes []byte

	// 🛠️ Internal URL Bridge: Resolve relative paths for clinical-vault
	resolveLogo := func(path string) string {
		if path == "" { return "" }
		if strings.HasPrefix(path, "http") { return path }
		
		// Map relative clinical-vault paths to internal service
		if strings.HasPrefix(path, "/api/auth") {
			return "http://auth-service:8080" + path
		}
		if strings.HasPrefix(path, "/api/file") {
			return "http://file-service:8080" + path
		}
		return path
	}

	if bUrl := resolveLogo(buyerLogo); bUrl != "" {
		if resp, err := http.Get(bUrl); err == nil && resp.StatusCode == 200 {
			bLogoBytes, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
	if sUrl := resolveLogo(sellerLogo); sUrl != "" {
		if resp, err := http.Get(sUrl); err == nil && resp.StatusCode == 200 {
			sLogoBytes, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	// 1. Header & Branding
	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New("HEALTHCHAIN", props.Text{
					Size:  18,
					Style: fontstyle.Bold,
					Align: align.Center,
				}),
				text.New("Clinical Marketplace Ledger", props.Text{
					Size:  8,
					Top:   8,
					Align: align.Center,
					Color: &props.Color{Red: 100, Green: 100, Blue: 100},
				}),
			),
		),
		row.New(10).Add(col.New(12).Add(line.New())),
	)

	// 1.1 Document Header Details
	m.AddRows(
		row.New(15).Add(
			col.New(8).Add(
				text.New("INVOICE / HANDOVER CERTIFICATE", props.Text{
					Size:  14,
					Style: fontstyle.Bold,
					Align: align.Left,
				}),
			),
			col.New(4).Add(
				text.New(fmt.Sprintf("REF: %s", orderNumber), props.Text{
					Size:  10,
					Align: align.Right,
					Style: fontstyle.Bold,
				}),
				text.New(fmt.Sprintf("DATE: %s", time.Now().Format("02 Jan 2006")), props.Text{
					Size:  8,
					Top:   6,
					Align: align.Right,
				}),
			),
		),
		row.New(10),
	)

	// 2. Transaction Participants
	m.AddRows(
		row.New(20).Add(
			// SELLER BLOCK
			col.New(1).Add(
				func() core.Component {
					if len(sLogoBytes) > 0 {
						ext := extension.Jpg
						if strings.HasSuffix(strings.ToLower(sellerLogo), ".png") { ext = extension.Png }
						return image.NewFromBytes(sLogoBytes, ext, props.Rect{Center: true, Percent: 100})
					}
					return text.New("LOGO", props.Text{Size: 5, Top: 8, Color: &props.Color{Red: 220, Green: 220, Blue: 220}})
				}(),
			),
			col.New(5).Add(
				text.New("SELLER / SUPPLIER", props.Text{Size: 7, Style: fontstyle.Bold, Color: &props.Color{Red: 150, Green: 150, Blue: 150}}),
				text.New(sellerName, props.Text{Size: 10, Style: fontstyle.Bold, Top: 4}),
				text.New("Verified HealthChain Supplier", props.Text{Size: 7, Top: 9, Color: &props.Color{Red: 100, Green: 100, Blue: 100}}),
			),

			// BUYER BLOCK
			col.New(1).Add(
				func() core.Component {
					if len(bLogoBytes) > 0 {
						ext := extension.Jpg
						if strings.HasSuffix(strings.ToLower(buyerLogo), ".png") { ext = extension.Png }
						return image.NewFromBytes(bLogoBytes, ext, props.Rect{Center: true, Percent: 100})
					}
					return text.New("LOGO", props.Text{Size: 5, Top: 8, Color: &props.Color{Red: 220, Green: 220, Blue: 220}})
				}(),
			),
			col.New(5).Add(
				text.New("BUYER / INSTITUTION", props.Text{Size: 7, Style: fontstyle.Bold, Color: &props.Color{Red: 150, Green: 150, Blue: 150}}),
				text.New(buyerName, props.Text{Size: 10, Style: fontstyle.Bold, Top: 4}),
				text.New(fmt.Sprintf("Purchased by: %s", operatorName), props.Text{Size: 8, Top: 9, Color: &props.Color{Red: 0, Green: 100, Blue: 200}}),
			),
		),
		row.New(10),
	)

	// 3. Line Items Header
	m.AddRows(
		row.New(10).Add(
			col.New(4).Add(text.New("PRODUCT DESCRIPTION", props.Text{Size: 8, Style: fontstyle.Bold})),
			col.New(3).Add(text.New("BATCH / EXPIRY", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
			col.New(2).Add(text.New("QTY", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
			col.New(3).Add(text.New(fmt.Sprintf("TOTAL (%s)", currencyCode), props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Right})),
		),
		row.New(5).Add(col.New(12).Add(line.New())),
	)

	// 4. Populate Items
	var grandTotal float64
	for _, itm := range items {
		item, ok := itm.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := item["name"].(string)
		batch, _ := item["batch_number"].(string)
		expiry, _ := item["expiry_date"].(string)
		qty, _ := item["quantity"].(float64)
		price, _ := item["price"].(float64)
		total := qty * price
		grandTotal += total

		// Format Expiry
		expDate := "N/A"
		if expiry != "" {
			if t, err := time.Parse(time.RFC3339, expiry); err == nil {
				expDate = t.Format("01/2006")
			} else if t, err := time.Parse("2006-01-02", expiry); err == nil {
				expDate = t.Format("01/2006")
			}
		}

		m.AddRows(
			row.New(12).Add(
				col.New(4).Add(text.New(name, props.Text{Size: 9, Style: fontstyle.Bold})),
				col.New(3).Add(text.New(fmt.Sprintf("%s | %s", batch, expDate), props.Text{Size: 8, Align: align.Center})),
				col.New(2).Add(text.New(fmt.Sprintf("%.0f Units", qty), props.Text{Size: 9, Align: align.Center})),
				col.New(3).Add(text.New(fmt.Sprintf("%.2f", total), props.Text{Size: 9, Align: align.Right, Style: fontstyle.Bold})),
			),
		)
	}

	// 5. Total Row
	m.AddRows(
		row.New(15),
		row.New(15).Add(
			col.New(8),
			col.New(4).Add(
				line.New(),
				text.New(fmt.Sprintf("TOTAL DUE: %.2f %s", grandTotal, currencyCode), props.Text{
					Size:  12,
					Top:   5,
					Style: fontstyle.Bold,
					Align: align.Right,
				}),
			),
		),
	)

	// 6. Verification Footer & QR
	m.AddRows(
		row.New(30),
		row.New(40).Add(
			col.New(8).Add(
				text.New("VERIFICATION HASH", props.Text{Size: 7, Style: fontstyle.Bold, Color: &props.Color{Red: 200, Green: 200, Blue: 200}}),
				text.New(fmt.Sprintf("HC-AUTH-%s", orderNumber), props.Text{Size: 6, Top: 3, Color: &props.Color{Red: 180, Green: 180, Blue: 180}}),
				text.New("This document is a certified digital record of the HealthChain Marketplace. It serves as a clinical handover and financial settlement instrument.", props.Text{
					Size:  8,
					Top:   10,
					Style: fontstyle.Italic,
					Color: &props.Color{Red: 150, Green: 150, Blue: 150},
				}),
			),
			col.New(4).Add(
				code.NewQr(fmt.Sprintf("https://healthchain.rw/verify/%s", orderNumber), props.Rect{
					Center:  true,
					Percent: 80,
				}),
			),
		),
	)

	// Render to Buffer
	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return doc.GetBytes(), nil
}
