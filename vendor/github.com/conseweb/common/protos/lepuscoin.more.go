/*
Copyright Mojing Inc. 2016 All Rights Reserved.
Written by mint.zhao.chiu@gmail.com. github.com: https://www.github.com/mintzhao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package protos

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/golang/protobuf/proto"
)

func (tx *TX) Base64Bytes() ([]byte, error) {
	txBytes, err := tx.Bytes()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(txBytes)))
	base64.StdEncoding.Encode(buf, txBytes)

	return buf, nil
}

func (tx *TX) Bytes() ([]byte, error) {
	return proto.Marshal(tx)
}

// AddTxIn adds a transaction input to the message.
func (tx *TX) AddTxIn(ti *TX_TXIN) {
	tx.Txin = append(tx.Txin, ti)
}

// AddTxOut adds a transaction output to the message.
func (tx *TX) AddTxOut(to *TX_TXOUT) {
	tx.Txout = append(tx.Txout, to)
}

func (e *ExecResult) Bytes() ([]byte, error) {
	return proto.Marshal(e)
}

// TxHash generates the Hash for the transaction.
func (tx *TX) TxHash() string {
	txBytes, err := tx.Bytes()
	if err != nil {
		return ""
	}

	fHash := sha256.Sum256(txBytes)
	lHash := sha256.Sum256(fHash[:])
	return hex.EncodeToString(lHash[:])
}

func (r *QueryAddrResult) Bytes() ([]byte, error) {
	return proto.Marshal(r)
}

// ParseTXBytes unmarshal txData into TX object
func ParseTXBytes(txData []byte) (*TX, error) {
	tx := new(TX)
	err := proto.Unmarshal(txData, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// marshal lepuscoininfo
func (info *LepuscoinInfo) Bytes() ([]byte, error) {
	return proto.Marshal(info)
}

// ParseLepuscoinInfoBytes unmarshal infoBytes into LepuscoinInfo
func ParseLepuscoinInfoBytes(infoBytes []byte) (*LepuscoinInfo, error) {
	info := new(LepuscoinInfo)
	if err := proto.Unmarshal(infoBytes, info); err != nil {
		return nil, err
	}

	return info, nil
}
