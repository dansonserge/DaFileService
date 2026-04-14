package services

import (
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
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
	items, _ := data["items"].([]interface{})

	m := maroto.New()

	// 1. Header & Branding
	m.AddRows(
		row.New(20).Add(
			col.New(6).Add(
				text.New("HEALTHCHAIN", props.Text{
					Size:  18,
					Style: fontstyle.Bold,
					Align: align.Left,
				}),
				text.New("Clinical Marketplace Ledger", props.Text{
					Size:  8,
					Top:   8,
					Color: &props.Color{Red: 100, Green: 100, Blue: 100},
				}),
			),
			col.New(6).Add(
				text.New("INVOICE", props.Text{
					Size:  24,
					Style: fontstyle.Bold,
					Align: align.Right,
				}),
				text.New(fmt.Sprintf("REF: %s", orderNumber), props.Text{
					Size:  9,
					Top:   10,
					Align: align.Right,
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(10).Add(col.New(12).Add(line.New())),
	)

	// 2. Transaction Participants
	m.AddRows(
		row.New(15).Add(
			col.New(6).Add(
				text.New("PURCHASING INSTITUTION", props.Text{Size: 8, Style: fontstyle.Bold, Color: &props.Color{Red: 150, Green: 150, Blue: 150}}),
				text.New(operatorName, props.Text{Size: 10, Style: fontstyle.Bold, Top: 4}),
			),
			col.New(6).Add(
				text.New("DATE OF SETTLEMENT", props.Text{Size: 8, Style: fontstyle.Bold, Color: &props.Color{Red: 150, Green: 150, Blue: 150}, Align: align.Right}),
				text.New(time.Now().Format("Jan 02, 2006 15:04:05 MST"), props.Text{Size: 10, Style: fontstyle.Bold, Top: 4, Align: align.Right}),
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
			col.New(3).Add(text.New("TOTAL (RWF)", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Right})),
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
				text.New(fmt.Sprintf("TOTAL AMOUNT: %.2f RWF", grandTotal), props.Text{
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
				text.New("VERIFICATION", props.Text{Size: 7, Style: fontstyle.Bold, Color: &props.Color{Red: 200, Green: 200, Blue: 200}}),
				text.New(fmt.Sprintf("HC-AUTH-%s", orderNumber), props.Text{Size: 6, Top: 3, Color: &props.Color{Red: 180, Green: 180, Blue: 180}}),
				text.New("This document is a certified digital record of the HealthChain Marketplace. It serves as a handover and financial settlement instrument.", props.Text{
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
