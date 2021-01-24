package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StravaActivity holds Strava-calculated attributes about an activity.
// I'm mainly interested in the filename of the route file (GPX/TCX/FIT).
type StravaActivity struct {
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

// ParseActivitiesFile reads a Strava-generated CSV file of all activities.
// Each record in the CSV holds the route filename as well as statistics about the activity.
func ParseActivitiesFile(recordFile *os.File) []StravaActivity {
	// Initialize the reader
	reader := csv.NewReader(recordFile)
	// Read all the records
	activityRecords, _ := reader.ReadAll()

	stravaActivities := make([]StravaActivity, len(activityRecords)-1)

	// Loop through lines, ignoring the headers
	for i, activityRecord := range activityRecords[1:] {
		time, _ := time.Parse("Jan 2, 2006, 3:04:05 PM", activityRecord[1])
		distance, _ := strconv.ParseFloat(strings.Replace(activityRecord[15], ",", "", -1), 32)
		elapsedTime, _ := strconv.ParseUint(strings.Replace(activityRecord[13], ",", "", -1), 10, 32)
		movingTime, _ := strconv.ParseUint(strings.Replace(activityRecord[14], ",", "", -1), 10, 32)
		elevationGain, _ := strconv.ParseFloat(strings.Replace(activityRecord[18], ",", "", -1), 32)
		elevationLoss, _ := strconv.ParseFloat(strings.Replace(activityRecord[19], ",", "", -1), 32)
		elevationMin, _ := strconv.ParseFloat(strings.Replace(activityRecord[20], ",", "", -1), 32)
		elevationMax, _ := strconv.ParseFloat(strings.Replace(activityRecord[21], ",", "", -1), 32)

		stravaActivities[i] = StravaActivity{
			Date:          time,
			Title:         activityRecord[2],
			Type:          activityRecord[3],
			Description:   activityRecord[4],
			Filename:      activityRecord[10],
			Distance:      float32(distance),
			ElapsedTime:   uint(elapsedTime),
			MovingTime:    uint(movingTime),
			ElevationGain: float32(elevationGain),
			ElevationLoss: float32(elevationLoss),
			ElevationMin:  float32(elevationMin),
			ElevationMax:  float32(elevationMax),
		}
	}

	return stravaActivities
}

// ParseActivity reads a StravaActivity struct and invokes the corresponding route parser.
// If the corresponding route is compressed, ParseActivity will decompress the file.
func ParseActivity(stravaActivity StravaActivity) {
	filename := stravaActivity.Filename
	if filename != "" { // Stationary activities (such as indoor row) do not have route files
		file, _ := ioutil.ReadFile(filepath.Join(ArchivePath, filename)) // Byte array of decompresed file
		fileParts := strings.Split(filename, ".")                        // Used to determine file type and compression status
		if len(fileParts) == 3 && fileParts[2] == "gz" {                 // File is compressed
			b := bytes.NewBuffer(file)
			var r io.Reader
			r, err := gzip.NewReader(b)
			if err != nil {
				return
			}
			var fileBuffer bytes.Buffer
			_, err = fileBuffer.ReadFrom(r)
			if err != nil {
				return
			}
			file = fileBuffer.Bytes()
			fmt.Println(string(file[:100]))
		}
	}
}
