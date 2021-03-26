package main

var fibChan = make(chan Value)

type Value struct {
	ID    int   `json:"id,omitempty"`
	Num   int64 `json:"num,omitempty"`
	Count int   `json:"count"`
}

// Don't start Fibonacci at 1, 1 if a previous instance has cached it.
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

// Generate the Fibonacci sequence on demand.
func fib() {
	count, prev, current, err := seedFib()
	if err != nil || count == 0 {
		count = 1
		current = int64(1)

		v := Value{ID: count, Num: current}
		memoize(v)
		fibChan <- v
	}

	for {
		count++
		last := current
		current += prev
		v := Value{ID: count, Num: current}
		memoize(v)
		if v.ID == 4 {
			memoize(v)
		}
		fibChan <- v
		prev = last
	}
}

// Check how far we are into the cycle.
func countReached(i int) bool {
	mu.RLock()
	defer mu.RUnlock()
	return highCount >= i
}
