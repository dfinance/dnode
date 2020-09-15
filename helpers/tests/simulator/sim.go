package simulator

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/app"
	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/genaccounts"
	"github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/poa"
)

var (
	EmulatedTimeHead = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
)

type Simulator struct {
	// configurable settings
	genesisState         map[string]json.RawMessage
	invariantCheckPeriod uint
	logOptions           []log.Option
	minSelfDelegationLvl sdk.Int
	nodeValidatorConfig  SimValidatorConfig
	operations           []*SimOperation
	accounts             []*SimAccount
	// predefined settings
	chainID   string
	monikerID string
	defFee    sdk.Coin
	defGas    uint64
	// state
	prevBlockTime time.Time
	t             *testing.T
	cdc           *codec.Codec
	logger        log.Logger
	app           *app.DnServiceApp
	// stat
	counter Counter
}

type Counter struct {
	Delegations   int64
	Undelegations int64
	Redelegations int64
	Rewards       int64
}

// Start creates the genesisState and perform ChainInit.
func (s *Simulator) Start() {
	require.GreaterOrEqual(s.t, len(s.accounts), 1)

	// generate wallet accounts
	genAccs := make(genaccounts.GenesisState, 0)
	poaAccs := make([]*SimAccount, 0)
	for accIdx := 0; accIdx < len(s.accounts); accIdx++ {
		acc := s.accounts[accIdx]
		acc.PrivateKey = secp256k1.GenPrivKey()
		acc.PublicKey = acc.PrivateKey.PubKey()
		acc.Address = sdk.AccAddress(acc.PublicKey.Address())
		acc.Number = uint64(len(s.accounts) + accIdx)

		baseAcc := &auth.BaseAccount{
			Address:       acc.Address,
			Coins:         acc.Coins,
			PubKey:        acc.PublicKey,
			AccountNumber: acc.Number,
		}

		genAcc, err := genaccounts.NewGenesisAccountI(baseAcc)
		require.NoError(s.t, err, "wallet account [%d]", accIdx)
		genAccs = append(genAccs, genAcc)

		if acc.IsPoAValidator {
			poaAccs = append(poaAccs, acc)
		}
	}

	// generate module accounts
	// gov module
	{
		prvKey := secp256k1.GenPrivKey()
		pubKey := prvKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		baseAcc := &auth.BaseAccount{
			Address:       addr,
			Coins:         sdk.NewCoins(),
			PubKey:        pubKey,
			AccountNumber: uint64(len(s.accounts)*2 + 0),
		}

		genAcc, err := genaccounts.NewGenesisAccountI(supply.NewModuleAccount(baseAcc, gov.ModuleName, supply.Burner))
		require.NoError(s.t, err, "module account: gov")
		genAccs = append(genAccs, genAcc)
	}
	// orders module
	{
		prvKey := secp256k1.GenPrivKey()
		pubKey := prvKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		baseAcc := &auth.BaseAccount{
			Address:       addr,
			Coins:         sdk.NewCoins(),
			PubKey:        pubKey,
			AccountNumber: uint64(len(s.accounts)*2 + 1),
		}

		genAcc, err := genaccounts.NewGenesisAccountI(supply.NewModuleAccount(baseAcc, orders.ModuleName))
		require.NoError(s.t, err, "module account: gov")
		genAccs = append(genAccs, genAcc)
	}

	// update genesisState
	// genAccounts
	{
		s.genesisState[genaccounts.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, genAccs)
	}
	// poa
	{
		validators := make(poa.Validators, 0, len(poaAccs))
		for _, acc := range poaAccs {
			validators = append(validators, poa.Validator{
				Address:    acc.Address,
				EthAddress: "0x17f7D1087971dF1a0E6b8Dae7428E97484E32615",
			})
		}
		s.genesisState[poa.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, poa.GenesisState{
			Parameters: poa.DefaultParams(),
			Validators: validators,
		})
	}
	// staking
	{
		state := staking.GenesisState{}
		s.cdc.MustUnmarshalJSON(s.genesisState[staking.ModuleName], &state)

		state.Params.BondDenom = config.MainDenom
		state.Params.MinSelfDelegationLvl = s.minSelfDelegationLvl
		state.Params.UnbondingTime = 15 * time.Hour

		s.genesisState[staking.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, state)
	}
	// mint
	{
		state := mint.GenesisState{}
		s.cdc.MustUnmarshalJSON(s.genesisState[mint.ModuleName], &state)

		state.Params.MintDenom = config.MainDenom

		s.genesisState[mint.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, state)
	}
	// crisis
	{
		state := crisis.GenesisState{}
		s.cdc.MustUnmarshalJSON(s.genesisState[crisis.ModuleName], &state)

		defFeeAmount, ok := sdk.NewIntFromString(config.DefaultFeeAmount)
		require.True(s.t, ok)

		state.ConstantFee.Denom = config.MainDenom
		state.ConstantFee.Amount = defFeeAmount

		s.genesisState[crisis.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, state)

		s.defFee = sdk.NewCoin(config.MainDenom, defFeeAmount)
	}
	// genutil, create node validator
	{
		nodeAcc := s.accounts[0]

		selfDelegation := sdk.NewCoin(config.MainDenom, s.minSelfDelegationLvl)
		msg := staking.NewMsgCreateValidator(
			nodeAcc.Address.Bytes(),
			nodeAcc.PublicKey,
			selfDelegation,
			staking.NewDescription(s.monikerID, "", "", "", ""),
			s.nodeValidatorConfig.Commission,
			s.minSelfDelegationLvl,
		)
		tx := s.GenTxAdvanced(msg, 0, 0, nodeAcc.PublicKey, nodeAcc.PrivateKey)

		s.genesisState[genutil.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, genutil.NewGenesisStateFromStdTx([]auth.StdTx{tx}))
	}

	// validate genesis
	require.NoError(s.t, app.ModuleBasics.ValidateGenesis(s.genesisState))
	genesisStateBz, err := codec.MarshalJSONIndent(s.cdc, s.genesisState)
	require.NoError(s.t, err)

	// init chain
	s.app.InitChain(
		abci.RequestInitChain{
			ChainId:       s.chainID,
			AppStateBytes: genesisStateBz,
		},
	)
	s.app.Commit()

	// get node validator
	validators := s.QueryStakingValidators(1, 10, sdk.BondStatusBonded)
	require.Len(s.t, validators, 1)
	s.accounts[0].OperatedValidator = &validators[0]

	// update node account delegations
	delegation := s.QueryStakingDelegation(s.accounts[0], s.accounts[0].OperatedValidator)
	s.accounts[0].AddDelegation(&delegation)
}

// GetCheckCtx returns a new CheckTx Context.
func (s *Simulator) GetCheckCtx() sdk.Context {
	return s.app.NewContext(true, abci.Header{Height: s.app.LastBlockHeight() + 1})
}

// GetDeliverCtx returns a new DeliverTx Context.
func (s *Simulator) GetDeliverCtx() sdk.Context {
	return s.app.NewContext(false, abci.Header{Height: s.app.LastBlockHeight() + 1})
}

// SimulatedDur returns current simulated duration and last block height.
func (s *Simulator) SimulatedDur() (int64, time.Duration) {
	return s.app.LastBlockHeight(), s.prevBlockTime.Sub(EmulatedTimeHead)
}

// Next creates a new block(s): single (no operations) / multiple.
func (s *Simulator) Next() {
	blockCreated := false

	for _, op := range s.operations {
		blockCreated = op.Exec(s, s.prevBlockTime)
	}

	if !blockCreated {
		s.beginBlock()
		s.endBlock()
	}
}

// BeginBlock starts a new block with blockTime [5s:7s] and randomly selected validator.
func (s *Simulator) beginBlock() {
	const (
		minBlockDur = 5 * time.Second
		maxBlockDur = 7 * time.Second
	)

	// calculate next block height and time
	nextHeight := s.app.LastBlockHeight() + 1
	nextBlockDur := minBlockDur + time.Duration(rand.Int63n(int64(maxBlockDur-minBlockDur)))
	nextBlockTime := s.prevBlockTime.Add(nextBlockDur)
	s.prevBlockTime = nextBlockTime

	// pick next proposer
	validators := s.GetValidators()
	proposerIdx := rand.Intn(len(validators))
	proposer := validators[proposerIdx]

	// set TM voteInfos
	lastCommitInfo := abci.LastCommitInfo{}
	for _, val := range validators {
		lastCommitInfo.Votes = append(lastCommitInfo.Votes, abci.VoteInfo{
			Validator: abci.Validator{
				Address: val.ConsPubKey.Address(),
				Power:   val.GetConsensusPower(),
			},
			SignedLastBlock: true,
		})
	}

	s.app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			ChainID:         s.chainID,
			Height:          nextHeight,
			Time:            nextBlockTime,
			ProposerAddress: proposer.ConsPubKey.Address(),
		},
		LastCommitInfo: lastCommitInfo,
	})
}

// EndBlock ends the current block and checks if it is time to report.
func (s *Simulator) endBlock() {
	s.app.EndBlock(abci.RequestEndBlock{})
	s.app.Commit()
}

// NewSimulator creates a new Simulator.
func NewSimulator(t *testing.T, options ...SimOption) *Simulator {
	// defaults init
	minSelfDelegation, ok := sdk.NewIntFromString(config.DefMinSelfDelegation)
	require.True(t, ok)

	nodeValCommissionRate, err := sdk.NewDecFromStr("0.100000000000000000")
	require.NoError(t, err)

	nodeValCommissionMaxRate, err := sdk.NewDecFromStr("0.200000000000000000")
	require.NoError(t, err)

	nodeValCommissionMaxChangeRate, err := sdk.NewDecFromStr("0.010000000000000000")
	require.NoError(t, err)

	s := &Simulator{
		genesisState:         app.ModuleBasics.DefaultGenesis(),
		invariantCheckPeriod: 1,
		logOptions:           make([]log.Option, 0),
		minSelfDelegationLvl: minSelfDelegation,
		nodeValidatorConfig: SimValidatorConfig{
			Commission: staking.CommissionRates{
				Rate:          nodeValCommissionRate,
				MaxRate:       nodeValCommissionMaxRate,
				MaxChangeRate: nodeValCommissionMaxChangeRate,
			},
		},
		accounts: make([]*SimAccount, 0),
		//
		chainID:   "simChainID",
		monikerID: "simMoniker",
		defGas:    500000,
		//
		prevBlockTime: EmulatedTimeHead,
		logger:        log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("simulator"),
		t:             t,
		cdc:           app.MakeCodec(),
	}

	for _, option := range options {
		option(s)
	}

	// set app logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	s.logOptions = append(s.logOptions, log.AllowError())
	logger = log.NewFilter(logger, s.logOptions...)

	// set mock VM config
	vmConfig := &config.VMConfig{
		Address:        "127.0.0.1:0",
		DataListen:     "127.0.0.1:0",
		MaxAttempts:    0,
		ReqTimeoutInMs: 0,
	}

	// set application
	s.app = app.NewDnServiceApp(logger, dbm.NewMemDB(), vmConfig, s.invariantCheckPeriod, app.AppRestrictions{})

	return s
}
