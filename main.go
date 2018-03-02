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

type tagsKV struct {
	filePath string
	tags     interface{}
}

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
	filePaths := []string{}
	filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		filePaths = append(filePaths, path)
		return nil
	})

	tags := make(chan tagsKV, 150)
	finished := make(chan bool)

	go tagsProcessor(tags, finished, boltDB)

	go tagsReader(tags, filePaths)

	<-finished
}

func tagsReader(tagsChan chan tagsKV, filePaths []string) {
	// loop through the files
	for _, filepath := range filePaths {
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

		log.Printf("sending -->> %s :: %v\n", filepath, tags)
		tagsChan <- tagsKV{filepath, tags}
	}

	close(tagsChan)
}

func tagsProcessor(tagsPipe chan tagsKV, finished chan bool, boltDB *bolt.DB) {
	// start a database transaction
	tx, err := boltDB.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	// get the bucket
	bucket := tx.Bucket([]byte("tags"))

	// loop for tags received on the channel
	for {
		input, more := <-tagsPipe
		if !more {
			log.Println("finished")
			finished <- true
			close(finished)
			return
		}

		log.Printf("received <<-- %s\n", input.filePath)

		// marshal to json
		tagsAsJSON, err := json.Marshal(input.tags)
		if err != nil {
			log.Printf("%v: %v\n", err, input)
			continue
		}

		// save to bolt
		bucket.Put([]byte(input.filePath), tagsAsJSON)
		tx.Commit()
		log.Printf("saved %s :: %s\n", input.filePath, tagsAsJSON)
	}
}
