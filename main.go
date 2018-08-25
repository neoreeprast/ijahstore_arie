package main

import (
	"encoding/json"
	"log"
	"net/http"

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
	json.NewEncoder(w).Encode(product)
}

// CreatePO : create purchse order handler
func CreatePO(w http.ResponseWriter, r *http.Request) {
	var po model.PurchaseOrder
	_ = json.NewDecoder(r.Body).Decode(&po)
	db.CreatePO(po)
	json.NewEncoder(w).Encode(po)
}

func ReceiveProduct(w http.ResponseWriter, r *http.Request) {
	var ir model.ItemReceived
	_ = json.NewDecoder(r.Body).Decode(&ir)
	db.InsertReceiptItem(ir)
	json.NewEncoder(w).Encode(ir)
}

func CreateSO(w http.ResponseWriter, r *http.Request) {
	var so model.SalesOrder
	_ = json.NewDecoder(r.Body).Decode(&so)
	db.CreateSalesOrder(so)
	json.NewEncoder(w).Encode(so)
}

func ViewValuationReport(w http.ResponseWriter, r *http.Request) {
	valuationReport := db.GetValuationReport()
	json.NewEncoder(w).Encode(valuationReport)
}

func ViewSalesReport(w http.ResponseWriter, r *http.Request) {
	salesReport := db.GetSalesReport()
	json.NewEncoder(w).Encode(salesReport)
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
	r.HandleFunc("/sales", ViewValuationReport).Methods("GET")
	defer db.Close()
	log.Fatal(http.ListenAndServe(":8000", r))
}
