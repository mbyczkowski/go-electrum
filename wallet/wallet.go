package wallet

import (
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"

	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
)

var (
	waddrmgrNamespaceKey = []byte("waddrmgrNamespace")
	wtxmgrNamespaceKey   = []byte("wtxmgr")

	bitcoinNetwork = &netparams.MainNetParams
)

type Wallet struct {
	wallet *wallet.Wallet
}

// Addresses returns all addresses generated in the current bitcoin wallet.
func (w *Wallet) Addresses() ([]btcutil.Address, error) {
	acc, err := w.wallet.Manager.LastAccount()
	if err != nil {
		return nil, err
	}
	var addrs []btcutil.Address
	err = w.wallet.Manager.ForEachAccountAddress(acc, func(maddr waddrmgr.ManagedAddress) error {
		addrs = append(addrs, maddr.Address())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

// GenAddresses generates a number of addresses for the wallet.
func (w *Wallet) GenAddresses(n int) ([]btcutil.Address, error) {
	acc, err := w.wallet.Manager.LastAccount()
	if err != nil {
		return nil, err
	}
	managedAddrs, err := w.wallet.Manager.NextExternalAddresses(acc, uint32(n))
	if err != nil {
		return nil, err
	}
	return stripManagedAddrs(managedAddrs), nil
}

// SendBitcoin sends some amount of bitcoin specifying minimum confirmations.
func (w *Wallet) SendBitcoin(to map[string]btcutil.Amount, minconf int) error {
	acc, err := w.wallet.Manager.LastAccount()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	tx, err := w.wallet.CreateSimpleTx(acc, to, int32(minconf))
}

func stripManagedAddrs(mAddrs []waddrmgr.ManagedAddress) []btcutil.Address {
	addrs := make([]btcutil.Address, len(mAddrs))
	for i, addr := range mAddrs {
		addrs[i] = addr.Address()
	}
	return addrs
}

// CreateWallet creates a wallet with the specified path, private key password, and seed.
// Seed can be created using: hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
func CreateWallet(path, privPass string, seed []byte) (*Wallet, error) {
	db, err := walletdb.Create("bdb", path)
	if err != nil {
		return nil, err
	}
	namespace, err := db.Namespace(waddrmgrNamespaceKey)
	if err != nil {
		return nil, err
	}
	manager, err := waddrmgr.Create(namespace, seed, nil,
		[]byte(privPass), bitcoinNetwork.Params, nil)
	if err != nil {
		return nil, err
	}
	manager.Close()

	return openWallet(db, privPass, seed)
}

func returnBytes(bytes []byte) func() ([]byte, error) {
	return func() ([]byte, error) {
		return bytes, nil
	}
}

func LoadWallet(path, privPass string, seed []byte) (*Wallet, error) {
	db, err := walletdb.Open("bdb", path)
	if err != nil {
		return nil, err
	}
	return openWallet(db, privPass, seed)
}

func openWallet(db walletdb.DB, privPass string, seed []byte) (*Wallet, error) {
	addrMgrNS, err := db.Namespace(waddrmgrNamespaceKey)
	if err != nil {
		return nil, err
	}
	txMgrNS, err := db.Namespace(wtxmgrNamespaceKey)
	if err != nil {
		return nil, err
	}

	cbs := &waddrmgr.OpenCallbacks{
		ObtainSeed:        returnBytes(seed),
		ObtainPrivatePass: returnBytes([]byte(privPass)),
	}
	w, err := wallet.Open(nil, bitcoinNetwork.Params, db, addrMgrNS, txMgrNS, cbs)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		wallet: w,
	}, nil
}