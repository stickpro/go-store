# Kafka Integration Guide

## Overview

The system consumes messages from the Kafka topic `store.products` and synchronizes products.

Logic: if a product with the given `external_id` already exists — it is updated, otherwise — it is created.

---

## Connection

| Parameter | Value |
|---|---|
| Broker (external) | `localhost:19092` |
| Topic | `store.products` |
| Partitioning | by `external_id` (message key) |

---

## Message Format

**Key:** product `external_id` — a string, e.g. `ERP-12345`

**Value:** JSON

```json
{
  "external_id": "ERP-12345",
  "sku": "APPL-IP15P-256-BLK",
  "price_retail": "999.99",
  "price_business": "949.99",
  "price_wholesale": "899.99",
  "stock_status": "IN_STOCK",
  "quantity": 50,
  "is_enable": true,
  "weight": "0.5",
  "length": "10.0",
  "width": "5.0",
  "height": "0.5",
  "image_main": "https://example.com/images/iphone15pro.jpg",
  "images" : [
    "https://example.com/images/iphone15pro-front.jpg",
    "https://example.com/images/iphone15pro-back.jpg"
  ],
  "attributes": [
    {
      "name": "Цвет",
      "slug": "color",
      "type": "select",
      "value": "Чёрный"
    },
    {
      "name": "Объём памяти",
      "slug": "storage",
      "type": "select",
      "unit": "GB",
      "value": "256"
    },
    {
      "name": "Диагональ экрана",
      "slug": "screen_size",
      "type": "number",
      "unit": "inch",
      "value": "6.1"
    }
  ]
}
```

---

## Field Reference

| Field | Type | Required | Description |
|---|---|---|---|
| `external_id` | string | **yes** | Unique product ID in your system. Used to find the product on update |
| `model` | string | **yes** | Product name / model |
| `sku` | string | no | Stock Keeping Unit |
| `price` | string/number | **yes** | Price. Use a dot as decimal separator: `"999.99"` |
| `stock_status` | string | **yes** | Availability status (see below) |
| `quantity` | number | **yes** | Stock quantity. Use `0` when out of stock |
| `is_enable` | bool | **yes** | Whether the product is visible on the storefront. `true` / `false` |

### `stock_status` values

| Value | Description |
|---|---|
| `IN_STOCK` | Available |
| `OUT_OF_STOCK` | Not available |
| `PRE_ORDER` | Available for pre-order |

---

## Examples

**Create / update an available product:**
```json
{
  "external_id": "ERP-12345",
  "model": "Samsung Galaxy S24 Ultra",
  "sku": "SAM-S24U-256-BLK",
  "price": "1199.00",
  "stock_status": "IN_STOCK",
  "quantity": 15,
  "is_enable": true
}
```

**Update stock (product sold out):**
```json
{
  "external_id": "ERP-12345",
  "model": "Samsung Galaxy S24 Ultra",
  "price": "1199.00",
  "stock_status": "OUT_OF_STOCK",
  "quantity": 0,
  "is_enable": true
}
```

**Hide a product:**
```json
{
  "external_id": "ERP-12345",
  "model": "Samsung Galaxy S24 Ultra",
  "price": "1199.00",
  "stock_status": "OUT_OF_STOCK",
  "quantity": 0,
  "is_enable": false
}
```

---

## Notes

- The `external_id` field is **immutable** — do not change it after the first message, otherwise a new product will be created
- Use a dot as the decimal separator for price: `"999.99"`, not `"999,99"`
- Always send **all fields** on every message — partial updates are not supported
