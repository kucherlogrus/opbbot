package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

type DBHandler struct {
	Connection *sql.DB
}

func (db *DBHandler) Init() (err error) {

	path, _ := os.Getwd()

	var sqlite *sql.DB

	var db_file = path + "/bot.db"

	if _, err = os.Stat(db_file); os.IsNotExist(err) {
		return fmt.Errorf("Database file is not exist in path " + path)

	} else {
		sqlite, _ = sql.Open("sqlite3", db_file)
	}
	sqlite.SetMaxOpenConns(1)
	db.Connection = sqlite
	return
}

func (db *DBHandler) GetAccessToken(service_name string) (token string, err error) {
	statement, err := db.Connection.Prepare("SELECT s.access_token FROM service s WHERE s.name = ?")
	if err != nil {
		return
	}
	err = statement.QueryRow(service_name).Scan(&token)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (db *DBHandler) CheckAccessTokenExpired(service_name string) (status bool, err error) {
	statement, err := db.Connection.Prepare("SELECT s.expired_at,  FROM service s WHERE s.name = ?")
	now := time.Now()
	var expired_at time.Time
	err = statement.QueryRow(service_name).Scan(&expired_at)
	if err != nil {
		log.Fatal(err)
	}
	if now.After(expired_at) {
		return true, nil
	}

	if expired_at.Sub(now).Seconds() < 300 {
		return true, nil
	}
	return false, nil
}

func (db *DBHandler) RefreshAccessToken(service_name string, access_token string, expired_at time.Time) (err error) {
	statement, err := db.Connection.Prepare("INSERT INTO service(access_token, expired_at) VALUES (?, ?, datetime('now'))")
	_, err = statement.Exec(service_name, access_token, expired_at)
	return
}
