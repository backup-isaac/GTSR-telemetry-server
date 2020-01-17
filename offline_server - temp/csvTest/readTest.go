package main

import (
    "fmt"
    "os"
    "log"
    "encoding/csv"
    "io"
)

var data = [][]string{{"4:30am", "4:45am"}, {"5:00am", "7:00am"}}

// Structure for start/end time pairs
type pair struct {
    start string
    end string
}

// Contains times that have been merged
type merged struct {
    times []*pair
}

func main() {

    // Open a file if it exists, create it if it doesn't
    file, err := os.OpenFile("times.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Create reader object
    reader := csv.NewReader(file)
    m := new(merged)

    // Read data from csv file
    for {
        record, err := reader.Read()
        if err == io.EOF {
          break
        }
        if err != nil {
          fmt.Println("Error in reading csv")
          log.Fatal(err)
        }

        p := new(pair)
        p.start = record[0]
        p.end = record[1]

        m.times = append(m.times, p)
    }

    for _, tuple := range m.times {
        fmt.Println(tuple.start)
        fmt.Println(tuple.end)
        fmt.Println()
    }

}
