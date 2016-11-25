package account

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func getDB() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	var err error
	db, err = sql.Open("sqlite3", fmt.Sprintf("/tmp/farmerKV-%v.db", time.Now().Unix()))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func TestDB(t *testing.T) {
	db, err := getDB()
	if err != nil {
		t.Errorf("get db failed, %s", err)
		return
	}

	if err = (*Contact).InitDB(nil, db); err != nil {
		t.Errorf("init table, %s", err)
		return
	}

	c := &Contact{
		Name:        "test",
		Email:       "test@test.com",
		Phone:       "12312341234",
		Addr:        "xxxyyyzzz",
		Tag:         "dev",
		Description: "for test user's contact.",
	}

	if err = c.Insert(db); err != nil {
		t.Errorf("insert failed, ", err)
		return
	}

	t.Logf("inserted. %+v", c)

	if err = c.Update(db, &Contact{Name: "newtest", Email: "a@a.com", Description: "hshsh"}); err != nil {
		t.Errorf("updated failed, %s", err)
		return
	}
	t.Logf("updated. %+v", c)

	cs, err := (*Contact).List(nil, db)
	if err != nil {
		t.Errorf("list failed, ", err)
		return
	}

	t.Logf("contacts: %+v", cs[0])
}
