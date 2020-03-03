package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	rest2 "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	sdkutils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	wbcnf "github.com/WingsDao/wings-blockchain/cmd/config"
	"github.com/WingsDao/wings-blockchain/oracle-app/internal/exchange"
	"github.com/WingsDao/wings-blockchain/oracle-app/internal/utils"
	"github.com/WingsDao/wings-blockchain/x/oracle"
)

const (
	accountName = "oracle"
	passphrase  = "12345678"
)

type Client struct {
	nodeAddress string
	chainID     string
	fees        sdk.Coins

	keyBase keys.Keybase
	keyInfo keys.Info

	cl *http.Client

	cdc       *codec.Codec
	txBuilder auth.TxBuilder
}

func init() {
	config := sdk.GetConfig()
	wbcnf.InitBechPrefixes(config)
	config.Seal()
}

func NewClient(mnemonic string, account, index uint32, gas uint64, chainID string, nodeAddress string, fees sdk.Coins) (*Client, error) {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	sdk.RegisterCodec(cdc)
	oracle.RegisterCodec(cdc)

	kb := keys.NewInMemory()
	ki, err := kb.CreateAccount(accountName, mnemonic, "", passphrase, account, index)
	fmt.Printf("Client address is %s\n", ki.GetAddress())
	if err != nil {
		return nil, err
	}
	cl := &http.Client{
		Timeout: time.Second * 10,
	}

	txBuilder := auth.NewTxBuilder(sdkutils.GetTxEncoder(cdc), 0, 0, gas, 0, false, chainID, "", fees, nil).WithKeybase(kb)

	return &Client{keyBase: kb, keyInfo: ki, cl: cl, nodeAddress: nodeAddress, cdc: cdc, chainID: chainID, fees: fees, txBuilder: txBuilder}, err
}

func (c *Client) PostPrice(t exchange.Ticker) error {
	intPrice, err := utils.NewIntFromString(t.Price, 8)
	if err != nil {
		return err
	}
	broadcastReq := rest2.BroadcastReq{Mode: "block"}

	acc, err := c.getAccount()
	if err != nil {
		return err
	}
	msgSigned, err := c.txBuilder.
		WithAccountNumber(acc.AccountNumber).
		WithSequence(acc.Sequence).
		WithChainID(c.chainID).
		BuildAndSign(accountName, passphrase, []sdk.Msg{oracle.NewMsgPostPrice(acc.Address, t.Asset.Code, intPrice, time.Now().Add(time.Hour))})
	if err != nil {
		return err
	}
	err = c.cdc.UnmarshalBinaryLengthPrefixed(msgSigned, &broadcastReq.Tx)
	if err != nil {
		return err
	}
	bz, err := codec.MarshalJSONIndent(c.cdc, broadcastReq)
	if err != nil {
		return err
	}
	bz, err = sdk.SortJSON(bz)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/txs", c.nodeAddress)
	apiReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bz))
	if err != nil {
		return err
	}
	resp, err := c.cl.Do(apiReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	exchange.Logger().Debug(string(body))

	return nil
}

func (c *Client) getAccount() (*auth.BaseAccount, error) {
	url := fmt.Sprintf("http://%s/auth/accounts/%s", c.nodeAddress, c.keyInfo.GetAddress())
	resp, err := c.cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	var rd rest.ResponseWithHeight
	err = c.cdc.UnmarshalJSON(bz, &rd)
	if err != nil {
		return nil, err
	}
	var acc = struct {
		Type  string           `json:"type"`
		Value auth.BaseAccount `json:"value"`
	}{}
	err = c.cdc.UnmarshalJSON(rd.Result, &acc)
	if err != nil {
		return nil, err
	}

	return &acc.Value, err
}
