/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kvledger

import (
	"github.com/hyperledger/fabric/common/ledger/util/leveldbhelper"
	"github.com/hyperledger/fabric/core/ledger/kvledger/msgs"
	"github.com/pkg/errors"
)

// PauseChannel updates the channel status to inactive in ledgerProviders.
func PauseChannel(rootFSPath, ledgerID string) error {
	if err := pauseOrResumeChannel(rootFSPath, ledgerID, msgs.Status_INACTIVE); err != nil {
		return err
	}
	logger.Infof("The channel [%s] has been successfully paused", ledgerID)
	return nil
}

// ResumeChannel updates the channel status to active in ledgerProviders
func ResumeChannel(rootFSPath, ledgerID string) error {
	logger.Debugf("ResumeChannel() started... If you are using RocksDB errors migh happen") //TODO: delete this
	if err := pauseOrResumeChannel(rootFSPath, ledgerID, msgs.Status_ACTIVE); err != nil {
		return err
	}
	logger.Infof("The channel [%s] has been successfully resumed", ledgerID)
	return nil
}

func pauseOrResumeChannel(rootFSPath, ledgerID string, status msgs.Status) error {
	//TODO: we should change this part if want to use other state database (e.g. RocksDB)
	logger.Debugf("pauseOrResumeChannel() was called... if you use RocksDB errors might happen") //TODO: add logic if RocksDB is used as state db
	fileLock := leveldbhelper.NewFileLock(fileLockPath(rootFSPath))
	if err := fileLock.Lock(); err != nil {
		return errors.Wrap(err, "as another peer node command is executing,"+
			" wait for that command to complete its execution or terminate it before retrying")
	}
	defer fileLock.Unlock()

	idStore, err := openIDStore(LedgerProviderPath(rootFSPath))
	if err != nil {
		return err
	}
	defer idStore.db.Close()
	return idStore.updateLedgerStatus(ledgerID, status)
}
