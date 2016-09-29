package farmer

import (
	"github.com/hyperledger/fabric/farmer/api"
	"github.com/hyperledger/fabric/farmer/daemon"
	// "github.com/hyperledger/fabric/flogging"
	// pb "github.com/conseweb/common/protos"
	// "github.com/op/go-logging"
)

func StartFarmer() {
	d, err := daemon.LoadDaemon()
	if err != nil {
		return
	}

	go func() {
		if err := d.StartPeer(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := api.Serve(d); err != nil {
			panic(err)
		}
	}()
	d.WaitExit()
}

func StopFarmer() error {

	return nil
}
