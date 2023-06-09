/*
	Copyright (C) CESS. All rights reserved.
	Copyright (C) Cumulus Encrypted Storage System. All rights reserved.

	SPDX-License-Identifier: Apache-2.0
*/

package chain

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/CESSProject/cess-go-sdk/core/event"
	"github.com/CESSProject/cess-go-sdk/core/pattern"
	"github.com/CESSProject/cess-go-sdk/core/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/pkg/errors"
)

func (c *ChainSDK) Register(role string, puk []byte, earnings string, pledge uint64) (string, string, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		err         error
		txhash      string
		pubkey      []byte
		minerinfo   pattern.MinerInfo
		acc         *types.AccountID
		call        types.Call
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, earnings, pattern.ERR_RPC_CONNECTION
	}

	var peerid pattern.PeerId
	if len(peerid) != len(puk) {
		return txhash, earnings, fmt.Errorf("invalid peerid: %v", puk)
	}
	for i := 0; i < len(peerid); i++ {
		peerid[i] = types.U8(puk[i])
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.SYSTEM, pattern.ACCOUNT, c.keyring.PublicKey)
	if err != nil {
		return txhash, earnings, errors.Wrap(err, "[CreateStorageKey]")
	}

	switch role {
	case pattern.Role_OSS, pattern.Role_DEOSS, "deoss", "oss", "Deoss", "DeOSS":
		id, err := c.QueryDeossPeerPublickey(c.keyring.PublicKey)
		if err != nil {
			if err.Error() != pattern.ERR_Empty {
				return txhash, earnings, err
			}
		} else {
			if !utils.CompareSlice(id, puk) {
				txhash, err = c.updateAddress(key, role, peerid)
				return txhash, earnings, err
			}
			return "", earnings, nil
		}
		call, err = types.NewCall(c.metadata, pattern.TX_OSS_REGISTER, peerid)
		if err != nil {
			return txhash, earnings, errors.Wrap(err, "[NewCall]")
		}
	case pattern.Role_BUCKET, "SMINER", "bucket", "Bucket", "Sminer", "sminer":
		minerinfo, err = c.QueryStorageMiner(c.keyring.PublicKey)
		if err != nil {
			if err.Error() != pattern.ERR_Empty {
				return txhash, earnings, err
			}
		} else {
			if !utils.CompareSlice([]byte(string(minerinfo.PeerId[:])), puk) {
				txhash, err = c.updateAddress(key, role, peerid)
				return txhash, earnings, err
			}
			acc, _ := utils.EncodePublicKeyAsCessAccount(minerinfo.BeneficiaryAcc[:])
			if earnings != "" {
				if acc != earnings {
					puk, err := utils.ParsingPublickey(earnings)
					if err != nil {
						return txhash, acc, err
					}
					txhash, err = c.updateEarningsAcc(key, puk)
					return txhash, earnings, err
				}
			}
			return "", acc, nil
		}

		pubkey, err = utils.ParsingPublickey(earnings)
		if err != nil {
			return txhash, earnings, errors.Wrap(err, "[DecodeToPub]")
		}
		acc, err = types.NewAccountID(pubkey)
		if err != nil {
			return txhash, earnings, errors.Wrap(err, "[NewAccountID]")
		}
		realTokens, ok := new(big.Int).SetString(strconv.FormatUint(pledge, 10)+pattern.TokenPrecision_CESS, 10)
		if !ok {
			return txhash, earnings, errors.New("[big.Int.SetString]")
		}
		call, err = types.NewCall(c.metadata, pattern.TX_SMINER_REGISTER, *acc, peerid, types.NewU128(*realTokens))
		if err != nil {
			return txhash, earnings, errors.Wrap(err, "[NewCall]")
		}
	default:
		return "", earnings, fmt.Errorf("invalid role name")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return txhash, earnings, errors.Wrap(err, "[GetStorageLatest]")
	}

	if !ok {
		keyStr, _ := utils.NumsToByteStr(key, map[string]bool{})
		return txhash, earnings, fmt.Errorf(
			"chain rpc.state.GetStorageLatest[%v]: %v",
			keyStr,
			pattern.ERR_RPC_EMPTY_VALUE,
		)
	}

	o := types.SignatureOptions{
		BlockHash:          c.genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        c.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        c.runtimeVersion.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: c.runtimeVersion.TransactionVersion,
	}

	ext := types.NewExtrinsic(call)

	// Sign the transaction
	err = ext.Sign(c.keyring, o)
	if err != nil {
		return txhash, earnings, errors.Wrap(err, "[Sign]")
	}

	// Do the transfer and track the actual status
	sub, err := c.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		c.SetChainState(false)
		return txhash, earnings, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
	}
	defer sub.Unsubscribe()

	timeout := time.NewTimer(c.timeForBlockOut)
	defer timeout.Stop()

	for {
		select {
		case status := <-sub.Chan():
			if status.IsInBlock {
				events := event.EventRecords{}
				txhash, _ = codec.EncodeToHex(status.AsInBlock)
				h, err := c.api.RPC.State.GetStorageRaw(c.keyEvents, status.AsInBlock)
				if err != nil {
					return txhash, earnings, errors.Wrap(err, "[GetStorageRaw]")
				}
				err = types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)
				if err != nil {
					return txhash, earnings, nil
				}
				switch role {
				case pattern.Role_OSS, pattern.Role_DEOSS, "deoss", "oss", "Deoss", "DeOSS":
					if len(events.Oss_OssRegister) > 0 {
						return txhash, earnings, nil
					}
				case pattern.Role_BUCKET, "SMINER", "bucket", "Bucket", "Sminer", "sminer":
					if len(events.Sminer_Registered) > 0 {
						return txhash, earnings, nil
					}
				default:
					return txhash, earnings, errors.New(pattern.ERR_Failed)
				}
			}
		case err = <-sub.Err():
			return txhash, earnings, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, earnings, pattern.ERR_RPC_TIMEOUT
		}
	}
}

func (c *ChainSDK) updateAddress(key types.StorageKey, name string, peerid pattern.PeerId) (string, error) {
	var (
		err         error
		txhash      string
		call        types.Call
		accountInfo types.AccountInfo
	)

	switch name {
	case pattern.Role_OSS, pattern.Role_DEOSS, "deoss", "oss", "Deoss", "DeOSS":

		call, err = types.NewCall(c.metadata, pattern.TX_OSS_UPDATE, peerid)
		if err != nil {
			return txhash, errors.Wrap(err, "[NewCall]")
		}
	case pattern.Role_BUCKET, "SMINER", "bucket", "Bucket", "Sminer", "sminer":
		call, err = types.NewCall(c.metadata, pattern.TX_SMINER_UPDATEPEERID, peerid)
		if err != nil {
			return txhash, errors.Wrap(err, "[NewCall]")
		}
	default:
		return "", fmt.Errorf("Invalid role name")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return txhash, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return txhash, pattern.ERR_RPC_EMPTY_VALUE
	}

	o := types.SignatureOptions{
		BlockHash:          c.genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        c.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        c.runtimeVersion.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: c.runtimeVersion.TransactionVersion,
	}

	ext := types.NewExtrinsic(call)

	// Sign the transaction
	err = ext.Sign(c.keyring, o)
	if err != nil {
		return txhash, errors.Wrap(err, "[Sign]")
	}

	// Do the transfer and track the actual status
	sub, err := c.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		c.SetChainState(false)
		return txhash, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
	}
	defer sub.Unsubscribe()
	timeout := time.NewTimer(c.timeForBlockOut)
	defer timeout.Stop()
	for {
		select {
		case status := <-sub.Chan():
			if status.IsInBlock {
				events := event.EventRecords{}
				txhash, _ = codec.EncodeToHex(status.AsInBlock)
				h, err := c.api.RPC.State.GetStorageRaw(c.keyEvents, status.AsInBlock)
				if err != nil {
					return txhash, errors.Wrap(err, "[GetStorageRaw]")
				}
				err = types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)
				if err != nil {
					return txhash, nil
				}
				switch name {
				case pattern.Role_OSS, pattern.Role_DEOSS, "deoss", "oss", "Deoss", "DeOSS":
					if len(events.Oss_OssUpdate) > 0 {
						return txhash, nil
					}
				case pattern.Role_BUCKET, "SMINER", "bucket", "Bucket", "Sminer", "sminer":
					if len(events.Sminer_UpdataIp) > 0 {
						return txhash, nil
					}
				default:
					return txhash, errors.New(pattern.ERR_Failed)
				}
				return txhash, errors.New(pattern.ERR_Failed)
			}
		case err = <-sub.Err():
			return txhash, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, pattern.ERR_RPC_TIMEOUT
		}
	}
}

func (c *ChainSDK) Exit(role string) (string, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		err         error
		txhash      string
		call        types.Call
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, pattern.ERR_RPC_CONNECTION
	}

	switch role {
	case pattern.Role_OSS, pattern.Role_DEOSS, "deoss", "oss", "Deoss", "DeOSS":
		call, err = types.NewCall(c.metadata, pattern.TX_OSS_DESTORY)
		if err != nil {
			return txhash, errors.Wrap(err, "[NewCall]")
		}
	case pattern.Role_BUCKET, "SMINER", "bucket", "Bucket", "Sminer", "sminer":
		call, err = types.NewCall(c.metadata, pattern.TX_FILEBANK_MINEREXITPREP)
		if err != nil {
			return txhash, errors.Wrap(err, "[NewCall]")
		}
	default:
		return "", fmt.Errorf("Invalid role name")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.SYSTEM, pattern.ACCOUNT, c.keyring.PublicKey)
	if err != nil {
		return txhash, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return txhash, errors.Wrap(err, "[GetStorageLatest]")
	}

	if !ok {
		return txhash, pattern.ERR_RPC_EMPTY_VALUE
	}

	o := types.SignatureOptions{
		BlockHash:          c.genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        c.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        c.runtimeVersion.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: c.runtimeVersion.TransactionVersion,
	}

	ext := types.NewExtrinsic(call)

	// Sign the transaction
	err = ext.Sign(c.keyring, o)
	if err != nil {
		return txhash, errors.Wrap(err, "[Sign]")
	}

	// Do the transfer and track the actual status
	sub, err := c.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		c.SetChainState(false)
		return txhash, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
	}
	defer sub.Unsubscribe()

	timeout := time.NewTimer(c.timeForBlockOut)
	defer timeout.Stop()

	for {
		select {
		case status := <-sub.Chan():
			if status.IsInBlock {
				events := event.EventRecords{}
				txhash, _ = codec.EncodeToHex(status.AsInBlock)
				h, err := c.api.RPC.State.GetStorageRaw(c.keyEvents, status.AsInBlock)
				if err != nil {
					return txhash, errors.Wrap(err, "[GetStorageRaw]")
				}
				err = types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)
				if err != nil {
					return txhash, nil
				}
				switch role {
				case pattern.Role_OSS, pattern.Role_DEOSS, "deoss", "oss", "Deoss", "DeOSS":
					if len(events.Oss_OssDestroy) > 0 {
						return txhash, nil
					}
				case pattern.Role_BUCKET, "SMINER", "bucket", "Bucket", "Sminer", "sminer":
					if len(events.Sminer_MinerExitPrep) > 0 {
						return txhash, nil
					}
				default:
					return txhash, errors.New(pattern.ERR_Failed)
				}
			}
		case err = <-sub.Err():
			return txhash, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, pattern.ERR_RPC_TIMEOUT
		}
	}
}
