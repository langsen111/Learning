// Copyright 2017 The go-ethereum Authors
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
//（总结： 账户在钱包里，钱包在backend里，backend在manager里面操作处理）
package accounts

import (
	"reflect"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/event"
)

// Manager is an overarching account manager that can communicate with various
// backends for signing transactions.
//通过backend管理钱包的管理工具/账户在钱包里
type Manager struct {
	//所有已经注册的backend
	backends map[reflect.Type][]Backend // Index of backends currently registered
	//所有backend的更新订阅器
	updaters []event.Subscription       // Wallet update subscriptions for all backends
	//backend更新的订阅槽
	updates  chan WalletEvent           // Subscription sink for backend wallet changes
	//所有已经注册的backends的钱包的缓存
	wallets  []Wallet                   // Cache of all wallets from all registered backends

	//钱包到达和离开的通知
	feed event.Feed // Wallet feed notifying of arrivals/departures

	//退出队列
	quit chan chan error
	lock sync.RWMutex
}

// NewManager creates a generic account manager to sign transaction via various
// supported backends.
func NewManager(backends ...Backend) *Manager {
	// Retrieve the initial list of wallets from the backends and sort by URL
	var wallets []Wallet
	for _, backend := range backends {
		wallets = merge(wallets, backend.Wallets()...)
	}
	// Subscribe to wallet notifications from all backends
	updates := make(chan WalletEvent, 4*len(backends))

	subs := make([]event.Subscription, len(backends))
	for i, backend := range backends {
		subs[i] = backend.Subscribe(updates)
	}
	// Assemble the account manager and return
	am := &Manager{
		backends: make(map[reflect.Type][]Backend),
		updaters: subs,
		updates:  updates,
		wallets:  wallets,
		quit:     make(chan chan error),
	}
	for _, backend := range backends {
		kind := reflect.TypeOf(backend)
		am.backends[kind] = append(am.backends[kind], backend)
	}
	go am.update()

	return am
}

// Close terminates the account manager's internal notification processes.
func (am *Manager) Close() error {
	errc := make(chan error)
	am.quit <- errc
	return <-errc
}

// update is the wallet event loop listening for notifications from the backends
// and updating the cache of wallets.
//goroutine线程
func (am *Manager) update() {
	// Close all subscriptions when the manager terminates
	defer func() {
		am.lock.Lock()
		for _, sub := range am.updaters {
			sub.Unsubscribe()
		}
		am.updaters = nil
		am.lock.Unlock()
	}()

	// Loop until termination
	for {
		select {
		case event := <-am.updates:
			// Wallet event arrived, update local cache
			am.lock.Lock()
			switch event.Kind {
			case WalletArrived:
				am.wallets = merge(am.wallets, event.Wallet)
			case WalletDropped:
				am.wallets = drop(am.wallets, event.Wallet)
			}
			am.lock.Unlock()

			// Notify any listeners of the event
			am.feed.Send(event)

		case errc := <-am.quit:
			// Manager terminating, return
			errc <- nil
			return
		}
	}
}

// Backends retrieves the backend(s) with the given type from the account manager.
func (am *Manager) Backends(kind reflect.Type) []Backend {
	return am.backends[kind]
}

// Wallets returns all signer accounts registered under this account manager.
func (am *Manager) Wallets() []Wallet {
	am.lock.RLock()
	defer am.lock.RUnlock()

	cpy := make([]Wallet, len(am.wallets))
	copy(cpy, am.wallets)
	return cpy
}

// Wallet retrieves the wallet associated with a particular URL.
func (am *Manager) Wallet(url string) (Wallet, error) {
	am.lock.RLock()
	defer am.lock.RUnlock()

	parsed, err := parseURL(url)
	if err != nil {
		return nil, err
	}
	for _, wallet := range am.Wallets() {
		if wallet.URL() == parsed {
			return wallet, nil
		}
	}
	return nil, ErrUnknownWallet
}

// Find attempts to locate the wallet corresponding to a specific account. Since
// accounts can be dynamically added to and removed from wallets, this method has
// a linear runtime in the number of wallets.
//查找是否有账号了
func (am *Manager) Find(account Account) (Wallet, error) {
	am.lock.RLock()
	defer am.lock.RUnlock()

	for _, wallet := range am.wallets {
		if wallet.Contains(account) {
			return wallet, nil
		}
	}
	return nil, ErrUnknownAccount
}

//订阅消息
// Subscribe creates an async subscription to receive notifications when the
// manager detects the arrival or departure of a wallet from any of its backends.
func (am *Manager) Subscribe(sink chan<- WalletEvent) event.Subscription {
	return am.feed.Subscribe(sink)
}

// merge is a sorted analogue of append for wallets, where the ordering of the
// origin list is preserved by inserting new wallets at the correct position.
//
// The original slice is assumed to be already sorted by URL.    融合钱包
func merge(slice []Wallet, wallets ...Wallet) []Wallet {
	for _, wallet := range wallets {
		n := sort.Search(len(slice), func(i int) bool { return slice[i].URL().Cmp(wallet.URL()) >= 0 })
		if n == len(slice) {
			slice = append(slice, wallet)
			continue
		}
		slice = append(slice[:n], append([]Wallet{wallet}, slice[n:]...)...)
	}
	return slice
}

// drop is the couterpart of merge, which looks up wallets from within the sorted
// cache and removes the ones specified.
//删除钱包
func drop(slice []Wallet, wallets ...Wallet) []Wallet {
	for _, wallet := range wallets {
		n := sort.Search(len(slice), func(i int) bool { return slice[i].URL().Cmp(wallet.URL()) >= 0 })
		if n == len(slice) {
			// Wallet not found, may happen during startup
			continue
		}
		slice = append(slice[:n], slice[n+1:]...)
	}
	return slice
}
