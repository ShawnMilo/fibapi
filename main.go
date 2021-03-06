package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var mu sync.RWMutex
var highCount int

// Force the generator to reach the desired level.
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
