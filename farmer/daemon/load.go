package daemon

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

func LoadDaemon() (*Daemon, error) {
	if std != nil {
		return std, nil
	}

	fsPath := viper.GetString("peer.fileSystemPath")
	// Check farmer root path.
	_, err := os.Stat(fsPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(fsPath, 0755); err != nil {
			return nil, fmt.Errorf("can not enable mkdir %s,errpr: %s", fsPath, err.Error())
		}
	}

	pidFile := filepath.Join(fsPath, "farmer.pid")
	// Check farmer pid file.
	pidf, err := os.Open(pidFile)
	if os.IsExist(err) {
		pidbs, err := ioutil.ReadAll(pidf)
		if err != nil {
			return nil, fmt.Errorf("Open pid file %s faild, error: %s", pidFile, err.Error())
		}
		return nil, fmt.Errorf("Farmer daemon is running(PID: %s)", pidbs)
	}

	addr := viper.GetString("daemon.address")
	if addr == "" {
		addr = DefaultListenAddr
	}

	d := NewDaemon()

	if err := d.Init(); err != nil {
		log.Errorf("daemon init failed,, %s", err.Error())
		return nil, err
	}

	std = d
	return d, nil
}
