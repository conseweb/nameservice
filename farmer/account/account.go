package account

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/conseweb/common/hdwallet"
	pb "github.com/conseweb/common/protos"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	// "golang.org/x/crypto/bcrypt"
	// "google.golang.org/grpc"
)

const (
	defaultTimeout = time.Second * 3
)

var (
	logger = logging.MustGetLogger("farmer")
)

var LanguageSupport = map[string]pb.PassphraseLanguage{
	"English":  pb.PassphraseLanguage_English,
	"简体中文":     pb.PassphraseLanguage_SimplifiedChinese,
	"繁體中文":     pb.PassphraseLanguage_TraditionalChinese,
	"日本語":      pb.PassphraseLanguage_JAPANESE,
	"español":  pb.PassphraseLanguage_SPANISH,
	"français": pb.PassphraseLanguage_FRENCH,
	"ITALIAN":  pb.PassphraseLanguage_ITALIAN,
}

type Account struct {
	ID       string                `json:"id"`
	NickName string                `json:"nicename"`
	Phone    string                `json:"phone"`
	Email    string                `json:"email"`
	Password string                `json:"password"`
	Lang     pb.PassphraseLanguage `json:"lang"`

	Passphrase string `json:"passphrase"`
	Wallet     *hdwallet.HDWallet

	Devices []Device `json:"devices"`

	logger *logging.Logger
}

func NewAccount(nickname, phone, email, pass, lang string) *Account {
	language := pb.PassphraseLanguage_English
	if l, ok := LanguageSupport[lang]; ok {
		language = l
	}

	dev_mode := viper.GetBool("daemon.dev")
	ph, hd := hdwallet.NewHDWallet(pass, language, dev_mode)
	return &Account{
		Phone:    phone,
		Email:    email,
		NickName: nickname,
		Password: pass,
		Lang:     language,

		Passphrase: ph,
		Wallet:     hd,
	}
}

func LoadFromFile() (*Account, error) {
	fpath := filepath.Join(viper.GetString("peer.fileSystemPath"), "farmerAccount.json")

	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	a := new(Account)
	err = json.NewDecoder(f).Decode(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Account) Save() error {
	fpath := filepath.Join(viper.GetString("peer.fileSystemPath"), "farmerAccount.json")
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("  ", "")
	err = encoder.Encode(a)
	if err != nil {
		return err
	}

	return nil
}

func (a *Account) Registry(idpCli pb.IDPPClient) error {
	priv, err := primitives.NewECDSAKey()
	if err != nil {
		return err
	}

	pubraw, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}

	regUser := &pb.RegisterUserReq{
		SignUpType: pb.SignUpType_MOBILE,
		SignUp:     a.Phone,
		Nick:       a.NickName,
		Pass:       a.Password,
		UserType:   pb.UserType_NORMAL,

		Wpub: []byte(a.Wallet.Pub().String()),
		Spub: pubraw,
		Sign: []byte("ffff"),
	}
	if ok := checkPhone(regUser.SignUp); !ok {
		if ok := checkEmail(a.Email); !ok {
			return fmt.Errorf("Email and phone need at least one.")
		}
		regUser.SignUpType = pb.SignUpType_EMAIL
		regUser.SignUp = a.Email
	}

	logger.Debugf("reg user: %#v", regUser)

	resp, err := idpCli.RegisterUser(context.Background(), regUser)
	if err != nil {
		return err
	}
	if resp.GetError() != nil && resp.GetError().ErrorType != pb.ErrorType_NONE_ERROR {
		return resp.GetError()
	}

	ru := resp.GetUser()
	if ru == nil {
		return fmt.Errorf("got nil user.")
	}
	a.ID = ru.UserID

	// bind local device.
	if err := a.BindLocalDevice(idpCli); err != nil {
		logger.Errorf("bind failed, %v", err.Error())
		return err
	}
	logger.Debugf("account %+v", a)

	err = a.Save()
	if err != nil {
		logger.Errorf("account save local file failed, %v", err.Error())
		return err
	}
	a.Wallet.Address()
	return nil
}

func Login(idpCli pb.IDPPClient, typ pb.SignInType, signup, password string) (a *Account, err error) {
	req := &pb.LoginUserReq{
		SignInType: typ,
		SignIn:     signup,
		Password:   password,
		Sign:       []byte("ffff"),
	}

	logger.Debugf("login user: %#v", req)
	resp, err := idpCli.LoginUser(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if resp.GetError() != nil && resp.GetError().ErrorType != pb.ErrorType_NONE_ERROR {
		logger.Errorf("login failed, %v", err.Error())
		return nil, resp.GetError()
	}

	ru := resp.GetUser()
	if ru == nil {
		return nil, fmt.Errorf("got nil user.")
	}
	a = &Account{
		ID:       ru.UserID,
		Phone:    ru.Mobile,
		Email:    ru.Email,
		NickName: ru.Nick,
		Devices:  []Device{},
	}
	exiLocal := false
	for _, device := range ru.Devices {
		dwlt, err := hdwallet.ParseStringWallet(string(device.Wpub))
		if err != nil {
			logger.Errorf("parse devices' wallet failed, %s", err.Error())
			continue
		}
		var addr string
		if dwlt != nil {
			addr = dwlt.Address()
		}

		if getLocalMAC() == device.Mac {
			a.Devices = append(a.Devices, Device{Device: device, IsLocal: true, Address: addr})
			exiLocal = true
		} else {
			a.Devices = append(a.Devices, Device{Device: device, Address: addr})
		}
		logger.Debugf("devices... %#v", a.Devices)
	}
	if !exiLocal {
		// try to bind device.
		if err := a.BindLocalDevice(idpCli); err != nil {
			logger.Errorf("bind failed, %v", err.Error())
			return a, err
		}
	}

	err = a.Save()
	if err != nil {
		logger.Errorf("account save local file failed, %v", err.Error())
		return a, err
	}

	return a, nil
}

func (a *Account) BindLocalDevice(idpCli pb.IDPPClient) error {
	priv, err := primitives.NewECDSAKey()
	if err != nil {
		return err
	}

	pubraw, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}

	dv := a.NewLocalDevice()
	devReq := &pb.BindDeviceReq{
		UserID: a.ID,
		Os:     dv.Os,
		For:    pb.DeviceFor_FARMER,
		Mac:    dv.Mac,
		Alias:  dv.Alias,

		Wpub: []byte(dv.Wpub),
		// device signature public key
		Spub: pubraw,

		Sign: []byte("ffff"),
	}

	resp, err := idpCli.BindDeviceForUser(context.Background(), devReq)
	if err != nil {
		logger.Errorf("BindDevice failed, %v", err.Error())
		return err
	}

	a.Devices = append(a.Devices, Device{Device: resp.Device, Wallet: dv.Wallet, IsLocal: true})
	return nil
}

func (a *Account) Logout() error {
	return nil
}

func (a *Account) Online(client pb.FarmerPublicClient) error {
	if a.ID == "" {
		return fmt.Errorf("account id required")
	}

	onlineReq := &pb.FarmerOnLineReq{FarmerID: a.ID}
	logger.Debugf("login with %+v", a)

	onlineRes, err := client.FarmerOnLine(context.Background(), onlineReq)
	if err != nil {
		return err
	}
	if onlineRes.Error != nil {
		logger.Errorf("login error: %#v", onlineRes.Error)
		return onlineRes.Error
	}

	return nil
}

func (a *Account) getSignInType() (st pb.SignInType, sv string) {
	st, sv = pb.SignInType_SI_MOBILE, a.Phone
	if a.ID != "" {
		st, sv = pb.SignInType_SI_USERID, a.ID
	} else if a.Email != "" {
		st, sv = pb.SignInType_SI_EMAIL, a.Email
	}
	return
}
