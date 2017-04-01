package base

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

// userBucket for users
var usersBucket = []byte("users")

// sessionBucket for session
var sessionsBucket = []byte("sessions")

// bucketsList for bucket
var bucketsList = [][]byte{sessionsBucket, usersBucket}

// ErrNoRows for no row in result
var ErrNoRows = errors.New("db: no rows in result set")

// ErrDuplicateRow for duplicate row in result
var ErrDuplicateRow = errors.New("db: duplicate row found for unique constraint")

// DB Wrapper for bolt db. This allows us to attach methods
// to the db object.
type DB struct {
	*bolt.DB
}

// CreateAllBuckets create all buckets for project
func (db *DB) CreateAllBuckets() error {
	err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range bucketsList {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// TimeNow return the time now
func TimeNow() time.Time {
	return time.Now().UTC()
}
