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
  "model": "iPhone 15 Pro 256GB Black",
  "sku": "APPL-IP15P-256-BLK",
  "price": "999.99",
  "stock_status": "IN_STOCK",
  "quantity": 50,
  "is_enable": true
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
