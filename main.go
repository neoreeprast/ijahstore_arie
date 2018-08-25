package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"./dao"
	"./model"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db dao.IjahDB

// CreateProduct : create product handler
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product model.Product
	_ = json.NewDecoder(r.Body).Decode(&product)
	db.AddProduct(product)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// CreatePO : create purchse order handler
func CreatePO(w http.ResponseWriter, r *http.Request) {
	var po model.PurchaseOrder
	_ = json.NewDecoder(r.Body).Decode(&po)
	db.CreatePO(po)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(po)
}

func ReceiveProduct(w http.ResponseWriter, r *http.Request) {
	var ir model.ItemReceived
	_ = json.NewDecoder(r.Body).Decode(&ir)
	db.InsertReceiptItem(ir)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ir)
}

func CreateSO(w http.ResponseWriter, r *http.Request) {
	var so model.SalesOrder
	_ = json.NewDecoder(r.Body).Decode(&so)
	db.CreateSalesOrder(so)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(so)
}

func ViewValuationReport(w http.ResponseWriter, r *http.Request) {
	valuationReport := db.GetValuationReport()
	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	for _, vr := range valuationReport {
		wr.Write([]string{
			vr.SKU,
			vr.ItemName,
			strconv.FormatInt(vr.CurrentStock,10),
			fmt.Sprintf("%.0f",vr.AverageBuyPrice),
			fmt.Sprintf("%.0f",vr.Total),
		})
	}
	wr.Flush()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=ValuationReport.csv")
	w.Write(b.Bytes())
}

func ViewSalesReport(w http.ResponseWriter, r *http.Request) {
	salesReport := db.GetSalesReport()
	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	for _, sr := range salesReport {
		wr.Write([]string{
			sr.OrderID,
			sr.OrderTime.Format("2006-01-02 15:04:05"),
			sr.SKU,
			sr.ItemName,
			strconv.FormatInt(sr.Qty,10),
			fmt.Sprintf("%.0f",sr.SalePrice),
			fmt.Sprintf("%.0f",sr.TotalSale),
			fmt.Sprintf("%.0f",sr.BuyPrice),
			fmt.Sprintf("%.0f",sr.Revenue),
		})
	}
	wr.Flush()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=SalesReport.csv")
	w.Write(b.Bytes())
}

func init() {
	db = dao.InitDB("./ijah.db")
	db.CreateDB()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/products", CreateProduct).Methods("POST")
	r.HandleFunc("/pos", CreatePO).Methods("POST")
	r.HandleFunc("/pos/receipts", ReceiveProduct).Methods("POST")
	r.HandleFunc("/sos", CreateSO).Methods("POST")
	r.HandleFunc("/valuation", ViewValuationReport).Methods("GET")
	r.HandleFunc("/sales", ViewSalesReport).Methods("GET")
	defer db.Close()
	log.Fatal(http.ListenAndServe(":8000", r))
}
