//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type ProductMessage struct {
	ExternalID  string `json:"external_id"`
	Model       string `json:"model"`
	SKU         string `json:"sku"`
	Price       string `json:"price"`
	StockStatus string `json:"stock_status"`
	Quantity    int    `json:"quantity"`
	IsEnable    bool   `json:"is_enable"`
}

func main() {
	broker := "46.148.236.163:9092"
	topic := "store.products"

	products := []ProductMessage{
		{ExternalID: "ERP-00001", Model: "iPhone 15 Pro 256GB Black", SKU: "APPL-IP15P-256-BLK", Price: "999.99", StockStatus: "IN_STOCK", Quantity: 50, IsEnable: true},
		{ExternalID: "ERP-00002", Model: "iPhone 15 256GB White", SKU: "APPL-IP15-256-WHT", Price: "799.99", StockStatus: "IN_STOCK", Quantity: 30, IsEnable: true},
		{ExternalID: "ERP-00003", Model: "Samsung Galaxy S24 Ultra 512GB", SKU: "SAMS-S24U-512-BLK", Price: "1199.99", StockStatus: "IN_STOCK", Quantity: 20, IsEnable: true},
		{ExternalID: "ERP-00004", Model: "Samsung Galaxy A55 128GB Blue", SKU: "SAMS-A55-128-BLU", Price: "399.99", StockStatus: "IN_STOCK", Quantity: 100, IsEnable: true},
		{ExternalID: "ERP-00005", Model: "Google Pixel 9 Pro 256GB", SKU: "GOOG-P9P-256-GRY", Price: "899.99", StockStatus: "OUT_OF_STOCK", Quantity: 0, IsEnable: false},
		{ExternalID: "ERP-00006", Model: "OnePlus 12 256GB Green", SKU: "ONEP-12-256-GRN", Price: "699.99", StockStatus: "IN_STOCK", Quantity: 15, IsEnable: true},
		{ExternalID: "ERP-00007", Model: "Xiaomi 14 Pro 512GB White", SKU: "XIAO-14P-512-WHT", Price: "749.99", StockStatus: "IN_STOCK", Quantity: 25, IsEnable: true},
		{ExternalID: "ERP-00008", Model: "Sony Xperia 1 VI 256GB Black", SKU: "SONY-X1VI-256-BLK", Price: "1099.99", StockStatus: "IN_STOCK", Quantity: 10, IsEnable: true},
		{ExternalID: "ERP-00009", Model: "Motorola Edge 50 Pro 256GB", SKU: "MOTO-E50P-256-BLK", Price: "499.99", StockStatus: "LOW_STOCK", Quantity: 5, IsEnable: true},
		{ExternalID: "ERP-00010", Model: "Nothing Phone 2a 256GB White", SKU: "NOTH-P2A-256-WHT", Price: "349.99", StockStatus: "IN_STOCK", Quantity: 40, IsEnable: true},
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
	}
	defer w.Close()

	ctx := context.Background()

	for i, p := range products {
		data, err := json.Marshal(p)
		if err != nil {
			log.Fatalf("marshal error: %v", err)
		}

		err = w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(p.ExternalID),
			Value: data,
		})
		if err != nil {
			log.Fatalf("failed to send message %d: %v", i+1, err)
		}

		fmt.Printf("[%d/10] sent: %s\n", i+1, p.ExternalID)
	}

	fmt.Println("done")
}
