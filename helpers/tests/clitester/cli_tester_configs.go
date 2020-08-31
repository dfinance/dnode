package clitester

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/dfinance/glav"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/orders"
)

type DirConfig struct {
	RootDir  string
	DncliDir string
	UDSDir   string
}

func NewTempDirConfig(testName string) (c DirConfig, retErr error) {
	rootDir, err := ioutil.TempDir("/tmp", fmt.Sprintf("wd-cli-test-%s-", testName))
	if err != nil {
		retErr = fmt.Errorf("creating TempDir: %w", err)
		return
	}

	dncliDir := path.Join(rootDir, "dncli")
	udsDir := path.Join(rootDir, "sockets")

	if err := os.Mkdir(udsDir, 0777); err != nil {
		retErr = fmt.Errorf("creating sockets dir: %w", err)
		return
	}

	c.RootDir = rootDir
	c.DncliDir = dncliDir
	c.UDSDir = udsDir

	return
}

type NodeIdConfig struct {
	ChainID   string
	MonikerID string
}

func NewTestNodeIdConfig() NodeIdConfig {
	return NodeIdConfig{
		ChainID:   "test-chain",
		MonikerID: "test-moniker",
	}
}

type BinaryPathConfig struct {
	wbd   string
	wbcli string
}

func NewTestBinaryPathConfig() BinaryPathConfig {
	return BinaryPathConfig{
		wbd:   "dnode",
		wbcli: "dncli",
	}
}

type CurrencyInfo struct {
	Decimals       uint8
	BalancePathHex string
	BalancePath    []byte
	InfoPathHex    string
	InfoPath       []byte
}

func NewCurrencyMap(cdc *codec.Codec, state GenesisState) map[string]CurrencyInfo {
	currencies := make(map[string]CurrencyInfo)

	var ccsGenesis ccstorage.GenesisState
	cdc.MustUnmarshalJSON(state[ccstorage.ModuleName], &ccsGenesis)

	for _, params := range ccsGenesis.CurrenciesParams {
		info := CurrencyInfo{
			Decimals:       params.Decimals,
			BalancePath:    glav.BalanceVector(params.Denom),
			InfoPath:       glav.CurrencyInfoVector(params.Denom),
			BalancePathHex: hex.EncodeToString(glav.BalanceVector(params.Denom)),
			InfoPathHex:    hex.EncodeToString(glav.CurrencyInfoVector(params.Denom)),
		}

		currencies[params.Denom] = info
	}

	return currencies
}

type CLIAccount struct {
	Name            string
	Address         string
	EthAddress      string
	PubKey          string
	Mnemonic        string
	Number          uint64
	Coins           map[string]sdk.Coin
	IsModuleAcc     bool
	IsPOAValidator  bool
	IsOracleNominee bool
	IsOracle        bool
}

func NewAccountMap() (accounts map[string]*CLIAccount, retErr error) {
	accounts = make(map[string]*CLIAccount)

	smallAmount, ok := sdk.NewIntFromString("1000000000000000000000") // 1000xfi
	if !ok {
		retErr = fmt.Errorf("NewInt for smallAmount")
		return
	}

	bigAmount, ok := sdk.NewIntFromString("1000000000000000000000000") // 1000000xfi
	if !ok {
		retErr = fmt.Errorf("NewInt for bigAmount")
		return
	}

	accounts["pos"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			DenomSXFI:        sdk.NewCoin(DenomSXFI, bigAmount),
			config.MainDenom: sdk.NewCoin(config.MainDenom, bigAmount),
		},
	}
	accounts["bank"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, bigAmount),
		},
	}
	accounts["validator1"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
			DenomSXFI:        sdk.NewCoin(DenomSXFI, smallAmount),
		},
		IsPOAValidator: true,
	}
	accounts["validator2"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	accounts["validator3"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	accounts["validator4"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	accounts["validator5"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	accounts["nominee"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracleNominee: true,
	}
	accounts["oracle1"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracle: true,
	}
	accounts["oracle2"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracle: false,
	}
	accounts["oracle3"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracle: false,
	}
	accounts["plain"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
	}
	accounts[orders.ModuleName] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsModuleAcc: true,
	}
	accounts[gov.ModuleName] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsModuleAcc: true,
	}
	accounts[DenomSXFI] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
			DenomSXFI:        sdk.NewCoin(DenomSXFI, smallAmount),
		},
	}

	return
}

type NodePortConfig struct {
	RPCPort    string
	RPCAddress string
	P2PPort    string
	P2PAddress string
}

func NewTestNodePortConfig() (c NodePortConfig, retErr error) {
	srvAddr, srvPort, err := server.FreeTCPAddr()
	if err != nil {
		retErr = fmt.Errorf("FreeTCPAddr for srv: %w", err)
		return
	}

	p2pAddr, p2pPort, err := server.FreeTCPAddr()
	if err != nil {
		retErr = fmt.Errorf("FreeTCPAddr for p2p: %w", err)
		return
	}

	c.RPCAddress, c.RPCPort = srvAddr, srvPort
	c.P2PAddress, c.P2PPort = p2pAddr, p2pPort

	return
}

type VMConnectionConfig struct {
	BaseAddress     string
	ListenPort      string
	ListenAddress   string
	ConnectPort     string
	ConnectAddress  string
	CompilerAddress string
}

func NewTestVMConnectionConfigTCP() (c VMConnectionConfig, retErr error) {
	_, listenPort, err := server.FreeTCPAddr()
	if err != nil {
		retErr = fmt.Errorf("FreeTCPAddr for VM listen: %w", err)
		return
	}
	_, connectPort, err := server.FreeTCPAddr()
	if err != nil {
		retErr = fmt.Errorf("FreeTCPAddr for VM connect: %w", err)
		return
	}

	baseAddress := "127.0.0.1"
	connectAddress := fmt.Sprintf("%s:%s", baseAddress, connectPort)
	listenAddress := fmt.Sprintf("%s:%s", baseAddress, listenPort)

	c.BaseAddress = baseAddress
	c.ListenPort, c.ListenAddress = listenPort, listenAddress
	c.ConnectPort, c.ConnectAddress = connectPort, connectAddress
	c.CompilerAddress = c.ConnectAddress

	return
}

type VMCommunicationConfig struct {
	MaxAttempts    uint
	ReqTimeoutInMs uint
}

func NewTestVMCommunicationConfig() VMCommunicationConfig {
	return VMCommunicationConfig{
		MaxAttempts:    1,
		ReqTimeoutInMs: 5000,
	}
}

type ConsensusTimingConfig struct {
	UseDefaults           bool
	TimeoutPropose        string
	TimeoutProposeDelta   string
	TimeoutPreVote        string
	TimeoutPreVoteDelta   string
	TimeoutPreCommit      string
	TimeoutPreCommitDelta string
	TimeoutCommit         string
}

func NewTestConsensusTimingConfig() ConsensusTimingConfig {
	return ConsensusTimingConfig{
		UseDefaults:           false,
		TimeoutPropose:        "250ms",
		TimeoutProposeDelta:   "250ms",
		TimeoutPreVote:        "250ms",
		TimeoutPreVoteDelta:   "250ms",
		TimeoutPreCommit:      "250ms",
		TimeoutPreCommitDelta: "250ms",
		TimeoutCommit:         "250ms",
	}
}

type GovernanceConfig struct {
	MinVotingDur time.Duration
}

func NewGovernanceConfig() GovernanceConfig {
	return GovernanceConfig{
		MinVotingDur: 10 * time.Second,
	}
}

type MempoolConfig struct {
	UseDefault  bool
	Size        int64
	CacheSize   int64
	MaxTxBytes  int64
	MaxTxsBytes int64
}

func NewMempoolConfig() MempoolConfig {
	return MempoolConfig{
		UseDefault:  true,
		Size:        5000,
		CacheSize:   10000,
		MaxTxBytes:  1048576,
		MaxTxsBytes: 1073741824,
	}
}
