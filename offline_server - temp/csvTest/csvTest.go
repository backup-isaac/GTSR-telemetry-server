package main

import (
    "fmt"
    "os"
    "log"
    "encoding/csv"
)

var data = [][]string{{"4:30am", "4:45am"}, {"5:00am", "7:00am"}}

// Structure for start/end time pairs
type pair struct {
    start time.Time
    end time.Time
}

// Contains times that have been merged
type merged struct {
    times []pair
}

// Instance of previously merged times
var mergedTimes := loadTimes()

// Turns date/timezone strings onto RFC3339Nano time.Time
func formatRFC3339(date string, timezone string) (time.Time,error) {

    // Make sure there are seconds
    if len(date) == 16 {
        date += ":00"
    }

    // Add nanoseconds
    date += ".999999999"

    // Add timezone
    if timezone[0] == '+' || timezone[0] == '-' {
        date = date + timezone
    } else {
        date = date + "Z"
    }

    // Parse string into time.Time() datatype
    t,err := time.Parse(time.RFC3339Nano,date)

    // Return the time and any errors
    return t,err
}

// Loads times and stores it into merged struct
func loadTimes() (merged) {

    // Open a file if it exists, create it if it doesn't
    file, err := os.OpenFile("times.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Create reader object
    reader := csv.NewReader(file)
    m := new(merged)

    // Read data from csv file
    for {
        // Keep reading each line
        record, err := reader.Read()

        // If there are no more lines, break out of the loop
        if err == io.EOF {
          break
        }

        // If there is an error, log and display it
        if err != nil {
          fmt.Println("Error in reading csv")
          log.Fatal(err)
        }

        // Turn the strings into time.Time() datatype
        time1,err = formatRFC3339(record[0])
        if err != nil{
          fmt.Println("Error in converting 1st string in pair to time.Time()")
          log.Fatal(err)
        }
        time2,err = formatRFC3339(record[1])
        if err != nil{
          fmt.Println("Error in converting 2nd string in pair to time.Time()")
          log.Fatal(err)
        }

        // Store the times into the pair struct
        p := new(pair)
        p.start = time1
        p.end = time2

        // Append the times into the merged struct
        m.times = append(m.times, p)
    }

    // Debug check if times are stored in merged struct
    for _, tuple := range m.times {
        fmt.Println(tuple.start)
        fmt.Println(tuple.end)
        fmt.Println()
    }

    // Return merged times to mergedTimes var
    return m
}

// Test to add times into merged struct
func addTimes(start time.Time, end time.Time) {

}

func saveTimes() {

    // Open a file if it exists, create it if it doesn't
    file, err := os.OpenFile("times.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Create writer object
    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Convert times to strings

    // Write data to csv file
    for _, value := range mergedTimes {
        fmt.Println("Writing " + value[0] + " and " + value[1] + " to file")
        err := writer.Write(value)
        if err != nil{
            log.Fatal(err)
        }
    }
}

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
