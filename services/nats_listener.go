package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

type NATSListener struct {
	nc         *nats.Conn
	pdfService *PDFService
	minio      *MinioService
	bucket     string
}

func NewNATSListener(pdfService *PDFService, minio *MinioService, bucket string) (*NATSListener, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSListener{
		nc:         nc,
		pdfService: pdfService,
		minio:      minio,
		bucket:     bucket,
	}, nil
}

func (l *NATSListener) Listen() {
	// Listen for settled orders to generate invoices
	l.nc.Subscribe("marketplace.order.settled", func(m *nats.Msg) {
		var orderData map[string]interface{}
		if err := json.Unmarshal(m.Data, &orderData); err != nil {
			log.Printf("❌ Failed to unmarshal order data: %v", err)
			return
		}

		orderNumber, _ := orderData["order_number"].(string)
		log.Printf("🧾 Order Settled: %s. Generating Invoice...", orderNumber)

		// 1. Generate PDF
		pdfContent, err := l.pdfService.GenerateInvoice(orderData)
		if err != nil {
			log.Printf("❌ Failed to generate invoice for %s: %v", orderNumber, err)
			return
		}

		// 2. Upload to MinIO
		objectName := fmt.Sprintf("invoices/%s.pdf", orderNumber)
		reader := bytes.NewReader(pdfContent)
		size := int64(len(pdfContent))

		err = l.minio.UploadFile(context.Background(), l.bucket, objectName, reader, size, "application/pdf")
		if err != nil {
			log.Printf("❌ Failed to upload invoice for %s to MinIO: %v", orderNumber, err)
			return
		}
		log.Printf("✅ Invoice archived successfully: %s", objectName)

		// 3. Emit "Invoice Ready" event for Notification Service
		eventPayload, _ := json.Marshal(map[string]interface{}{
			"order_number": orderNumber,
			"object_name":  objectName,
			"bucket":       l.bucket,
			"buyer_id":     orderData["operator_id"],
			"buyer_name":   orderData["operator_name"],
			"timestamp":    time.Now(),
		})
		l.nc.Publish("document.invoice.ready", eventPayload)
	})
}
