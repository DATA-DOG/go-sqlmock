package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/kisielk/sqlstruct"
)

const ORDER_PENDING = 0
const ORDER_CANCELLED = 1

type User struct {
	Id       int     `sql:"id"`
	Username string  `sql:"username"`
	Balance  float64 `sql:"balance"`
}

type Order struct {
	Id          int     `sql:"id"`
	Value       float64 `sql:"value"`
	ReservedFee float64 `sql:"reserved_fee"`
	Status      int     `sql:"status"`
}

func cancelOrder(id int, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}

	var order Order
	var user User
	sql := fmt.Sprintf(`
SELECT %s, %s
FROM orders AS o
INNER JOIN users AS u ON o.buyer_id = u.id
WHERE o.id = ?
FOR UPDATE`,
		sqlstruct.ColumnsAliased(order, "o"),
		sqlstruct.ColumnsAliased(user, "u"))

	// fetch order to cancel
	rows, err := tx.Query(sql, id)
	if err != nil {
		tx.Rollback()
		return
	}

	defer rows.Close()
	// no rows, nothing to do
	if !rows.Next() {
		tx.Rollback()
		return
	}

	// read order
	err = sqlstruct.ScanAliased(&order, rows, "o")
	if err != nil {
		tx.Rollback()
		return
	}

	// ensure order status
	if order.Status != ORDER_PENDING {
		tx.Rollback()
		return
	}

	// read user
	err = sqlstruct.ScanAliased(&user, rows, "u")
	if err != nil {
		tx.Rollback()
		return
	}
	rows.Close() // manually close before other prepared statements

	// refund order value
	sql = "UPDATE users SET balance = balance + ? WHERE id = ?"
	refundStmt, err := tx.Prepare(sql)
	if err != nil {
		tx.Rollback()
		return
	}
	defer refundStmt.Close()
	_, err = refundStmt.Exec(order.Value+order.ReservedFee, user.Id)
	if err != nil {
		tx.Rollback()
		return
	}

	// update order status
	order.Status = ORDER_CANCELLED
	sql = "UPDATE orders SET status = ?, updated = NOW() WHERE id = ?"
	orderUpdStmt, err := tx.Prepare(sql)
	if err != nil {
		tx.Rollback()
		return
	}
	defer orderUpdStmt.Close()
	_, err = orderUpdStmt.Exec(order.Status, order.Id)
	if err != nil {
		tx.Rollback()
		return
	}
	return tx.Commit()
}

func main() {
	// @NOTE: the real connection is not required for tests
	db, err := sql.Open("mysql", "root:@/orders")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = cancelOrder(1, db)
	if err != nil {
		log.Fatal(err)
	}
}
