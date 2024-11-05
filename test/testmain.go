package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	f, e := os.Open("buckets.csv")
	if e != nil {
		fmt.Println(e)
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	var n int
	for {
		if l, e := reader.Read(); e != nil {
			break
		} else {
			n++
			fmt.Println(n, ": ", l, len(l))
		}
	}
}
