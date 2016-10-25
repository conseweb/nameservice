package account

import (
	"github.com/conseweb/common/hdwallet"
	pb "github.com/conseweb/common/protos"
)

func (a *Account) NewLocalDevice() Device {
	wlt, _ := a.Wallet.Child(uint32(len(a.Devices)))
	return Device{
		Device: &pb.Device{
			UserID: a.ID,
			Mac:    getLocalMAC(),
			Os:     GetLocalOS(),
			For:    pb.DeviceFor_FARMER,

			Alias: GetLocalOS(),
			Wpub:  wlt.Pub().Serialize(),
			Spub:  []byte(""),
		},
		isLocal: true,
		Wallet:  wlt,
	}
}

type Device struct {
	*pb.Device

	isLocal bool
	Wallet  *hdwallet.HDWallet
}

func (d *Device) IsLocal() bool {
	return d.Mac == getLocalMAC()
}
