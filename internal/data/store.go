package data

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// Store struct
type Store struct {
	db *bolt.DB
}

// OpenStore the store
func OpenStore(file string) (*Store, error) {
	db, err := bolt.Open(file, 0600, nil)
	s := &Store{
		db: db,
	}

	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("motions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return s, err
}

// Close the store
func (s *Store) Close() error {
	return s.db.Close()
}

// UpdateMotion updates motion
func (s *Store) UpdateMotion(item *Motion) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("motions"))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		item.ID = id

		// Marshal user data into bytes.
		buf, err := json.Marshal(item)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return b.Put(itob(item.ID), buf)
	})
}

// GetAllMotions returns all motions
func (s *Store) GetAllMotions() ([]Motion, error) {
	var result []Motion
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("motions"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			item := Motion{}
			err := json.Unmarshal(v, &item)
			// lets just skip there which can not be deserialized
			if err != nil {
				continue
			}
			result = append(result, item)
		}
		return nil
	})
	return result, err
}

// GetMotion retns all motions
func (s *Store) GetMotion(id uint64) (Motion, error) {
	result := Motion{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("motions"))
		v := b.Get(itob(id))
		err := json.Unmarshal(v, &result)
		if err != nil {
			return err
		}
		return nil
	})
	return result, err
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
