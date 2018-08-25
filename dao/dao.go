package dao

import (
	"database/sql"
	"fmt"
	"time"

	"../model"
)

// IjahDB : represent dao
type IjahDB struct {
	db *sql.DB
}

// InitDB : init database connection
func InitDB(filepath string) IjahDB {
	d, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if d == nil {
		panic("db nil")
	}
	return IjahDB{d}
}

// CreateDB : create database
func (i IjahDB) CreateDB() {
	sql := `
	CREATE TABLE IF NOT EXISTS product (
		SKU TEXT NOT NULL PRIMARY KEY,
		item_name text not null,
		current_stock INTEGER not null
	);
	`

	_, err := i.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	sql = `
	CREATE TABLE IF NOT EXISTS purchase_order (
		receipt_no text not null primary key,
		SKU TEXT not null,
		price REAL not null,
		ordered_qty integer not null,
		purchase_time DATETIME
	);
	`
	_, err = i.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	sql = `
	CREATE TABLE IF NOT EXISTS item_receipt (
		receipt_no text not null,
		received_qty integer not null,
		received_time DATETIME
	);
	`
	_, err = i.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	sql = `
	CREATE TABLE IF NOT EXISTS sales_order (
		order_id text not null primary key,
		order_time DATETIME
	);
	`
	_, err = i.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	sql = `
	CREATE TABLE IF NOT EXISTS order_item (
		order_id text not null,
		SKU text not null,
		order_qty integer,
		price REAL
	);
	`
	_, err = i.db.Exec(sql)
	if err != nil {
		panic(err)
	}
}

// Close : close db connection
func (i IjahDB) Close() {
	i.db.Close()
}

// AddProduct : add product
func (i IjahDB) AddProduct(p model.Product) {
	sqlAddItem := `
	INSERT INTO product (SKU, item_name, current_stock)
	VALUES(?, ?, ?);
	`

	stmt, err := i.db.Prepare(sqlAddItem)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err2 := stmt.Exec(p.Sku, p.ItemName, p.Stock)
	if err2 != nil {
		panic(err2)
	}
}

// GetProduct : get product by sku
func (i IjahDB) GetProduct(s string) (model.Product, bool) {
	sqlGetProduct := `
	select sku, item_name, current_stock from product where sku = ?;
	`
	stmt, err := i.db.Prepare(sqlGetProduct)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	r, err2 := stmt.Query(s)
	if err2 != nil {
		panic(err2)
	}
	defer r.Close()
	var sku string
	var itemName string
	var stock int64
	if r.Next() {
		r.Scan(&sku, &itemName, &stock)
		return model.Product{
			Sku:      sku,
			ItemName: itemName,
			Stock:    stock,
		}, true
	} else {
		return model.Product{}, false
	}
}

// CreatePO : create new Purchase Order
func (i IjahDB) CreatePO(po model.PurchaseOrder) {
	ep, s := i.GetProduct(po.Product.Sku)
	if s {
		sqlInsertPO := `
		INSERT INTO purchase_order
		values (?,?,?,?, CURRENT_TIMESTAMP);
		`
		stmt, err := i.db.Prepare(sqlInsertPO)
		defer stmt.Close()
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		_, err2 := stmt.Exec(po.ReceiptNumber, ep.Sku, po.Price, po.OrderedQty)
		if err2 != nil {
			panic(err2)
		}
	}
}

// GetPO : get purchase order by receipt number
func (i IjahDB) GetPO(rn string) (model.PurchaseOrder, bool) {
	sqlGetPo := `
	select receipt_no, SKU, price, ordered_qty, purchase_time  from purchase_order 
	where receipt_no = ?;
	`
	stmt, err := i.db.Prepare(sqlGetPo)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	r, err2 := stmt.Query(rn)
	if err2 != nil {
		panic(err2)
	}
	defer r.Close()
	var receiptNo string
	var sku string
	var price float64
	var qty int64
	var purchaseTime time.Time
	if r.Next() {
		r.Scan(&receiptNo, &sku, &price, &qty, &purchaseTime)
		return model.PurchaseOrder{
			Product: model.Product{
				Sku: sku,
			},
			Price:         price,
			ReceiptNumber: receiptNo,
			OrderedQty:    qty,
			PurchaseTime:  purchaseTime,
		}, true
	} else {
		return model.PurchaseOrder{}, false
	}
}

// InsertReceiptItem : receive an item
func (i IjahDB) InsertReceiptItem(ir model.ItemReceived) {
	po, suc := i.GetPO(ir.PurchaseOrder.ReceiptNumber)
	if suc {
		sqlInsertReceipt := `
			INSERT INTO item_receipt (receipt_no, received_qty, received_time)
			VALUES(?, ?, CURRENT_TIMESTAMP);
		`
		stmt, err := i.db.Prepare(sqlInsertReceipt)
		defer stmt.Close()
		if err != nil {
			panic(err)
		}
		tx, e1 := i.db.Begin()
		if e1 != nil {
			panic(e1)
		}
		_, err2 := tx.Stmt(stmt).Exec(po.ReceiptNumber, ir.ReceivedQty)
		if err2 != nil {
			fmt.Println(err2)
			tx.Rollback()
		}
		sqlUpdateStock := `
		UPDATE product SET current_stock = current_stock + ? where SKU = ?;
		`
		stmt2, err3 := i.db.Prepare(sqlUpdateStock)
		defer stmt2.Close()
		if err3 != nil {
			fmt.Println(err3)
			tx.Rollback()
		}
		_, err4 := tx.Stmt(stmt2).Exec(ir.ReceivedQty, po.Product.Sku)
		if err4 != nil {
			fmt.Println(err4)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}
}

func (i IjahDB) CreateSalesOrder(so model.SalesOrder) {
	sqlInsertSO := `
	INSERT INTO sales_order(order_id, order_time)
	VALUES(?, CURRENT_TIMESTAMP);
	`
	stmtInsert, err := i.db.Prepare(sqlInsertSO)
	if err != nil {
		panic(err)
	}
	tx, err1 := i.db.Begin()
	if err1 != nil {
		panic(err1)
	}
	_, err2 := tx.Stmt(stmtInsert).Exec(so.OrderID)
	if err2 != nil {
		fmt.Println(err2)
		tx.Rollback()
	}
	sqlInsertOI := `
	INSERT INTO order_item(order_id, SKU, order_qty, price)
	VALUES(?, ?, ?, ?);
	`
	sqlUpdateProduct := `
	UPDATE product SET current_stock=current_stock - ? where SKU = ?;
	`
	stmtInsertOI, _ := i.db.Prepare(sqlInsertOI)
	stmtUpdateP, _ := i.db.Prepare(sqlUpdateProduct)
	allSuccess := true
	for _, oi := range so.Items {
		_, err3 := tx.Stmt(stmtInsertOI).Exec(so.OrderID, oi.Product.Sku, oi.Qty, oi.Price)
		if err3 != nil {
			fmt.Println(err3)
			tx.Rollback()
			break
		}
		_, err4 := tx.Stmt(stmtUpdateP).Exec(oi.Qty, oi.Product.Sku)
		if err4 != nil {
			fmt.Println(err4)
			tx.Rollback()
			break
		}
	}
	if allSuccess {
		tx.Commit()
	}
}

// GetValuationReport : get current valuation report
func (i IjahDB) GetValuationReport() []model.ValuationReportItem {
	sqlSelect := `
	select sku, item_name, current_stock, average, current_stock*average from (
	select p.sku, p.item_name, current_stock, sum(po.price*received_qty)/sum(received_qty) average from product p
                inner join purchase_order po on p.SKU=po.SKU
                inner join item_receipt ir on po.receipt_no=ir.receipt_no
	group by p.sku, p.item_name, current_stock);
	`
	r, err := i.db.Query(sqlSelect)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	var sku string
	var itemName string
	var stock int64
	var average float64
	var total float64
	var result []model.ValuationReportItem
	for r.Next() {
		err1 := r.Scan(&sku, &itemName, &stock, &average, &total)
		if err1 != nil {
			panic(err1)
		}
		vri := model.ValuationReportItem{
			SKU:sku,
			ItemName:itemName,
			CurrentStock:stock,
			AverageBuyPrice:average,
			Total:total,
		}
		result = append(result, vri)
	}
	return result
}

// GetSalesReport : Get Current Sales Report
func (i IjahDB) GetSalesReport() []model.SalesReportItem {
	sqlSelect := `
	select a.order_id, a.order_time, a.sku, a.item_name, a.order_qty, a.price, a.total_sales,
       a.order_qty*b.average buy_price, a.total_sales - (a.order_qty*b.average) revenue
	from (
		select so.order_id, so.order_time, p.sku, p.item_name, oi.order_qty, oi.price, oi.order_qty*oi.price total_sales
		from sales_order so
		inner join order_item oi on so.order_id = oi.order_id
		inner join product p on oi.sku = p.sku) a
	inner join (
		select p.sku, sum(po.price*received_qty)/sum(received_qty) average
		from product p
		inner join purchase_order po on p.SKU=po.SKU
		inner join item_receipt ir on po.receipt_no=ir.receipt_no
		group by p.sku) b on a.sku = b.sku;
	`
	r, err := i.db.Query(sqlSelect)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	var orderId string
	var orderTime time.Time
	var sku string
	var itemName string
	var qty int64
	var price float64
	var total float64
	var averageBuy float64
	var revenue float64
	var result []model.SalesReportItem
	for r.Next() {
		err1 := r.Scan(&orderId, &orderTime, &sku, &itemName, &qty, &price, &total, &averageBuy, &revenue)
		if err1 != nil {
			panic(err1)
		}
		sri := model.SalesReportItem{
			orderId,
			orderTime,
			sku,
			itemName,
			qty,
			price,
			total,
			averageBuy,
			revenue,
		}
		result = append(result, sri)
	}
	return result
}
