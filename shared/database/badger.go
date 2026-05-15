package database

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var main *DB

type DB struct {
	b *badger.DB
}

func Init(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}

	main = &DB{b: db}
	return nil
}

func New(path string) (*DB, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions(path).WithLogger(nil).WithValueLogFileSize(1 << 20)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &DB{b: db}, nil
}

func Get() *DB {
	return main
}

func (db *DB) Close() error {
	return db.b.Close()
}

func (db *DB) Set(key string, value []byte) error {
	return db.b.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

func (db *DB) SetTTL(key string, value []byte, ttl time.Duration) error {
	return db.b.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte(key), value).WithTTL(ttl))
	})
}

func (db *DB) SetTTLJSON(key string, v interface{}, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return db.SetTTL(key, data, ttl)
}

func (db *DB) Get(key string) ([]byte, error) {
	var value []byte
	err := db.b.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

func (db *DB) Delete(key string) error {
	return db.b.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (db *DB) Has(key string) bool {
	err := db.b.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})
	return err == nil
}

func (db *DB) SetJSON(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return db.Set(key, data)
}

func (db *DB) GetJSON(key string, v interface{}) error {
	data, err := db.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (db *DB) Keys(prefix string) []string {
	var keys []string
	db.b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = []byte(prefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			keys = append(keys, string(it.Item().Key()))
		}
		return nil
	})
	return keys
}

func (db *DB) Scan(prefix string, fn func(key string, value []byte) error) error {
	return db.b.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				return fn(string(item.Key()), val)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func Set(key string, value []byte) error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.Set(key, value)
}

func SetTTL(key string, value []byte, ttl time.Duration) error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.SetTTL(key, value, ttl)
}

func Key(key string) ([]byte, error) {
	if main == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return main.Get(key)
}

func Delete(key string) error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.Delete(key)
}

func Has(key string) bool {
	if main == nil {
		return false
	}
	return main.Has(key)
}

func SetJSON(key string, v interface{}) error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.SetJSON(key, v)
}

func GetJSON(key string, v interface{}) error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.GetJSON(key, v)
}

func Keys(prefix string) []string {
	if main == nil {
		return nil
	}
	return main.Keys(prefix)
}

func Scan(prefix string, fn func(key string, value []byte) error) error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.Scan(prefix, fn)
}

func Close() error {
	if main == nil {
		return fmt.Errorf("database not initialized")
	}
	return main.Close()
}
