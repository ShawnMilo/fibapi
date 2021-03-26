package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
)

const duplicateRecordCode = "23505"

var dbConn = "postgres://postgres:%s@localhost:%s/postgres?sslmode=disable"

var db *sql.DB

// Initialize and test database connection.
func init() {
	connStr := fmt.Sprintf(dbConn, os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"))
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("could not connect to %s: %s", connStr, err)
	}
	db = conn

	for i := 0; i < 3; i++ {
		if err = conn.Ping(); err != nil {
			time.Sleep(time.Second)
		} else {
			initializeDB()
			go fib()
			return
		}
	}
	log.Fatal("could not ping database")
}

// Create database table if it doesn't already exist.
func initializeDB() {
	row := db.QueryRow("SELECT id FROM fibonacci LIMIT 1")
	var i int
	err := row.Scan(&i)
	if err == nil {
		return
	}
	log.Print("attempting to initialize db")
	_, err = db.Exec("CREATE TABLE fibonacci (id INTEGER PRIMARY KEY, val INTEGER)")
	if err != nil {
		log.Printf("table creation failed: %s", err)
	} else {
		log.Println("created table OK")
	}

}

// Store Fibonacci data in database.
func memoize(v Value) {
	mu.Lock()
	defer mu.Unlock()
	_, err := db.Exec("INSERT INTO fibonacci (id, val) VALUES ($1, $2)", v.ID, v.Num)
	if err != nil && !isDuplicate(err) {
		log.Printf("failed to write to DB: %s", err)
		return
	}
	highCount = v.ID
}

// Ignore errors due to attempting to cache an already-cached value.
func isDuplicate(err error) bool {
	pe, ok := err.(*pq.Error)
	if !ok {
		log.Println("failed to convert")
		return false
	}
	return pe.Code == duplicateRecordCode
}

func ordinalFromDB(id int) (Value, error) {
	row := db.QueryRow("SELECT id, val FROM fibonacci WHERE id = $1", id)
	var i int
	var num int64
	err := row.Scan(&i, &num)
	return Value{ID: i, Num: num}, err
}

func belowFromDB(below int) (Value, error) {
	mu.RLock()
	c, found := belowCache[below]
	mu.RUnlock()
	if found {
		log.Printf("%d from below cache", below)
		return Value{Count: c}, nil
	}
	row := db.QueryRow("SELECT COUNT(id) FROM fibonacci WHERE val < $1", below)
	var i int
	err := row.Scan(&i)
	if err == nil {
		mu.Lock()
		belowCache[below] = i
		mu.Unlock()
	}
	return Value{Count: i}, err
}
