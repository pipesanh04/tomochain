// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package tomox_state

import (
	"sync"

	"github.com/tomochain/go-tomochain/common"
)

type exchanges struct {
	stateObject *stateExchanges
	nstart      uint64
	nonces      []bool
}

type TomoXManagedState struct {
	*TomoXStateDB
	mu        sync.RWMutex
	exchanges map[common.Hash]*exchanges
}

// TomoXManagedState returns a new managed state with the statedb as it's backing layer
func ManageState(statedb *TomoXStateDB) *TomoXManagedState {
	return &TomoXManagedState{
		TomoXStateDB: statedb.Copy(),
		exchanges:    make(map[common.Hash]*exchanges),
	}
}

// SetState sets the backing layer of the managed state
func (ms *TomoXManagedState) SetState(statedb *TomoXStateDB) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.TomoXStateDB = statedb
}

// RemoveNonce removed the nonce from the managed state and all future pending nonces
func (ms *TomoXManagedState) RemoveNonce(addr common.Hash, n uint64) {
	if ms.hasAccount(addr) {
		ms.mu.Lock()
		defer ms.mu.Unlock()

		account := ms.getAccount(addr)
		if n-account.nstart <= uint64(len(account.nonces)) {
			reslice := make([]bool, n-account.nstart)
			copy(reslice, account.nonces[:n-account.nstart])
			account.nonces = reslice
		}
	}
}

// NewNonce returns the new canonical nonce for the managed orderId
func (ms *TomoXManagedState) NewNonce(addr common.Hash) uint64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	account := ms.getAccount(addr)
	for i, nonce := range account.nonces {
		if !nonce {
			return account.nstart + uint64(i)
		}
	}
	account.nonces = append(account.nonces, true)

	return uint64(len(account.nonces)-1) + account.nstart
}

// GetNonce returns the canonical nonce for the managed or unmanaged orderId.
//
// Because GetNonce mutates the DB, we must take a write lock.
func (ms *TomoXManagedState) GetNonce(addr common.Hash) uint64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.hasAccount(addr) {
		account := ms.getAccount(addr)
		return uint64(len(account.nonces)) + account.nstart
	} else {
		return ms.TomoXStateDB.GetNonce(addr)
	}
}

// SetNonce sets the new canonical nonce for the managed state
func (ms *TomoXManagedState) SetNonce(addr common.Hash, nonce uint64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	so := ms.GetOrNewStateExchangeObject(addr)
	so.SetNonce(nonce)

	ms.exchanges[addr] = newAccount(so)
}

// HasAccount returns whether the given address is managed or not
func (ms *TomoXManagedState) HasAccount(addr common.Hash) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.hasAccount(addr)
}

func (ms *TomoXManagedState) hasAccount(addr common.Hash) bool {
	_, ok := ms.exchanges[addr]
	return ok
}

// populate the managed state
func (ms *TomoXManagedState) getAccount(addr common.Hash) *exchanges {
	if account, ok := ms.exchanges[addr]; !ok {
		so := ms.GetOrNewStateExchangeObject(addr)
		ms.exchanges[addr] = newAccount(so)
	} else {
		// Always make sure the state orderId nonce isn't actually higher
		// than the tracked one.
		so := ms.TomoXStateDB.getStateExchangeObject(addr)
		if so != nil && uint64(len(account.nonces))+account.nstart < so.Nonce() {
			ms.exchanges[addr] = newAccount(so)
		}

	}

	return ms.exchanges[addr]
}

func newAccount(so *stateExchanges) *exchanges {
	return &exchanges{so, so.Nonce(), nil}
}
