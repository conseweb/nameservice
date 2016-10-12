package store

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type KVStore struct {
	db *sql.DB
}

func NewKV() (*KVStore, error) {
	fsPath := viper.GetString("farmer.fileSystemPath")
	db, err := sql.Open("sqlite3", filepath.Join(fsPath, "farmerKV.db"))
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	kv := &KVStore{
		db: db,
	}
	return kv, kv.initDB()
}

func (kv *KVStore) initDB() error {
	_, err := kv.db.Exec(`
CREATE TABLE IF NOT EXISTS kv (
	key VARCHAR(32) PRIMARY KEY,
	value VARCHAR(255)
)`)
	return err
}

func (kv *KVStore) Set(k, v string) error {
	_, err := kv.db.Exec("REPLACE INTO kv (key, value) VALUES (?, ?)", k, v)
	return err
}

func (kv *KVStore) Get(key string) (value string, exists bool) {
	err := kv.db.QueryRow("SELECT value FROM kv WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", false
	}
	return value, true
}
