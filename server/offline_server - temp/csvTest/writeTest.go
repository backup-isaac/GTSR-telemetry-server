package main

import (
    "fmt"
    "os"
    "log"
    "encoding/csv"
)

var data = [][]string{{"4:30am", "4:45am"}, {"5:00am", "7:00am"}}

func main() {

    // Create writer object
    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write data to csv file
    for _, value := range data {
        fmt.Println("Writing " + value[0] + " and " + value[1] + " to file")
        err := writer.Write(value)
        if err != nil{
            log.Fatal(err)
        }
    }
}
