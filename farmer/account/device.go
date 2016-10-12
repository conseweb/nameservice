package account

import (
	"github.com/conseweb/common/hdwallet"
	pb "github.com/conseweb/common/protos"
)

func NewDevice(a *Account) Device {
	wlt, _ := a.Wallet.Child(uint32(len(a.Devices)))
	return Device{
		Device: &pb.Device{
			UserID: a.ID,
			Mac:    getLocalMAC(),
			Os:     GetLocalOS(),
			For:    pb.DeviceFor_FARMER,

			Alias: GetLocalOS(),
			// Wpub:   a.Wallet.Child(len(a.Devices)).Serialize(),
			Spub: []byte(""),
		},
		Wallet: wlt,
	}
}

type Device struct {
	*pb.Device

	ID  string
	Mac string

	Wallet *hdwallet.HDWallet
}
