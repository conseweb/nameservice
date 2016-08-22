package daemon

import (
	"fmt"
	"io/ioutil"
	"os"
)

func LoadDaemon() (*Daemon, error) {
	if std != nil {
		return std, nil
	}

	// Check farmer root path.
	_, err := os.Stat(DefaultFarmerPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(DefaultFarmerPath, 0755); err != nil {
			return nil, fmt.Errorf("can not enable mkdir %s,errpr: %s", DefaultFarmerPath, err.Error())
		}
	}

	// Check farmer pid file.
	pidf, err := os.Open(DefaultPidFile)
	if os.IsExist(err) {
		pidbs, err := ioutil.ReadAll(pidf)
		if err != nil {
			return nil, fmt.Errorf("Open pid file %s faild, error: %s", DefaultPidFile, err.Error())
		}
		return nil, fmt.Errorf("Farmer daemon is running(PID: %s)", pidbs)
	}

	addr := viper.GetString("daemon.address")
	if addr == "" {
		addr = DefaultDaemonAddress
	}

	d := NewDaemon()

	if err := d.Init(); err != nil {
		return nil, err
	}

	std = d
	return d, nil
}
