package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type DBHandler struct {
	Connection *sql.DB
}

func (db *DBHandler) Init() (err error) {

	path, _ := os.Getwd()
	fmt.Println(path)

	var sqlite *sql.DB

	var db_file = path + "/bot.db"

	if _, err = os.Stat(db_file); os.IsNotExist(err) {
		fmt.Println("Create new database file")
		var file *os.File
		file, err = os.Create(db_file)
		if err != nil {
			return
		}
		file.Close()
		fmt.Println("bot database created")
		sqlite, _ = sql.Open("sqlite3", db_file)
		fmt.Println("Create tables...")
		var statement *sql.Stmt
		statement, err = sqlite.Prepare(create_db) // Prepare SQL Statement
		if err != nil {
			return
		}
		_, err = statement.Exec()
		if err != nil {
			return
		}

		statement, err = sqlite.Prepare(action_insert)
		_, err = statement.Exec()
		if err != nil {
			return
		}

	} else {
		sqlite, _ = sql.Open("sqlite3", db_file)
	}
	sqlite.SetMaxOpenConns(1)
	db.Connection = sqlite
	return
}
