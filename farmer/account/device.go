package account

import (
	"github.com/conseweb/common/hdwallet"
)

type Device struct {
	ID  string
	Mac string

	Wallet *hdwallet.HDWallet
}

func LocalDevice(wlt *hdwallet.HDWallet) *Device {
	return &Device{
		Mac:    getLocalMAC(),
		Wallet: wlt,
	}
}
