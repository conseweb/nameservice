package account

import (
	"fmt"
	"math"

	"github.com/conseweb/common/assets/lepuscoin/client"
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

// invoke chaincode coinbase
func (d *Device) InvokeCoinBase() error {
	if d.Address == "" {
		return fmt.Errorf("invoke coinbase transaction failed, address is nil")
	}
	cli := client.NewTransactionV1("")
	cli.AddTxOut(client.NewTxOut(1000000, d.Address, 0))
	cli.AddTxIn(client.NewTxIn("", math.MaxUint32))
	bs, err := cli.Base64Bytes()
	if err != nil {
		logger.Errorf("encode base64 failed, error: %s", err.Error())
		return err
	}
	logger.Debugf("conbase tx: %s", bs)
	return nil
}
