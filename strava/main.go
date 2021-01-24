package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	Date          time.Time
	Title         string
	Type          string
	Description   string
	Filename      string
	Distance      float32
	ElapsedTime   uint
	MovingTime    uint
	ElevationGain float32
	ElevationLoss float32
	ElevationMin  float32
	ElevationMax  float32
}

func main() {
	// Open the file
	recordFile, err := os.Open("./archive/activities.csv")
	if err != nil {
		fmt.Println("An error encountered ::", err)
	}
	// Initialize the reader
	reader := csv.NewReader(recordFile)
	// Read all the records
	records, _ := reader.ReadAll()

	// Loop through lines, ignoring the headers
	for _, record := range records[1:] {
		time, _ := time.Parse("Jan 2, 2006, 3:04:05 PM", record[1])
		distance, _ := strconv.ParseFloat(strings.Replace(record[15], ",", "", -1), 32)
		elapsedTime, _ := strconv.ParseUint(strings.Replace(record[13], ",", "", -1), 10, 32)
		movingTime, _ := strconv.ParseUint(strings.Replace(record[14], ",", "", -1), 10, 32)
		elevationGain, _ := strconv.ParseFloat(strings.Replace(record[18], ",", "", -1), 32)
		elevationLoss, _ := strconv.ParseFloat(strings.Replace(record[19], ",", "", -1), 32)
		elevationMin, _ := strconv.ParseFloat(strings.Replace(record[20], ",", "", -1), 32)
		elevationMax, _ := strconv.ParseFloat(strings.Replace(record[21], ",", "", -1), 32)

		data := Record{
			Date:          time,
			Title:         record[2],
			Type:          record[3],
			Description:   record[4],
			Filename:      record[10],
			Distance:      float32(distance),
			ElapsedTime:   uint(elapsedTime),
			MovingTime:    uint(movingTime),
			ElevationGain: float32(elevationGain),
			ElevationLoss: float32(elevationLoss),
			ElevationMin:  float32(elevationMin),
			ElevationMax:  float32(elevationMax),
		}
		fmt.Println(data.Filename)
	}
}
