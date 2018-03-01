// Author: Josh Kendrick
// Do whatever you want with this code

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"

	"github.com/mostlygeek/go-exiftool"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify a directory")
	}
	// parent dir
	directory := os.Args[1]

	// open db
	boltDB, err := bolt.Open(
		"export-file-tags.db",
		0600,
		&bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer boltDB.Close()

	// create the bucket if it doesnt exist
	boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tags"))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	// get all files
	fileList := []string{}
	filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	// loop through the files
	for _, filepath := range fileList {
		// try to get the metadata of the file
		metadata, err := exiftool.Extract(filepath)
		// if an error or no metadata, skip the file
		if err != nil || metadata == nil {
			continue
		}

		// try to pull the tags from the Subject field
		tags, err := metadata.Get("Subject")

		// if still no tags found, try to pull from the Category field
		if tags == nil {
			tags, err = metadata.Get("Category")
		}

		// if still no tags found, log and skip
		if tags == nil {
			log.Printf("**NO TAGS FOUND**    %s\n", filepath)
			continue
		}

		// otherwise, save to bolt
		tagsAsJSON, err := json.Marshal(tags)
		if err != nil {
			log.Printf("%s: %v\n", filepath, err)
			continue
		}

		// TODO save to bolt
		log.Println(string(tagsAsJSON))
	}
}
