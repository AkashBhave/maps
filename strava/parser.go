package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GPSBabel is the location of the GPSBabel executable.
const GPSBabel = "/Applications/GPSBabelFE.app/Contents/MacOS/gpsbabel"

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

// Point holds useful information about a point along an activity track.
type Point struct {
	Lat       float64
	Lon       float64
	Ele       float64
	Time      time.Time
	HeartRate float64
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
		}

		switch fileParts[1] { // Parse the actual file
		case "gpx":
			ParseGPXFile(file)
		case "tcx":
			ParseTCXFile(file)
		case "fit":
			ParseFITFile(file)
		default:
			break
		}
	}
}

// ParseGPXFile extracts information from a GPX file.
func ParseGPXFile(file []byte) {
	type GPXPoint struct {
		Lat         float64 `xml:"lat,attr"`
		Lon         float64 `xml:"lon,attr"`
		Ele         float64 `xml:"ele,omitempty"`
		Time        string  `xml:"time,omitempty"`
		MagVar      string  `xml:"magvar,omitempty"`
		GeoidHeight string  `xml:"geoidheight,omitempty"`
	}
	type Result struct {
		XMLName xml.Name `xml:"gpx"`
		// Track   *Track   `xml:"trk"`
		Points []*GPXPoint `xml:"trk>trkseg>trkpt"`
	}
	result := &Result{}
	err := xml.Unmarshal(file, &result)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
}

// ParseTCXFile extracts information from a TCX file.
func ParseTCXFile(file []byte) {
	type TCXPoint struct {
		Lat       float64 `xml:"Position>LatitudeDegrees"`
		Lon       float64 `xml:"Position>LongitudeDegrees"`
		Ele       float64 `xml:"AltitudeMeters,omitempty"`
		Time      time.Time
		HeartRate float64 `xml:"HeartRateBpm>Value,omitempty"`
	}
	type Result struct {
		XMLName xml.Name    `xml:"TrainingCenterDatabase"`
		Points  []*TCXPoint `xml:"Activities>Activity>Lap>Track>Trackpoint"`
	}
	result := &Result{}
	err := xml.Unmarshal(file, &result)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
}

// ParseFITFile goes through a bunch of hurdles to extract information from a FIT file.
// First, I need to write the byte array to a temporary file (since the orig. file may be compressed).
// Next, I use GPSBabel to convert the temporary FIT file into a GPX file.
// Finally, I invoke the good ol' ParseGPXFile function.
// C'mon, Garmin, why is this so hard?
func ParseFITFile(file []byte) {
	tempFile, err := ioutil.TempFile("", "*.fit")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tempFile.Name()) // Automatically delete the file

	if _, err := tempFile.Write(file); err != nil {
		log.Fatal(err)
	}
	if err := tempFile.Close(); err != nil {
		log.Fatal(err)
	}

	newFile, err := exec.Command(
		GPSBabel,
		"-t",
		"-i", "garmin_fit",
		"-f", tempFile.Name(), // Read the temporary filename
		"-o", "gpx",
		"-F", "-", // Write to stdout
	).Output()

	ParseGPXFile(newFile)
}
