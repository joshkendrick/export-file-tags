// Author: Josh Kendrick
// Version: v0.0.1
// Do whatever you want with this code

package main

import (
	"encoding/json"
	"io/ioutil"
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

	// get directories
	directories := []string{}
	filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})

	tags := make(chan tagsKV, 250)
	finished := make(chan bool)

	go tagsProcessor(tags, finished, boltDB)

	// start an exif reader per directory
	for _, directory := range directories {
		go tagsReader(tags, finished, directory)
	}

	// wait for the readers to finish
	for index := 0; index < len(directories); index++ {
		<-finished
	}

	// close the tags channel, not sending anymore
	close(tags)

	// wait for the processor to finish
	<-finished
}

func tagsReader(tagsChan chan<- tagsKV, finished chan<- bool, dirPath string) {
	// be sure to definitely report done
	defer func() {
		finished <- true
	}()

	// get everything in this directory
	fileNames, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("!!ERROR!!     %v: %s\n", err, dirPath)
		return
	}

	// loop through and build paths for non directories
	filePaths := []string{}
	for _, fileName := range fileNames {
		if !fileName.IsDir() {
			filePaths = append(filePaths, filepath.Join(dirPath, fileName.Name()))
		}
	}

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
			log.Printf("******TAGS NOT FOUND****** %s\n", filepath)
			continue
		}

		// convert singleVal strings into an array
		// so all values in bolt database are the same format
		switch t := tags.(type) {
		case string:
			tags = []string{t}
		}

		log.Printf("found   %s :: %v\n", filepath, tags)
		tagsChan <- tagsKV{filepath, tags}
	}
}

func tagsProcessor(tagsPipe <-chan tagsKV, finished chan<- bool, boltDB *bolt.DB) {
	count := 0

	// loop for tags received on the channel
	for {
		input, more := <-tagsPipe
		if !more {
			log.Printf("finished: %d tagsets processed\n", count)
			finished <- true
			close(finished)
			return
		}

		count++

		// marshal to json
		tagsAsJSON, err := json.Marshal(input.tags)
		if err != nil {
			log.Printf("!!ERROR!!     %v: %v\n", err, input)
			continue
		}

		// save to bolt
		err = boltDB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("tags"))
			err := bucket.Put([]byte(input.filePath), tagsAsJSON)
			return err
		})

		if err == nil {
			log.Printf("saved   %s :: %s\n", input.filePath, input.tags)
		} else {
			log.Printf("!!ERROR!!     %v: %s\n", err, input.filePath)
		}
	}
}
