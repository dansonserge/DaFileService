package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
	minio          *MinioService
	httpClient     *http.Client
	authServiceURL string
}

func NewPDFService(minio *MinioService, authURL string) *PDFService {
	return &PDFService{
		minio:          minio,
		authServiceURL: authURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// getImageBytes helper to fetch remote logos
func (s *PDFService) getImageBytes(url string) ([]byte, extension.Type) {
	if url == "" {
		return nil, ""
	}

	fullURL := url
	if !strings.HasPrefix(url, "http") {
		// Resolve relative path similar to frontend
		if strings.HasPrefix(url, "/api/auth/v1/") {
			fullURL = s.authServiceURL + url
		} else if strings.HasPrefix(url, "/clinical-vault/") {
			fullURL = s.authServiceURL + "/api/auth/v1" + url
		} else if strings.HasPrefix(url, "/") {
			fullURL = s.authServiceURL + url
		} else {
			fullURL = s.authServiceURL + "/" + url
		}
	}

	resp, err := s.httpClient.Get(fullURL)
	if err != nil {
		fmt.Printf("⚠️ Remote Logo Fetch Failed (%s): %v\n", fullURL, err)
		return nil, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("⚠️ Remote Logo Fetch HTTP Error (%s): %d\n", fullURL, resp.StatusCode)
		return nil, ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ""
	}

	ext := extension.Jpg
	if strings.HasSuffix(strings.ToLower(url), ".png") {
		ext = extension.Png
	}

	return data, ext
}

// GenerateInvoice creates a high-density, professional PDF invoice matching User Design
func (s *PDFService) GenerateInvoice(data map[string]interface{}) ([]byte, error) {
	orderNumber, _ := data["order_number"].(string)
	payerName, _ := data["payer_name"].(string)
	buyerName, _ := data["buyer_name"].(string)
	buyerLogo, _ := data["buyer_logo"].(string)
	sellerName, _ := data["seller_name"].(string)
	sellerLogo, _ := data["seller_logo"].(string)
	items, _ := data["items"].([]interface{})

	m := maroto.New()

	// Colors matching the design
	gray := &props.Color{Red: 150, Green: 150, Blue: 150}
	blue := &props.Color{Red: 0, Green: 122, Blue: 255}

	// 1. Branding Header
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New("HEALTHCHAIN", props.Text{
					Size:  24,
					Style: fontstyle.Bold,
					Align: align.Center,
				}),
			),
		),
		row.New(10).Add(
			col.New(12).Add(
				text.New("Clinical Marketplace Ledger", props.Text{
					Size:  10,
					Align: align.Center,
					Color: gray,
				}),
			),
		),
		row.New(10).Add(col.New(12).Add(line.New(props.Line{Thickness: 0.2}))),
	)

	// 2. Document Title & Metadata
	m.AddRows(
		row.New(20).Add(
			col.New(8).Add(
				text.New("INVOICE / HANDOVER CERTIFICATE", props.Text{
					Size:  15,
					Style: fontstyle.Bold,
					Top:   5,
				}),
			),
			col.New(4).Add(
				text.New(fmt.Sprintf("REF: %s", orderNumber), props.Text{
					Size:  12,
					Style: fontstyle.Bold,
					Align: align.Right,
				}),
				text.New(fmt.Sprintf("DATE: %s", time.Now().Format("02 Jan 2006")), props.Text{
					Size:  9,
					Top:   6,
					Align: align.Right,
				}),
			),
		),
		row.New(10),
	)

	// 3. Participants (Seller & Buyer)
	// Fetch Logos
	sellerLogoBytes, sellerExt := s.getImageBytes(sellerLogo)
	buyerLogoBytes, buyerExt := s.getImageBytes(buyerLogo)

	m.AddRows(
		row.New(30).Add(
			// Seller Section
			col.New(2).Add(func() core.Component {
				if sellerLogoBytes != nil {
					return image.NewFromBytes(sellerLogoBytes, sellerExt, props.Rect{Percent: 80, Center: true})
				}
				return text.New("LOGO", props.Text{Size: 6, Color: gray, Align: align.Center, Top: 10})
			}()),
			col.New(4).Add(
				text.New("SELLER / SUPPLIER", props.Text{Size: 8, Color: gray, Style: fontstyle.Bold}),
				text.New(sellerName, props.Text{Size: 12, Style: fontstyle.Bold, Top: 4}),
				text.New("Verified HealthChain Supplier", props.Text{Size: 8, Color: gray, Top: 8}),
			),

			// Buyer Section
			col.New(2).Add(func() core.Component {
				if buyerLogoBytes != nil {
					return image.NewFromBytes(buyerLogoBytes, buyerExt, props.Rect{Percent: 80, Center: true})
				}
				return text.New("LOGO", props.Text{Size: 6, Color: gray, Align: align.Center, Top: 10})
			}()),
			col.New(4).Add(
				text.New("BUYER / INSTITUTION", props.Text{Size: 8, Color: gray, Style: fontstyle.Bold}),
				text.New(buyerName, props.Text{Size: 12, Style: fontstyle.Bold, Top: 4}),
				text.New(fmt.Sprintf("Purchased by: %s", payerName), props.Text{
					Size:  9,
					Style: fontstyle.Bold,
					Top:   8,
					Color: blue,
				}),
			),
		),
		row.New(10),
	)

	// 4. Line Items Table Header
	m.AddRows(
		row.New(10).Add(
			col.New(5).Add(text.New("PRODUCT DESCRIPTION", props.Text{Size: 9, Style: fontstyle.Bold})),
			col.New(3).Add(text.New("BATCH / EXPIRY", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center})),
			col.New(2).Add(text.New("QTY", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center})),
			col.New(2).Add(text.New("TOTAL (RWF)", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Right})),
		),
		row.New(5).Add(col.New(12).Add(line.New(props.Line{Thickness: 0.1}))),
	)

	// 5. Populate Items
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
			} else {
				// Try alternative format if needed
				expDate = expiry
			}
		}

		m.AddRows(
			row.New(12).Add(
				col.New(5).Add(text.New(name, props.Text{Size: 10, Style: fontstyle.Bold})),
				col.New(3).Add(text.New(fmt.Sprintf("%s | %s", batch, expDate), props.Text{Size: 9, Align: align.Center})),
				col.New(2).Add(text.New(fmt.Sprintf("%.0f Units", qty), props.Text{Size: 9, Align: align.Center})),
				col.New(2).Add(text.New(fmt.Sprintf("%.2f", total), props.Text{Size: 10, Align: align.Right, Style: fontstyle.Bold})),
			),
		)
	}

	// 6. Total Summary
	m.AddRows(
		row.New(20),
		row.New(15).Add(
			col.New(7),
			col.New(5).Add(
				line.New(props.Line{Thickness: 0.1}),
				text.New(fmt.Sprintf("TOTAL DUE: %.2f RWF", grandTotal), props.Text{
					Size:  14,
					Top:   5,
					Style: fontstyle.Bold,
					Align: align.Right,
				}),
			),
		),
	)

	// 7. Footer - Verification & QR
	m.AddRows(
		row.New(40),
		row.New(40).Add(
			col.New(8).Add(
				text.New("VERIFICATION HASH", props.Text{Size: 8, Style: fontstyle.Bold, Color: gray}),
				text.New(fmt.Sprintf("HC-AUTH-%s", orderNumber), props.Text{Size: 7, Top: 3, Color: gray}),
				text.New("This document is a certified digital record of the HealthChain Marketplace. It serves as a clinical handover and financial settlement instrument.", props.Text{
					Size:  8,
					Top:   12,
					Style: fontstyle.Italic,
					Color: gray,
				}),
			),
			col.New(4).Add(
				code.NewQr(fmt.Sprintf("https://healthchain.rw/verify/%s", orderNumber), props.Rect{
					Center:  true,
					Percent: 100,
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
