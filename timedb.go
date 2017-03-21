package main

import (
	"crypto/md5"
	"fmt"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

func InitTimeDB(directory string) (*bolt.DB, error) {
	dbFile := path.Join(directory, "ticktick.db")
	db, err := bolt.Open(dbFile, 0644, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open time database: %s", err)
	}
	return db, nil
}

func EnsureBucket(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("lastRuns"))
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to init time database: %s", err)
	}
	return nil
}

func GetTaskHash(td *TaskDefinition) []byte {
	digest := md5.New()
	digest.Write([]byte(td.Name))
	digest.Write([]byte(td.Rule))
	return digest.Sum(nil)
}

func StoreLastRunTime(db *bolt.DB, td *TaskDefinition, runTime time.Time) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastRuns"))
		val, _ := runTime.GobEncode()
		return b.Put(GetTaskHash(td), val)
	})
	return err
}

func GetLastRunTime(db *bolt.DB, td *TaskDefinition) time.Time {
	var result time.Time
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastRuns"))
		val := b.Get(GetTaskHash(td))
		if val != nil {
			result.GobDecode(val)
		}
		return nil
	})
	return result
}

func GetLastRunTimeOr(db *bolt.DB, td *TaskDefinition, def time.Time) time.Time {
	t := GetLastRunTime(db, td)
	if t.Before(def) {
		return def
	}
	return t
}
