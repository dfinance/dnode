package app

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"os"
	"sort"
	"testing"
	vmConfig "wings-blockchain/cmd/config"

	poaTypes "wings-blockchain/x/poa/types"
)

// Type that combines an Address with the privKey and pubKey to that address
type AddrKeys struct {
	Address sdk.AccAddress
	PubKey  crypto.PubKey
	PrivKey crypto.PrivKey
}

func NewAddrKeys(address sdk.AccAddress, pubKey crypto.PubKey,
	privKey crypto.PrivKey) AddrKeys {

	return AddrKeys{
		Address: address,
		PubKey:  pubKey,
		PrivKey: privKey,
	}
}

// implement `Interface` in sort package.
type AddrKeysSlice []AddrKeys

func (b AddrKeysSlice) Len() int {
	return len(b)
}

// Sorts lexographically by Address
func (b AddrKeysSlice) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Address.Bytes(), b[j].Address.Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
	}
}

func (b AddrKeysSlice) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// CreateGenAccounts generates genesis accounts loaded with coins, and returns
// their addresses, pubkeys, and privkeys.
func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (genAccs []*auth.BaseAccount,
	addrs []sdk.AccAddress, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {

	addrKeysSlice := AddrKeysSlice{}

	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		addrKeysSlice = append(addrKeysSlice, NewAddrKeys(addr, pubKey, privKey))
	}

	sort.Sort(addrKeysSlice)

	for i := range addrKeysSlice {
		addrs = append(addrs, addrKeysSlice[i].Address)
		pubKeys = append(pubKeys, addrKeysSlice[i].PubKey)
		privKeys = append(privKeys, addrKeysSlice[i].PrivKey)
		genAccs = append(genAccs, &auth.BaseAccount{
			AccountNumber: uint64(i),
			Address:       addrKeysSlice[i].Address,
			Coins:         genCoins,
			PubKey:        addrKeysSlice[i].PubKey,
		})
	}

	return
}

func newTestWbApp() *WbServiceApp {
	//TODO: should not take config from file using cli flags.
	config, err := vmConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		panic(err)
	}

	return NewWbServiceApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app"), dbm.NewMemDB(), config)
}

func setGenesis(t *testing.T, app *WbServiceApp, accs []*auth.BaseAccount) (sdk.Context, error) {
	genesisState := ModuleBasics.DefaultGenesis()

	ctx := app.NewContext(true, abci.Header{})

	accounts := make(genaccounts.GenesisAccounts, len(accs))
	for idx, _acc := range accs {
		acc := genaccounts.NewGenesisAccount(_acc)
		accounts[idx] = acc
	}
	genesisState["accounts"] = codec.MustMarshalJSONIndent(app.cdc, accounts)

	validators := make(poaTypes.Validators, len(accs))
	for idx, _acc := range accs {
		validators[idx] = poaTypes.Validator{Address: _acc.Address, EthAddress: "0x17f7D1087971dF1a0E6b8Dae7428E97484E32615"}
	}

	genesisState["poa"] = codec.MustMarshalJSONIndent(app.cdc, poaTypes.GenesisState{
		Parameters:    poaTypes.DefaultParams(),
		PoAValidators: validators,
	})

	stateBytes := codec.MustMarshalJSONIndent(app.cdc, genesisState)
	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	return ctx, nil
}

// GenTx generates a signed mock transaction.
func genTx(msgs []sdk.Msg, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	sigs := make([]auth.StdSignature, len(priv))
	memo := "testmemotestmemo"

	fee := auth.StdFee{
		Amount: sdk.Coins{{Denom: "wings", Amount: sdk.NewInt(1)}},
		Gas:    200000,
	}

	for i, p := range priv {
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo))
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, memo)
}

func CheckDeliverTx(t *testing.T, app *WbServiceApp, tx auth.StdTx) {
	res := app.Simulate(app.cdc.MustMarshalJSON(tx), tx)
	require.True(t, res.IsOK(), res.Log)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	res = app.Deliver(tx)
	require.True(t, res.IsOK(), res.Log)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func CheckDeliverErrorTx(t *testing.T, app *WbServiceApp, tx auth.StdTx) {
	res := app.Simulate(app.cdc.MustMarshalJSON(tx), tx)
	require.True(t, !res.IsOK(), res.Log)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	res = app.Deliver(tx)
	require.True(t, !res.IsOK(), res.Log)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func CheckRunQuery(t *testing.T, app *WbServiceApp, requestData interface{}, path string, responseValue interface{}) (resp abci.ResponseQuery) {
	resp = app.Query(abci.RequestQuery{
		Data: codec.MustMarshalJSONIndent(app.cdc, requestData),
		Path: path,
	})
	require.True(t, resp.IsOK())
	if responseValue != nil {
		require.NoError(t, app.cdc.UnmarshalJSON(resp.Value, responseValue))
	}

	return
}

func GetContext(app *WbServiceApp, isCheckTx bool) sdk.Context {
	return app.NewContext(isCheckTx, abci.Header{Height: app.LastBlockHeight() + 1})
}

func GetAccount(app *WbServiceApp, address sdk.AccAddress) auth.Account {
	return app.accountKeeper.GetAccount(app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1}), address)
}
