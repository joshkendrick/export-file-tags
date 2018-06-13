// Author: Josh Kendrick
// Version: v0.0.3
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

	// produce the files to the channel for the consumers
	filepaths := make(chan string, 300)
	producedCount := 0
	go func() {
		filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
			filepaths <- path
			log.Printf("added path: %s", path)
			producedCount++

			return nil
		})
		close(filepaths)
	}()

	// number of processors
	consumerCount := 20
	// reporting channel
	done := make(chan int, consumerCount)

	// start the processors
	for index := 0; index < consumerCount; index++ {
		go tagsProcessor(filepaths, boltDB, done, index+1)
	}

	// wait for processors to finish
	consumedCount := 0
	for index := 0; index < consumerCount; index++ {
		consumedCount += <-done
	}

	log.Printf("produced: %d || consumed %d", producedCount, consumedCount)
}

func tagsProcessor(filepaths <-chan string, boltDB *bolt.DB, done chan<- int, id int) {
	count := 0

	// get a filepath
	for {
		filepath, more := <-filepaths
		if !more {
			log.Printf("%4d consumed %d files", id, count)
			done <- count
			return
		}

		count++

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
			log.Printf("******TAGS NOT FOUND****** %s", filepath)
			continue
		}

		// convert singleVal strings into an array
		// so all values in bolt database are the same format
		switch t := tags.(type) {
		case string:
			tags = []string{t}
		}

		log.Printf("%4d found tags - %s :: %v", id, filepath, tags)

		// marshal to json
		tagsAsJSON, err := json.Marshal(tags)
		if err != nil {
			log.Printf("%4d !!ERROR!! -- %v: %v", id, err, tags)
			continue
		}

		// save to bolt
		err = boltDB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("tags"))
			err := bucket.Put([]byte(filepath), tagsAsJSON)
			return err
		})

		if err == nil {
			log.Printf("%4d saved tags - %s :: %s", id, filepath, tags)
		} else {
			log.Printf("%4d !!ERROR!! -- %v: %s", id, err, filepath)
		}
	}
}
