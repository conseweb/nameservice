package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var (
	log = logging.MustGetLogger("indexer")
	orm *xorm.Engine
)

type FileInfo struct {
	ID       int64  `xorm:"pk autoincr 'id'" json:"id"`
	DeviceID string `xorm:"notnull index 'device_id'" json:"device_id"`
	Path     string `xorm:"notnull index 'path'" json:"path"`
	Hash     string `xorm:"notnull index 'hash'" json:"hash"`
	Size     int64  `xorm:"'size'" json:"hash"`

	Created time.Time `xorm:"created" json:"created"`
	Updated time.Time `xorm:"updated" json:"updated"`
}

type Device struct {
	ID      string `xorm:"pk autoincr" json:"id"`
	Address string `xorm:"notnull index" json:"address"`
}

func InitDB() (*xorm.Engine, error) {
	if orm != nil {
		if err := orm.Ping(); err != nil {
			return nil, err
		}
		return orm, nil
	}

	path := viper.GetString("peer.fileSystemPath")
	fi, err := os.Stat(filepath.Join(path, "indexer.db"))
	var Orm *xorm.Engine
	if err != nil && os.IsExist(err) {
		return nil, err
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s is a directory", path)
	}

	Orm, err = xorm.NewEngine("sqlite3", path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	Orm.ShowSQL(true)

	Orm.Sync2(&FileInfo{}, &Device{})

	orm = Orm

	return Orm, nil
}
