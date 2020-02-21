package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/etcd-io/bbolt"
)

type DatabaseItem struct {
	Name string
	Len  float64
	T    int64
}

var (
	database *bbolt.DB
	dbbucket = []byte("main")
)

func database_init() {
	var err error
	database, err = bbolt.Open("./db.db", 0755, nil)
	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}
	database.Update(func(tx *bbolt.Tx) error {
		tx.CreateBucketIfNotExists(dbbucket)
		return nil
	})
}

func database_store(item *DatabaseItem) {
	encoded, _ := json.Marshal(item)
	database.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbbucket)
		bucket.Put(itob(int(item.T)), encoded)
		return nil

	})
}

func database_get(s, l string) []*DatabaseItem {

	si, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("database_get error: %v", err)
		return nil
	}
	li, err := strconv.Atoi(l)
	if err != nil {
		log.Printf("database_get error: %v", err)
		return nil
	}

	var rt []*DatabaseItem

	database.View(func(tx *bbolt.Tx) error {
		start := itob(si * 1000000000)
		end := itob(si + li*60*1000000000)

		bucket := tx.Bucket(dbbucket)
		cursor := bucket.Cursor()

		for k, v := cursor.Seek(start); k != nil; k, v = cursor.Next() {
			if bytes.Compare(k, end) < 0 {
				return nil
			}
			var item *DatabaseItem
			err := json.Unmarshal(v, &item)
			if err != nil {
				log.Printf("error %v", err)
				continue
			}
			rt = append(rt, item)

		}
		return nil
	})

	return rt
}

func database_worker() {
	for {
		current := time.Now().UnixNano() - int64(*flagTail*60*60*int(time.Nanosecond))
		err := database.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(dbbucket)
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				_ = v
				if bytes.Compare(k, itob(int(current))) < 0 {
					return nil
				}
				var item DatabaseItem
				err := json.Unmarshal(v, &item)
				if err != nil {
					return err
				}
				log.Printf("removing ./files/%v", item.Name)
				os.RemoveAll(fmt.Sprintf("./files/%v", item.Name))
				bucket.Delete(k)

			}
			return nil
		})
		if err != nil {
			log.Printf("error %v", err)
		}
		time.Sleep(time.Second)
	}
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
