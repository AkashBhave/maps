package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ArchivePath is the relative path of the unzipped Strava bulk export
const ArchivePath = "./archive"

func main() {
	// Open the file
	recordFile, err := os.Open(filepath.Join(ArchivePath, "activities.csv"))
	if err != nil {
		fmt.Println("An error encountered: ", err)
	}

	stravaActivities := ParseActivitiesFile(recordFile)
	for _, stravaActivity := range stravaActivities[:100] {
		ParseActivity(stravaActivity)
	}
}
