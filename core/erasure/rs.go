/*
	Copyright (C) CESS. All rights reserved.
	Copyright (C) Cumulus Encrypted Storage System. All rights reserved.

	SPDX-License-Identifier: Apache-2.0
*/

package erasure

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/CESSProject/cess-go-sdk/core/pattern"
	"github.com/CESSProject/cess-go-sdk/core/utils"
	"github.com/klauspost/reedsolomon"
)

// ReedSolomon uses reed-solomon algorithm to redundancy files
func ReedSolomon(path string) ([]string, error) {
	var shardspath = make([]string, 0)
	fstat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fstat.IsDir() {
		return nil, errors.New("not a file")
	}
	if fstat.Size() != pattern.SegmentSize {
		return nil, errors.New("invalid size")
	}

	datashards, parshards := pattern.DataShards, pattern.ParShards
	basedir := filepath.Dir(path)

	enc, err := reedsolomon.New(datashards, parshards)
	if err != nil {
		return shardspath, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return shardspath, err
	}

	// Split the file into equally sized shards.
	shards, err := enc.Split(b)
	if err != nil {
		return shardspath, err
	}
	// Encode parity
	err = enc.Encode(shards)
	if err != nil {
		return shardspath, err
	}
	// Write out the resulting files.
	for _, shard := range shards {
		hash, err := utils.CalcSHA256(shard)
		if err != nil {
			return shardspath, err
		}
		newpath := filepath.Join(basedir, hash)
		_, err = os.Stat(newpath)
		if err != nil {
			err = ioutil.WriteFile(newpath, shard, os.ModePerm)
			if err != nil {
				return shardspath, err
			}
		}
		shardspath = append(shardspath, newpath)
	}
	return shardspath, nil
}

func ReedSolomonRestore(outpath string, shardspath []string) error {
	_, err := os.Stat(outpath)
	if err == nil {
		return nil
	}

	datashards, parshards := pattern.DataShards, pattern.ParShards

	enc, err := reedsolomon.New(datashards, parshards)
	if err != nil {
		return err
	}
	shards := make([][]byte, datashards+parshards)
	for k, v := range shardspath {
		//infn := fmt.Sprintf("%s.00%d", outfn, i)
		shards[k], err = ioutil.ReadFile(v)
		if err != nil {
			shards[k] = nil
		}
	}

	// Verify the shards
	ok, _ := enc.Verify(shards)
	if !ok {
		err = enc.Reconstruct(shards)
		if err != nil {
			return err
		}
		ok, err = enc.Verify(shards)
		if !ok {
			return err
		}
	}
	f, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = enc.Join(f, shards, len(shards[0])*datashards)
	return err

}
