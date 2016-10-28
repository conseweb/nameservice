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
			Wpub:  []byte(wlt.Pub().String()),
			Spub:  []byte("ffff"),
		},
		IsLocal: true,
		Wallet:  wlt,
		Address: wlt.Pub().Address(),
	}
}

type Device struct {
	*pb.Device

	IsLocal bool `json:"is_local"`
	Wallet  *hdwallet.HDWallet
	Address string `json:"address"` // wallet.Address
}
