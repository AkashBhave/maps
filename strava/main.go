package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// ArchivePath is the relative path of the unzipped Strava bulk export
const ArchivePath = "./archive"

func main() {
	// Open the file
	recordFile, err := os.Open(filepath.Join(ArchivePath, "activities.csv"))
	if err != nil {
		log.Fatal("An error while opening the activities file: ", err)
	}

	activities := ParseActivitiesFile(recordFile)
	for _, activity := range activities {
		points, err := ParseActivity(activity)
		if err != nil {
			log.Fatal("An error while reading an activity: ", err)
		}
		activity.Track = points
		fmt.Println(activity)
	}
}
