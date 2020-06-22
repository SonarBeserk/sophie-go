package main

import (
	"fmt"
	"strconv"
	"strings"

	bolt "go.etcd.io/bbolt"
)

var (
	statsBucket string = "STATS"
)

func openOrConfigureDatabase(databaseFile string) (*bolt.DB, error) {
	db, err := bolt.Open(databaseFile, 0666, nil)
	if err != nil {
		fmt.Printf("error loading database file %s: %v", databaseFile, err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(statsBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	return db, nil
}

func getEmoteSentUsage(emote string, userID string) (string, error) {
	count := 0

	err := db.View(func(tx *bolt.Tx) error {
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

	countStr := strconv.Itoa(count)
	return countStr, err
}

func getEmoteReceivedUsage(emote string, userID string) (string, error) {
	count := 0

	err := db.View(func(tx *bolt.Tx) error {
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

	countStr := strconv.Itoa(count)
	return countStr, err
}

func setEmoteSentUsage(emote string, userID string, count string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		key := strings.ToUpper(emote) + "|" + strings.ToUpper(userID) + "|Sent"
		err := tx.Bucket([]byte(statsBucket)).Put([]byte(key), []byte(count))
		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
	})
	return err
}

func setEmoteReceivedUsage(emote string, userID string, count string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		key := strings.ToUpper(emote) + "|" + strings.ToUpper(userID) + "|Received"
		err := tx.Bucket([]byte(statsBucket)).Put([]byte(key), []byte(count))
		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
	})
	return err
}

func getEmoteCountsForUser(emote string, userID string) (sent string, received string, err error) {
	sentCount, err := getEmoteSentUsage(smugKey, userID)
	if err != nil {
		return "", "", err
	}

	err = setEmoteSentUsage(smugKey, userID, sentCount)
	if err != nil {
		return "", "", err
	}

	receivedCount, err := getEmoteReceivedUsage(smugKey, userID)
	if err != nil {
		return "", "", err
	}

	err = setEmoteSentUsage(smugKey, userID, receivedCount)
	if err != nil {
		return "", "", err
	}

	return sentCount, receivedCount, nil
}
