package farmer

import (
	pb "github.com/conseweb/common/protos"
	"github.com/hyperledger/fabric/farmer/api"
	"github.com/hyperledger/fabric/farmer/daemon"
	"github.com/hyperledger/fabric/flogging"
	"github.com/op/go-logging"
)

func StartFarmer() error {
	d, err := daemon.LoadDaemon()
	if err != nil {
		return err
	}

	api.Serve(d)
}

func StopFarmer() error {
	return nil
}
