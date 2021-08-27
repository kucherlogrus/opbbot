package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

type ServiceDB struct {
	ID           int
	Name         string
	Client       string
	Secret       string
	Access_token string
	Expire_at    *time.Time
	Create_at    *time.Time
	Update_at    *time.Time
}

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
	sqlite.SetMaxOpenConns(4)
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
	statement, err := db.Connection.Prepare("UPDATE service set access_token=?, expired_at=?, update_at=datetime('now') WHERE name =?")
	if err != nil {
		fmt.Println(err)
	}
	_, err = statement.Exec(access_token, expired_at, service_name)
	return
}

func (db *DBHandler) GetService(service_name string) (*ServiceDB, error) {
	var service ServiceDB
	statement, err := db.Connection.Prepare("SELECT s.*  FROM service s WHERE s.name = ?")
	err = statement.QueryRow(service_name).Scan(&service.ID, &service.Name, &service.Client,
		&service.Secret, &service.Access_token, &service.Expire_at,
		&service.Create_at, &service.Update_at,
	)
	if err != nil {
		return nil, err
	}
	return &service, nil
}
