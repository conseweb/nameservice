package hdwallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"github.com/conseweb/common/passphrase"
	"github.com/conseweb/common/protos"
)

var (
	//MainNet
	version_mainnet_pub []byte
	version_mainnet_pri []byte
	//TestNet
	version_testnet_pub []byte
	version_testnet_pri []byte
)

func init() {
	version_mainnet_pub, _ = hex.DecodeString("0488B21E")
	version_mainnet_pri, _ = hex.DecodeString("0488ADE4")
	version_testnet_pub, _ = hex.DecodeString("043587CF")
	version_testnet_pri, _ = hex.DecodeString("04358394")
}

// HDWallet defines the components of a hierarchical deterministic wallet
type HDWallet struct {
	version     []byte //4 bytes
	depth       uint16 //1 byte
	fingerprint []byte //4 bytes
	childnumber []byte //4 bytes
	chaincode   []byte //32 bytes
	key         []byte //33 bytes
}

func NewHDWallet(pass string, lang protos.PassphraseLanguage) (string, *HDWallet) {
	ph, _ := passphrase.Passphrase(256, lang)
	seed := passphrase.NewSeed(ph, pass)
	return ph, MasterKey(seed)
}

// Child returns the ith child of wallet w. Values of i >= 2^31
// signify private key derivation. Attempting private key derivation
// with a public key will throw an error.
func (w *HDWallet) Child(i uint32) (*HDWallet, error) {
	var fingerprint, I, newkey []byte
	switch {
	case bytes.Compare(w.version, version_mainnet_pri) == 0, bytes.Compare(w.version, version_testnet_pri) == 0:
		pub := privToPub(w.key)
		mac := hmac.New(sha512.New, w.chaincode)
		if i >= uint32(0x80000000) {
			mac.Write(append(w.key, uint32ToByte(i)...))
		} else {
			mac.Write(append(pub, uint32ToByte(i)...))
		}
		I = mac.Sum(nil)
		iL := new(big.Int).SetBytes(I[:32])
		if iL.Cmp(curve.N) >= 0 || iL.Sign() == 0 {
			return &HDWallet{}, errors.New("Invalid Child")
		}
		newkey = addPrivKeys(I[:32], w.key)
		fingerprint = hash160(privToPub(w.key))[:4]

	case bytes.Compare(w.version, version_mainnet_pub) == 0, bytes.Compare(w.version, version_testnet_pub) == 0:
		mac := hmac.New(sha512.New, w.chaincode)
		if i >= uint32(0x80000000) {
			return &HDWallet{}, errors.New("Can't do Private derivation on Public key!")
		}
		mac.Write(append(w.key, uint32ToByte(i)...))
		I = mac.Sum(nil)
		iL := new(big.Int).SetBytes(I[:32])
		if iL.Cmp(curve.N) >= 0 || iL.Sign() == 0 {
			return &HDWallet{}, errors.New("Invalid Child")
		}
		newkey = addPubKeys(privToPub(I[:32]), w.key)
		fingerprint = hash160(w.key)[:4]
	}
	return &HDWallet{w.version, w.depth + 1, fingerprint, uint32ToByte(i), I[32:], newkey}, nil
}

// Serialize returns the serialized form of the wallet.
func (w *HDWallet) Serialize() []byte {
	depth := uint16ToByte(uint16(w.depth % 256))
	//bindata = vbytes||depth||fingerprint||i||chaincode||key
	bindata := append(w.version, append(depth, append(w.fingerprint, append(w.childnumber, append(w.chaincode, w.key...)...)...)...)...)
	chksum := dblSha256(bindata)[:4]
	return append(bindata, chksum...)
}

// String returns the base58-encoded string form of the wallet.
func (w *HDWallet) String() string {
	return base58.Encode(w.Serialize())
}

// StringWallet returns a wallet given a base58-encoded extended key
func StringWallet(data string) (*HDWallet, error) {
	dbin := base58.Decode(data)
	if err := ByteCheck(dbin); err != nil {
		return &HDWallet{}, err
	}
	if bytes.Compare(dblSha256(dbin[:(len(dbin) - 4)])[:4], dbin[(len(dbin)-4):]) != 0 {
		return &HDWallet{}, errors.New("Invalid checksum")
	}
	vbytes := dbin[0:4]
	depth := byteToUint16(dbin[4:5])
	fingerprint := dbin[5:9]
	i := dbin[9:13]
	chaincode := dbin[13:45]
	key := dbin[45:78]
	return &HDWallet{vbytes, depth, fingerprint, i, chaincode, key}, nil
}

// Pub returns a new wallet which is the public key version of w.
// If w is a public key, Pub returns a copy of w
func (w *HDWallet) Pub() *HDWallet {
	if bytes.Compare(w.version, version_mainnet_pub) == 0 {
		return &HDWallet{w.version, w.depth, w.fingerprint, w.childnumber, w.chaincode, w.key}
	} else {
		return &HDWallet{version_mainnet_pub, w.depth, w.fingerprint, w.childnumber, w.chaincode, privToPub(w.key)}
	}
}

// StringChild returns the ith base58-encoded extended key of a base58-encoded extended key.
func StringChild(data string, i uint32) (string, error) {
	w, err := StringWallet(data)
	if err != nil {
		return "", err
	} else {
		w, err = w.Child(i)
		if err != nil {
			return "", err
		} else {
			return w.String(), nil
		}
	}
}

//StringToAddress returns the Bitcoin address of a base58-encoded extended key.
func StringAddress(data string) (string, error) {
	w, err := StringWallet(data)
	if err != nil {
		return "", err
	} else {
		return w.Address(), nil
	}
}

// Address returns bitcoin address represented by wallet w.
func (w *HDWallet) Address() string {
	x, y := expand(w.key)
	four, _ := hex.DecodeString("04")
	padded_key := append(four, append(x.Bytes(), y.Bytes()...)...)
	var prefix []byte
	if bytes.Compare(w.version, version_testnet_pub) == 0 || bytes.Compare(w.version, version_testnet_pri) == 0 {
		prefix, _ = hex.DecodeString("6F")
	} else {
		prefix, _ = hex.DecodeString("00")
	}
	addr_1 := append(prefix, hash160(padded_key)...)
	chksum := dblSha256(addr_1)
	return base58.Encode(append(addr_1, chksum[:4]...))
}

// GenSeed returns a random seed with a length measured in bytes.
// The length must be at least 128.
func GenSeed(length int) ([]byte, error) {
	b := make([]byte, length)
	if length < 128 {
		return b, errors.New("length must be at least 128 bits")
	}
	_, err := rand.Read(b)
	return b, err
}

// MasterKey returns a new wallet given a random seed.
func MasterKey(seed []byte) *HDWallet {
	key := []byte("Bitcoin seed")
	mac := hmac.New(sha512.New, key)
	mac.Write(seed)
	I := mac.Sum(nil)
	secret := I[:len(I)/2]
	chain_code := I[len(I)/2:]
	depth := 0
	i := make([]byte, 4)
	fingerprint := make([]byte, 4)
	zero := make([]byte, 1)
	return &HDWallet{version_mainnet_pri, uint16(depth), fingerprint, i, chain_code, append(zero, secret...)}
}

// StringCheck is a validation check of a base58-encoded extended key.
func StringCheck(key string) error {
	return ByteCheck(base58.Decode(key))
}

func ByteCheck(dbin []byte) error {
	// check proper length
	if len(dbin) != 82 {
		return errors.New("invalid string")
	}
	// check for correct Public or Private vbytes
	if bytes.Compare(dbin[:4], version_mainnet_pub) != 0 && bytes.Compare(dbin[:4], version_mainnet_pri) != 0 && bytes.Compare(dbin[:4], version_testnet_pub) != 0 && bytes.Compare(dbin[:4], version_testnet_pri) != 0 {
		return errors.New("invalid string")
	}
	// if Public, check x coord is on curve
	x, y := expand(dbin[45:78])
	if bytes.Compare(dbin[:4], version_mainnet_pub) == 0 || bytes.Compare(dbin[:4], version_testnet_pub) == 0 {
		if !onCurve(x, y) {
			return errors.New("invalid string")
		}
	}
	return nil
}
