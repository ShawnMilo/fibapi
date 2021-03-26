package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var fibChan = make(chan Value)

var mu sync.RWMutex
var highCount int
var belowCache = make(map[int]int)

var dbConn = "postgres://postgres:%s@localhost:%s/postgres?sslmode=disable"

type Value struct {
	ID    int   `json:"id,omitempty"`
	Num   int64 `json:"num,omitempty"`
	Count int   `json:"count"`
}

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

func seedFib() (int, int64, int64, error) {
	var prev, current int64
	var count int

	rows, err := db.Query("SELECT id, val FROM fibonacci ORDER BY id DESC LIMIT 2")
	if err != nil {
		return count, prev, current, err
	}
	defer rows.Close()

	if !rows.Next() {
		return count, prev, current, err
	}

	err = rows.Scan(&count, &prev)
	if err != nil {
		return count, prev, current, err
	}

	if !rows.Next() {
		return count, prev, current, err
	}

	var temp int
	err = rows.Scan(&temp, &current)
	return count, prev, current, err
}

func fib() {
	count, prev, current, err := seedFib()
	if err != nil || count == 0 {
		count = 1
		current = int64(1)

		v := Value{ID: count, Num: current}
		memoize(v)
		fibChan <- v
	}
	log.Printf("starting: %d, %d, %d", count, prev, current)

	for {
		count++
		last := current
		current += prev
		v := Value{ID: count, Num: current}
		go memoize(v)
		fibChan <- v
		prev = last
	}
}

func memoize(v Value) {
	mu.Lock()
	defer mu.Unlock()
	_, err := db.Exec("INSERT INTO fibonacci (id, val) VALUES ($1, $2)", v.ID, v.Num)
	if err != nil {
		log.Printf("failed to write to DB: %s", err)
		return
	}
	highCount = v.ID
}

func countReached(i int) bool {
	mu.RLock()
	defer mu.RUnlock()
	return highCount >= i
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

func reachCount(i int) {
	for v := range fibChan {
		if v.ID >= i {
			return
		}
	}
}

func byOrdinal(i int) (Value, error) {
	if countReached(i) {
		return ordinalFromDB(i)
	}
	reachCount(i)
	time.Sleep(time.Second)
	return byOrdinal(i)
}

func main() {
	http.HandleFunc("/ordinal/", ordinal)
	http.HandleFunc("/below/", below)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func ordinal(w http.ResponseWriter, r *http.Request) {
	i := uriToInt(r.RequestURI)
	fib, err := byOrdinal(i)
	if err != nil {
		log.Printf("failed byOrdinal: %s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(fib)
	if err != nil {
		log.Printf("failed to marshal: %s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func below(w http.ResponseWriter, r *http.Request) {
	i := uriToInt(r.RequestURI)
	fib, err := belowFromDB(i)
	if err != nil {
		log.Printf("failed belowFromDB: %s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(fib)
	if err != nil {
		log.Printf("failed to marshal: %s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func uriToInt(uri string) int {
	parts := strings.Split(uri, "/")
	if len(parts) < 1 {
		return 0
	}
	num := parts[len(parts)-1]
	i, err := strconv.Atoi(num)
	if err != nil {
		log.Printf("Invalid ordinal %q", num)
	}
	return i
}
