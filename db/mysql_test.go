package db

import (
	"database/sql"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestMysql(t *testing.T) {
	t.Log("what?")
	db, err := sql.Open("mysql", "username:password@(192.168.163.133)/gls?charset=utf8")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}

	if err := db.Ping(); err != nil {
		t.Error(err)
	}

	defer db.Close()

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO animal VALUES(?, ?, ? )") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Insert square numbers for 0-24 in the database
	for i := 0; i < 1; i++ {
		result, err1 := stmtIns.Exec(0, "熊猫", "熊猫") // Insert tuples (i, i^2)
		if err1 != nil {
			panic(err1.Error()) // proper error handling instead of panic in your app
		}
		insertId, _ := result.LastInsertId()
		rowaffec, _ := result.RowsAffected()
		t.Error("insert result: ", insertId, rowaffec)
	}

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT id FROM `users`")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var id int

	// Query the square-number of 13
	err = stmtOut.QueryRow().Scan(&id) // WHERE number = 13
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	t.Log("The id is: ", id)

	// var squareNum int // we "scan" the result in here

	// // Query another number.. 1 maybe?
	// err = stmtOut.QueryRow(1).Scan(&squareNum) // WHERE number = 1
	// if err != nil {
	// 	panic(err.Error()) // proper error handling instead of panic in your app
	// }
	// fmt.Printf("The square number of 1 is: %d", squareNum)
}
