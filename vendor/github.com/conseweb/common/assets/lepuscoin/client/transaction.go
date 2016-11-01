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
package client

import (
	"errors"
	"time"

	pb "github.com/conseweb/common/assets/lepuscoin/protos"
)

// NewTransaction
// if founder is empty, tx is a coinbase transaction
func NewTransactionV1(founder string) *pb.TX {
	tx := new(pb.TX)
	tx.Version = 1
	tx.Timestamp = time.Now().UTC().Unix()
	tx.Founder = founder
	tx.Txin = make([]*pb.TX_TXIN, 0)
	tx.Txout = make([]*pb.TX_TXOUT, 0)

	return tx
}

// NewTxIn returns a new transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(owner, prevHash string, prevIdx uint32) *pb.TX_TXIN {
	return &pb.TX_TXIN{
		SourceHash: prevHash,
		Ix:         prevIdx,
		Addr:       owner,
	}
}

// NewTxOut returns a new transaction output with the provided
// transaction value and public key script.
func NewTxOut(value uint64, addr string, untilTime time.Time) *pb.TX_TXOUT {
	var until int64 = 0
	if untilTime.UTC().After(time.Now().UTC()) {
		until = untilTime.UTC().Unix()
	}

	return &pb.TX_TXOUT{
		Value: value,
		Addr:  addr,
		Until: until,
	}
}

// VerifyTx verify tx is valid or not
// If not, error returned
func VerifyTx(tx *pb.TX) error {
	// time check
	if time.Now().UTC().Before(time.Unix(tx.Timestamp, 0).UTC()) {
		return errors.New("tx occur time after now, invalid")
	}

	if tx.Founder == "" {
		return errors.New("no founder transaction")
	}

	if tx.Txout == nil || len(tx.Txout) == 0 {
		return errors.New("transaction output is empty")
	}

	return nil
}
