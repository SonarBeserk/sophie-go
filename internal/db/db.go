package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	statsBucket string = "STATS"
)

type Database struct {
	*bolt.DB
}

func OpenOrConfigureDatabase(databaseFile string) (*Database, error) {
	db, err := bolt.Open(databaseFile, 0666, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading database file %s: %v", databaseFile)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(statsBucket))
		if err != nil {
			return errors.Wrapf(err, "could not create root bucket: %v")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not set up buckets, %v")
	}

	return &Database{
		DB: db,
	}, nil
}

func (d Database) GetEmoteSentUsage(emote string, userID string) (int, error) {
	count := 0

	err := d.View(func(tx *bolt.Tx) error {
		key := strings.ToUpper(emote) + "|" + strings.ToUpper(userID) + "|Sent"

		curCountVal := tx.Bucket([]byte(statsBucket)).Get([]byte(key))

		if curCountVal != nil {
			curCount, err := strconv.Atoi(string(curCountVal))
			if err != nil {
				return err
			}

			count = curCount
		}

		return nil
	})

	return count, err
}

func (d Database) GetEmoteReceivedUsage(emote string, userID string) (int, error) {
	count := 0

	err := d.View(func(tx *bolt.Tx) error {
		key := strings.ToUpper(emote) + "|" + strings.ToUpper(userID) + "|Received"

		curCountVal := tx.Bucket([]byte(statsBucket)).Get([]byte(key))

		if curCountVal != nil {
			curCount, err := strconv.Atoi(string(curCountVal))
			if err != nil {
				return err
			}

			count = curCount
		}

		return nil
	})

	return count, err
}

func (d Database) SetEmoteSentUsage(emote string, userID string, count int) error {
	err := d.Update(func(tx *bolt.Tx) error {
		key := strings.ToUpper(emote) + "|" + strings.ToUpper(userID) + "|Sent"
		countStr := strconv.Itoa(count)
		err := tx.Bucket([]byte(statsBucket)).Put([]byte(key), []byte(countStr))
		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
	})
	return err
}

func (d Database) SetEmoteReceivedUsage(emote string, userID string, count int) error {
	err := d.Update(func(tx *bolt.Tx) error {
		key := strings.ToUpper(emote) + "|" + strings.ToUpper(userID) + "|Received"
		countStr := strconv.Itoa(count)
		err := tx.Bucket([]byte(statsBucket)).Put([]byte(key), []byte(countStr))
		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
	})
	return err
}

func (d Database) GetEmoteCountsForUser(emote string, userID string) (sent int, received int, err error) {
	sentCount, err := d.GetEmoteSentUsage(emote, userID)
	if err != nil {
		return 0, 0, err
	}

	err = d.SetEmoteSentUsage(emote, userID, sentCount)
	if err != nil {
		return 0, 0, err
	}

	receivedCount, err := d.GetEmoteReceivedUsage(emote, userID)
	if err != nil {
		return 0, 0, err
	}

	err = d.SetEmoteSentUsage(emote, userID, receivedCount)
	if err != nil {
		return 0, 0, err
	}

	return sentCount, receivedCount, nil
}
