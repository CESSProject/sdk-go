package chain

import (
	"log"
	"time"

	"github.com/CESSProject/sdk-go/core/event"
	"github.com/CESSProject/sdk-go/core/pattern"
	"github.com/CESSProject/sdk-go/core/rule"
	"github.com/CESSProject/sdk-go/core/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/pkg/errors"
)

func (c *ChainSDK) QueryBucketInfo(puk []byte, bucketname string) (pattern.BucketInfo, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var data pattern.BucketInfo

	if !c.GetChainState() {
		return data, pattern.ERR_RPC_CONNECTION
	}

	acc, err := types.NewAccountID(puk)
	if err != nil {
		return data, errors.Wrap(err, "[NewAccountID]")
	}

	owner, err := codec.Encode(*acc)
	if err != nil {
		return data, errors.Wrap(err, "[EncodeToBytes]")
	}

	name, err := codec.Encode(bucketname)
	if err != nil {
		return data, errors.Wrap(err, "[Encode]")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.FILEBANK, pattern.BUCKET, owner, name)
	if err != nil {
		return data, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &data)
	if err != nil {
		return data, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return data, pattern.ERR_RPC_EMPTY_VALUE
	}
	return data, nil
}

func (c *ChainSDK) QueryBucketList(puk []byte) ([]types.Bytes, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var data []types.Bytes

	if !c.GetChainState() {
		return data, pattern.ERR_RPC_CONNECTION
	}

	acc, err := types.NewAccountID(puk)
	if err != nil {
		return data, errors.Wrap(err, "[NewAccountID]")
	}

	owner, err := codec.Encode(*acc)
	if err != nil {
		return data, errors.Wrap(err, "[EncodeToBytes]")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.FILEBANK, pattern.BUCKETLIST, owner)
	if err != nil {
		return data, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &data)
	if err != nil {
		return data, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return data, pattern.ERR_RPC_EMPTY_VALUE
	}
	return data, nil
}

func (c *ChainSDK) QueryAllBucketName(owner []byte) ([]string, error) {
	bucketlist, err := c.QueryBucketList(owner)
	if err != nil {
		if err.Error() != pattern.ERR_Empty {
			return nil, err
		}
	}
	var buckets = make([]string, len(bucketlist))
	for i := 0; i < len(bucketlist); i++ {
		buckets[i] = string(bucketlist[i])
	}
	return buckets, nil
}

// QueryFileMetaData
func (c *ChainSDK) QueryFileMetadata(roothash string) (pattern.FileMetadata, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		data pattern.FileMetadata
		hash pattern.FileHash
	)

	if !c.GetChainState() {
		return data, pattern.ERR_RPC_CONNECTION
	}

	if len(hash) != len(roothash) {
		return data, errors.New("invalid filehash")
	}

	for i := 0; i < len(hash); i++ {
		hash[i] = types.U8(roothash[i])
	}

	b, err := codec.Encode(hash)
	if err != nil {
		return data, errors.Wrap(err, "[Encode]")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.FILEBANK, pattern.FILE, b)
	if err != nil {
		return data, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &data)
	if err != nil {
		return data, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return data, pattern.ERR_RPC_EMPTY_VALUE
	}
	return data, nil
}

func (c *ChainSDK) QueryStorageOrder(roothash string) (pattern.StorageOrder, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		data pattern.StorageOrder
		hash pattern.FileHash
	)

	if len(hash) != len(roothash) {
		return data, errors.New("invalid filehash")
	}

	for i := 0; i < len(hash); i++ {
		hash[i] = types.U8(roothash[i])
	}

	b, err := codec.Encode(hash)
	if err != nil {
		return data, errors.Wrap(err, "[Encode]")
	}

	if !c.GetChainState() {
		return data, pattern.ERR_RPC_CONNECTION
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.FILEBANK, pattern.DEALMAP, b)
	if err != nil {
		return data, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &data)
	if err != nil {
		return data, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return data, pattern.ERR_RPC_EMPTY_VALUE
	}
	return data, nil
}

func (c *ChainSDK) QueryPendingReplacements(puk []byte) (uint32, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var data types.U32

	acc, err := types.NewAccountID(puk)
	if err != nil {
		return 0, errors.Wrap(err, "[NewAccountID]")
	}

	owner, err := codec.Encode(*acc)
	if err != nil {
		return 0, errors.Wrap(err, "[EncodeToBytes]")
	}

	if !c.GetChainState() {
		return 0, pattern.ERR_RPC_CONNECTION
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.FILEBANK, pattern.PENDINGREPLACE, owner)
	if err != nil {
		return 0, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &data)
	if err != nil {
		return 0, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return 0, pattern.ERR_RPC_EMPTY_VALUE
	}
	return uint32(data), nil
}

func (c *ChainSDK) SubmitIdleMetadata(teeAcc []byte, idlefiles []pattern.IdleMetadata) (string, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, pattern.ERR_RPC_CONNECTION
	}

	acc, err := types.NewAccountID(teeAcc)
	if err != nil {
		return txhash, errors.Wrap(err, "[NewAccountID]")
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_UPLOADFILLER, *acc, idlefiles)
	if err != nil {
		return txhash, errors.Wrap(err, "[NewCall]")
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
		return txhash, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
	}
	defer sub.Unsubscribe()

	timeout := time.NewTimer(c.timeForBlockOut)
	defer timeout.Stop()

	for {
		select {
		case status := <-sub.Chan():
			if status.IsInBlock {
				//events := EventRecords{}
				txhash, _ = codec.EncodeToHex(status.AsInBlock)
				// h, err := c.api.RPC.State.GetStorageRaw(c.keyEvents, status.AsInBlock)
				// if err != nil {
				// 	return txhash, errors.Wrap(err, "[GetStorageRaw]")
				// }

				// types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)

				// if len(events.FileBank_DeleteFile) > 0 {
				// 	return txhash, events.FileBank_DeleteFile[0].FailedList
				// }
				return txhash, nil
			}
		case err = <-sub.Err():
			return txhash, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, pattern.ERR_RPC_TIMEOUT
		}
	}
}

func (c *ChainSDK) SubmitIdleFile(teeAcc []byte, idlefiles []pattern.IdleFileMeta) (string, error) {
	var submit = make([]pattern.IdleMetadata, 0)
	for i := 0; i < len(idlefiles); i++ {
		var filehash pattern.FileHash
		acc, err := types.NewAccountID(idlefiles[i].MinerAcc)
		if err != nil {
			continue
		}

		if len(idlefiles[i].Hash) != len(pattern.FileHash{}) {
			continue
		}

		for j := 0; j < len(idlefiles[i].Hash); j++ {
			filehash[j] = types.U8(idlefiles[i].Hash[j])
		}

		var ele = pattern.IdleMetadata{
			Size:      types.NewU64(idlefiles[i].Size),
			BlockNum:  types.NewU32(idlefiles[i].BlockNum),
			BlockSize: types.NewU32(idlefiles[i].BlockSize),
			ScanSize:  types.NewU32(idlefiles[i].ScanSize),
			Acc:       *acc,
			Hash:      filehash,
		}
		submit = append(submit, ele)
		if len(submit) >= rule.MaxSubmitedIdleFileMeta {
			break
		}
	}
	return c.SubmitIdleMetadata(teeAcc, submit)
}

func (c *ChainSDK) CreateBucket(owner_pkey []byte, name string) (string, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, pattern.ERR_RPC_CONNECTION
	}

	acc, err := types.NewAccountID(owner_pkey)
	if err != nil {
		return txhash, errors.Wrap(err, "[NewAccountID]")
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_PUTBUCKET, *acc, types.NewBytes([]byte(name)))
	if err != nil {
		return txhash, errors.Wrap(err, "[NewCall]")
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
				if err != nil || len(events.FileBank_CreateBucket) > 0 {
					return txhash, nil
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

func (c *ChainSDK) DeleteBucket(owner_pkey []byte, name string) (string, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, pattern.ERR_RPC_CONNECTION
	}

	acc, err := types.NewAccountID(owner_pkey)
	if err != nil {
		return txhash, errors.Wrap(err, "[NewAccountID]")
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_DELBUCKET, *acc, types.NewBytes([]byte(name)))
	if err != nil {
		return txhash, errors.Wrap(err, "[NewCall]")
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
				if err != nil || len(events.FileBank_DeleteBucket) > 0 {
					return txhash, nil
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

func (c *ChainSDK) UploadDeclaration(filehash string, dealinfo []pattern.SegmentList, user pattern.UserBrief) (string, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		hash        pattern.FileHash
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, pattern.ERR_RPC_CONNECTION
	}

	if len(filehash) != len(hash) {
		return txhash, errors.New("invalid filehash")
	}
	for i := 0; i < len(hash); i++ {
		hash[i] = types.U8(filehash[i])
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_UPLOADDEC, hash, dealinfo, user)
	if err != nil {
		return txhash, errors.Wrap(err, "[NewCall]")
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
				if err != nil || len(events.FileBank_UploadDeclaration) > 0 {
					return txhash, nil
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

func (c *ChainSDK) DeleteFile(puk []byte, filehash []string) (string, []pattern.FileHash, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		accountInfo types.AccountInfo
		hashs       = make([]pattern.FileHash, len(filehash))
	)

	if !c.GetChainState() {
		return txhash, hashs, pattern.ERR_RPC_CONNECTION
	}

	for j := 0; j < len(filehash); j++ {
		if len(filehash[j]) != len(hashs[j]) {
			return txhash, hashs, errors.New("invalid filehash")
		}
		for i := 0; i < len(hashs[j]); i++ {
			hashs[j][i] = types.U8(filehash[j][i])
		}
	}

	acc, err := types.NewAccountID(puk)
	if err != nil {
		return txhash, hashs, errors.Wrap(err, "[NewAccountID]")
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_DELFILE, *acc, hashs)
	if err != nil {
		return txhash, hashs, errors.Wrap(err, "[NewCall]")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.SYSTEM, pattern.ACCOUNT, c.keyring.PublicKey)
	if err != nil {
		return txhash, hashs, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return txhash, hashs, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return txhash, hashs, pattern.ERR_RPC_EMPTY_VALUE
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
		return txhash, hashs, errors.Wrap(err, "[Sign]")
	}

	// Do the transfer and track the actual status
	sub, err := c.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return txhash, hashs, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
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
					return txhash, hashs, errors.Wrap(err, "[GetStorageRaw]")
				}
				err = types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)
				if err != nil || len(events.FileBank_DeleteFile) > 0 {
					return txhash, events.FileBank_DeleteFile[0].Filehash, nil
				}
				return txhash, hashs, errors.New(pattern.ERR_Failed)
			}
		case err = <-sub.Err():
			return txhash, hashs, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, hashs, pattern.ERR_RPC_TIMEOUT
		}
	}
}

func (c *ChainSDK) SubmitFileReport(roothash []pattern.FileHash) (string, []pattern.FileHash, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, nil, pattern.ERR_RPC_CONNECTION
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_FILEREPORT, roothash)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[NewCall]")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.SYSTEM, pattern.ACCOUNT, c.keyring.PublicKey)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return txhash, nil, pattern.ERR_RPC_EMPTY_VALUE
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
		return txhash, nil, errors.Wrap(err, "[Sign]")
	}

	// Do the transfer and track the actual status
	sub, err := c.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
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
					return txhash, nil, errors.Wrap(err, "[GetStorageRaw]")
				}
				err = types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)
				if err != nil || len(events.FileBank_TransferReport) > 0 {
					return txhash, events.FileBank_TransferReport[0].Failed_list, nil
				}
				return txhash, nil, errors.New(pattern.ERR_Failed)
			}
		case err = <-sub.Err():
			return txhash, nil, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, nil, pattern.ERR_RPC_TIMEOUT
		}
	}
}

func (c *ChainSDK) ReportFiles(roothash []string) (string, []string, error) {
	var hashs = make([]pattern.FileHash, len(roothash))
	for i := 0; i < len(roothash); i++ {
		for j := 0; j < len(roothash[i]); j++ {
			hashs[i][j] = types.U8(roothash[i][j])
		}
	}
	txhash, failed, err := c.SubmitFileReport(hashs)
	var failedfiles = make([]string, len(failed))
	for k, v := range failed {
		failedfiles[k] = string(v[:])
	}
	return txhash, failedfiles, err
}

func (c *ChainSDK) ReplaceIdleFiles(roothash []pattern.FileHash) (string, []pattern.FileHash, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if err := recover(); err != nil {
			log.Println(utils.RecoverError(err))
		}
	}()

	var (
		txhash      string
		accountInfo types.AccountInfo
	)

	if !c.GetChainState() {
		return txhash, nil, pattern.ERR_RPC_CONNECTION
	}

	call, err := types.NewCall(c.metadata, pattern.TX_FILEBANK_REPLACEFILE, roothash)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[NewCall]")
	}

	key, err := types.CreateStorageKey(c.metadata, pattern.SYSTEM, pattern.ACCOUNT, c.keyring.PublicKey)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[CreateStorageKey]")
	}

	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[GetStorageLatest]")
	}
	if !ok {
		return txhash, nil, pattern.ERR_RPC_EMPTY_VALUE
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
		return txhash, nil, errors.Wrap(err, "[Sign]")
	}

	// Do the transfer and track the actual status
	sub, err := c.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return txhash, nil, errors.Wrap(err, "[SubmitAndWatchExtrinsic]")
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
					return txhash, nil, errors.Wrap(err, "[GetStorageRaw]")
				}
				err = types.EventRecordsRaw(*h).DecodeEventRecords(c.metadata, &events)
				if err != nil || len(events.FileBank_ReplaceFiller) > 0 {
					return txhash, events.FileBank_ReplaceFiller[0].Filler_list, nil
				}
				return txhash, nil, errors.New(pattern.ERR_Failed)
			}
		case err = <-sub.Err():
			return txhash, nil, errors.Wrap(err, "[sub]")
		case <-timeout.C:
			return txhash, nil, pattern.ERR_RPC_TIMEOUT
		}
	}
}

func (c *ChainSDK) ReplaceFile(roothash []string) (string, []string, error) {
	var hashs = make([]pattern.FileHash, len(roothash))
	for i := 0; i < len(roothash); i++ {
		for j := 0; j < len(roothash[i]); j++ {
			hashs[i][j] = types.U8(roothash[i][j])
		}
	}
	txhash, failed, err := c.ReplaceIdleFiles(hashs)
	var failedFiles = make([]string, len(failed))
	for k, v := range failed {
		failedFiles[k] = string(v[:])
	}
	return txhash, failedFiles, err
}