package account

import (
	"database/sql"
	"fmt"
)

type Contact struct {
	Id          int    `json:"id" sql:"id"`
	Name        string `json:"name" sql:"name"`
	Email       string `json:"email" sql:"email"`
	Phone       string `json:"phone" sql:"phone"`
	Addr        string `json:"addr" sql:"addr"`
	Tag         string `json:"tag" sql:"tag"`
	Description string `json:"description" sql:"description"`
}

func (c *Contact) InitDB(db *sql.DB) error {
	sqlstr := `
	CREATE TABLE IF NOT EXISTS 'contacts' (
		'id' INTEGER PRIMARY KEY AUTOINCREMENT,
		'name' VARCHAR(32) NOT NULL,
		'email' VARCHAR(32) UNIQUE,
		'phone' VARCHAR(16) UNIQUE,
		'addr' VARCHAR(64) UNIQUE,
		'tag' VARCHAR(16),
		'description' VARCHAR(255)
	)`
	if _, err := db.Exec(sqlstr); err != nil {
		logger.Errorf("create table contacts failed, %s", err)
		return err
	}

	return nil
}

// used by (*Contact).List
func (c *Contact) List(db *sql.DB) ([]*Contact, error) {
	query := `SELECT 
id,
name,
email,
phone,
addr,
tag,
description FROM contacts`
	rows, err := db.Query(query)
	if err != nil {
		logger.Errorf("query contact failed, %s", err)
		return nil, err
	}

	cs := []*Contact{}

	for rows.Next() {
		ret := &Contact{}
		if err = rows.Scan(&ret.Id, &ret.Name, &ret.Email, &ret.Phone, &ret.Addr, &ret.Tag, &ret.Description); err != nil {
			return nil, err
		}
		cs = append(cs, ret)
	}

	return cs, nil
}

// used by (*Contact).Get
func (c *Contact) Get(db *sql.DB, id int) (*Contact, error) {
	ret := &Contact{}
	query := `SELECT * FROM contacts WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&ret.Id, &ret.Name, &ret.Email, &ret.Phone, &ret.Addr, &ret.Tag, &ret.Description)
	if err != nil {
		logger.Errorf("query contact<%v> failed, %s", id, err)
		return nil, err
	}

	return ret, nil
}

func (c *Contact) Update(db *sql.DB, n *Contact) error {
	sqlstr := `
UPDATE contacts SET 
name = ?,
email = ?,
phone = ?,
addr = ?,
tag = ?,
description = ?
WHERE id = ?
`

	if _, err := db.Exec(sqlstr, n.Name, n.Email, n.Phone, n.Addr, n.Tag, n.Description, c.Id); err != nil {
		logger.Errorf("update<%+v> to <%+v> failed, %s", c, n, err)
		return err
	}
	c.Name = n.Name
	c.Email = n.Email
	c.Phone = n.Phone
	c.Addr = n.Addr
	c.Tag = n.Tag
	c.Description = n.Description
	return nil
}

func (c *Contact) Insert(db *sql.DB) error {
	sqlstr := `
INSERT INTO contacts (
	name,
	email,
	phone,
	addr,
	tag,
	description
) VALUES(?, ?, ?, ?, ?, ?)
`
	if _, err := db.Exec(sqlstr, c.Name, c.Email, c.Phone, c.Addr, c.Tag, c.Description); err != nil {
		logger.Errorf("insert %+v into table contacts failed, %s", c, err)
		return err
	}

	return c.loadId(db)
}

func (c *Contact) Remove(db *sql.DB, id int) error {
	sqlstr := `
	DELETE FROM contacts WHERE id = ?;
`
	if _, err := db.Exec(sqlstr, id); err != nil {
		logger.Errorf("remove %+v from contacts failed, %s", id, err)
		return err
	}

	return nil
}

func (c *Contact) RemoveAll(db *sql.DB) error {
	sqlstr := `DELETE FROM contacts`
	if _, err := db.Exec(sqlstr, c.Id); err != nil {
		logger.Errorf("remove all from contacts failed, %s", err)
		return err
	}

	return nil
}

func (c *Contact) loadId(db *sql.DB) error {
	if c.Email == "" || c.Phone == "" {
		return fmt.Errorf("need email or phone")
	}

	query := `SELECT id FROM contacts WHERE email = ? and phone = ?`
	err := db.QueryRow(query, c.Email, c.Phone).Scan(&c.Id)
	if err != nil {
		logger.Errorf("query contact failed, %s", err)
		return err
	}

	return nil
}
