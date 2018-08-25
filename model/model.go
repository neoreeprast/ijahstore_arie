package model

import "time"

// Product : represent product in the system
type Product struct {
	Sku      string
	ItemName string
	Stock    int64
}

// PurchaseOrder : represent purchase of a product
type PurchaseOrder struct {
	Product       Product
	Price         float64
	ReceiptNumber string
	OrderedQty    int64
	PurchaseTime  time.Time
}

// ItemReceived : one PurchaseOrder can have more than 1 receipt
type ItemReceived struct {
	PurchaseOrder PurchaseOrder
	ReceivedQty   int64
	ReceivedTime  time.Time
}

// OrderItem : one sold product
type OrderItem struct {
	Product Product
	Qty     int64
	Price   float64
}

// SalesOrder : created when product got sold
type SalesOrder struct {
	OrderDate time.Time
	OrderID   string
	Items     []OrderItem
}

type ValuationReportItem struct {
	SKU             string
	ItemName        string
	CurrentStock    int64
	AverageBuyPrice float64
	Total           float64
}

type SalesReportItem struct {
	OrderID   string
	OrderTime time.Time
	SKU       string
	ItemName  string
	Qty int64
	SalePrice float64
	TotalSale float64
	BuyPrice float64
	Revenue float64
}
