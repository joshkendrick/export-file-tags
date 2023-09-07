// Author: Josh Kendrick
// Version: v0.1.1
// License: do whatever you want with this code

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/barasher/go-exiftool"
	_ "github.com/mattn/go-sqlite3"
)

const EXIFTOOL_PATH = "./exiftool.exe"
const DB_NAME = "media-tags.db"

// struct to hold pieces of a db statement
type Statement struct {
	query string
	args  []any
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify a directory")
	}
	// parent dir
	directory := os.Args[1]

	// produce the files to the channel for the consumer
	filepaths := make(chan string)
	producedCount := 0
	go func() {
		filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
			filepaths <- path
			log.Printf("added path: %s", path)
			producedCount++

			return nil
		})
		// done adding files
		close(filepaths)
	}()

	// reporting channel
	done := make(chan int, 1)
	// db statements channel
	dbStmts := make(chan Statement)

	// start the processor
	go tagsProcessor(filepaths, done, dbStmts)

	// db status channel
	dbWriterDone := make(chan bool, 1)
	// start the database writer
	go tagsWriter(dbStmts, dbWriterDone)

	// wait for processor to finish
	consumedCount := <-done
	// processor is finished, close the database channel
	close(dbStmts)

	// wait for the dbWriter to finish
	<-dbWriterDone

	// print results
	log.Printf("produced: %d || consumed %d", producedCount, consumedCount)
}

func tagsProcessor(filepaths <-chan string, done chan<- int, dbStmts chan<- Statement) {
	count := 0

	// create the exifReader
	// this isnt flexible, as is will only work on windows with exiftool.exe in same location as execution
	exifReader, err := exiftool.NewExiftool(exiftool.SetExiftoolBinaryPath(EXIFTOOL_PATH))
	if err != nil {
		log.Printf("!!ERROR!! -- %v", err)
	}

	// get a filepath
	for {
		filenameAbs, more := <-filepaths
		if !more {
			log.Printf("consumed %d files", count)
			done <- count
			return
		}
		count++

		// get the filename
		fileNameRel := filepath.Base(filenameAbs)

		// get the metadata of the file
		// there should only be one FileInfo since we call for one filepath
		fileInfo := exifReader.ExtractMetadata(filenameAbs)[0]
		if fileInfo.Err != nil {
			log.Printf("!!ERROR!! -- %v: %v", fileNameRel, fileInfo.Err)
			continue
		}

		// try to pull the tags from the Subject field
		tags, _ := fileInfo.GetStrings("Subject")

		// if still no tags found, try to pull from the Category field
		if tags == nil || len(tags) < 1 {
			tags, _ = fileInfo.GetStrings("Category")
		}

		// if still no tags found, log and skip
		if tags == nil || len(tags) < 1 {
			log.Printf("******TAGS NOT FOUND****** %s", filenameAbs)
			continue
		}

		log.Printf("found tags: %s :: %v", filenameAbs, tags)

		// marshal to json
		tagsAsJSON, err := json.Marshal(tags)
		if err != nil {
			log.Printf("!!ERROR!! -- %v: %v", err, tags)
			continue
		}

		// push a statement to write the file we're processing
		statement := Statement{"INSERT OR REPLACE INTO files (filename, path, tags_json) VALUES (?, ?, ?)", []any{fileNameRel, filenameAbs, string(tagsAsJSON)}}
		dbStmts <- statement

		// this map will track tags we already inserted to try to save sql statements
		createdTags := make(map[string]bool)
		// loop through all the tags
		for _, tag := range tags {
			_, exists := createdTags[tag]
			if !exists {
				statement = Statement{"INSERT OR IGNORE INTO tags (tag) VALUES (?)", []any{tag}}
				dbStmts <- statement
				createdTags[tag] = true
				statement = Statement{"INSERT INTO file_tags (file, tag) VALUES (?, ?)", []any{fileNameRel, tag}}
				dbStmts <- statement
			}
		}
	}
}

func tagsWriter(dbStmts <-chan Statement, done chan<- bool) {
	// open the database, get an object for writing
	db, err := sql.Open("sqlite3", "file:"+DB_NAME+"?_recursive_triggers=true&_foreign_keys=true")
	if err != nil {
		log.Printf("!!ERROR!! -- tagsWriter couldnt open database: %v", err)
		done <- false
		close(done)
		return
	}

	for { // loop
		// until there's no more database statements to run
		statement, more := <-dbStmts
		if !more {
			db.Close()
			log.Printf("all db statements written")
			done <- true
			close(done)
			return
		}

		// write the statement
		_, err = db.Exec(statement.query, statement.args...)
		// if there's an error, stop the run, db errors should be addressed
		if err != nil {
			db.Close()
			log.Printf("!!ERROR!! -- database statement failed: %v", err)
			done <- false
			close(done)
			return
		}
	}
}
