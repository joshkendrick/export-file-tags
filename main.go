// Author: Josh Kendrick
// Do whatever you want with this code

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify parent directory")
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

	fileList := []string{}
	filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	for _, file := range fileList {
		fmt.Println(file)
	}
}
