package simulator

import (
	"encoding/json"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/app"
	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/genaccounts"
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
	useInMemDB           bool
	minBlockDur          time.Duration
	maxBlockDur          time.Duration
	// predefined settings
	chainID   string
	monikerID string
	defFee    sdk.Coin
	defGas    uint64
	// read-only params
	unbondingDur               time.Duration
	mainDenom                  string
	stakingDenom               string
	mainDenomDecimals          uint8
	stakingDenomDecimals       uint8
	workingDir                 string
	mainAmountDecimalsRatio    sdk.Dec
	stakingAmountDecimalsRatio sdk.Dec
	// state
	prevBlockTime time.Time
	t             *testing.T
	cdc           *codec.Codec
	logger        log.Logger
	app           *app.DnServiceApp
	defferQueue   *DefferOps
	// stat
	counter Counter
}

type Counter struct {
	Delegations          int64
	Undelegations        int64
	Redelegations        int64
	Rewards              int64
	Commissions          int64
	RewardsCollected     sdk.Int
	CommissionsCollected sdk.Int
}

// BuildTmpFilePath builds file name inside of the Simulator working dir.
func (s *Simulator) BuildTmpFilePath(fileName string) string {
	return path.Join(s.workingDir, fileName)
}

// Start creates the genesisState and perform ChainInit.
func (s *Simulator) Start() {
	require.GreaterOrEqual(s.t, len(s.accounts), 1)

	// generate wallet accounts
	genAccs := make(genaccounts.GenesisState, 0)
	poaAccs := make([]*SimAccount, 0)
	for accIdx := 0; accIdx < len(s.accounts); accIdx++ {
		acc := s.accounts[accIdx]

		// gen unique privKey
		for {
			acc.PrivateKey = secp256k1.GenPrivKey()
			//
			isUnique := true
			for idx := 0; idx < len(s.accounts); idx++ {
				if idx == accIdx {
					continue
				}
				if s.accounts[idx].PrivateKey.Equals(acc.PrivateKey) {
					isUnique = false
					break
				}
			}
			if isUnique {
				break
			}
		}

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

		state.Params.BondDenom = s.stakingDenom
		state.Params.MinSelfDelegationLvl = s.minSelfDelegationLvl

		s.genesisState[staking.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, state)

		s.unbondingDur = state.Params.UnbondingTime
	}
	// mint
	{
		state := mint.GenesisState{}
		s.cdc.MustUnmarshalJSON(s.genesisState[mint.ModuleName], &state)

		state.Params.MintDenom = s.stakingDenom

		s.genesisState[mint.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, state)
	}
	// crisis
	{
		state := crisis.GenesisState{}
		s.cdc.MustUnmarshalJSON(s.genesisState[crisis.ModuleName], &state)

		defFeeAmount, ok := sdk.NewIntFromString(config.DefaultFeeAmount)
		require.True(s.t, ok)

		state.ConstantFee.Denom = s.mainDenom
		state.ConstantFee.Amount = defFeeAmount

		s.genesisState[crisis.ModuleName] = codec.MustMarshalJSONIndent(s.cdc, state)

		s.defFee = sdk.NewCoin(s.mainDenom, defFeeAmount)
	}
	// genutil, create node validator
	{
		nodeAcc := s.accounts[0]

		selfDelegation := sdk.NewCoin(s.stakingDenom, s.minSelfDelegationLvl)
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
	validators := s.QueryStakeValidators(1, 10, sdk.BondStatusBonded)
	require.Len(s.t, validators, 1)
	s.accounts[0].OperatedValidator = &validators[0]

	// update node account delegations
	delegation := s.QueryStakeDelegation(s.accounts[0], s.accounts[0].OperatedValidator)
	s.accounts[0].Delegations = append(s.accounts[0].Delegations, delegation)

	s.t.Logf("Simulator working / output directory: %s", s.workingDir)
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

	s.defferQueue.Exec(s.prevBlockTime)
}

// BeginBlock starts a new block with blockTime [5s:7s] and randomly selected validator.
func (s *Simulator) beginBlock() {
	// calculate next block height and time
	nextHeight := s.app.LastBlockHeight() + 1
	nextBlockDur := s.minBlockDur + time.Duration(rand.Int63n(int64(s.maxBlockDur-s.minBlockDur)))
	nextBlockTime := s.prevBlockTime.Add(nextBlockDur)
	s.prevBlockTime = nextBlockTime

	// pick next proposer
	validators := s.GetValidators(true, false, false)
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
func NewSimulator(t *testing.T, workingDir string, defferQueue *DefferOps, options ...SimOption) *Simulator {
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
		accounts:    make([]*SimAccount, 0),
		minBlockDur: 5 * time.Second,
		maxBlockDur: 6 * time.Second,
		//
		chainID:   "simChainID",
		monikerID: "simMoniker",
		defGas:    500000,
		//
		mainDenom:                  config.MainDenom,
		stakingDenom:               config.StakingDenom,
		mainDenomDecimals:          18,
		stakingDenomDecimals:       18,
		mainAmountDecimalsRatio:    sdk.NewDecWithPrec(1, 0),
		stakingAmountDecimalsRatio: sdk.NewDecWithPrec(1, 0),
		workingDir:                 workingDir,
		//
		prevBlockTime: EmulatedTimeHead,
		logger:        log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("simulator"),
		t:             t,
		cdc:           app.MakeCodec(),
		defferQueue:   defferQueue,
	}
	s.counter.RewardsCollected = sdk.ZeroInt()
	s.counter.CommissionsCollected = sdk.ZeroInt()

	for _, option := range options {
		option(s)
	}

	if s.mainDenomDecimals > 0 {
		s.mainAmountDecimalsRatio = sdk.NewDecWithPrec(1, int64(s.mainDenomDecimals))
	}
	if s.stakingDenomDecimals > 0 {
		s.stakingAmountDecimalsRatio = sdk.NewDecWithPrec(1, int64(s.stakingDenomDecimals))
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

	// create DB
	var db dbm.DB
	if s.useInMemDB {
		db = dbm.NewMemDB()
	} else {
		db = dbm.NewDB("simulator", dbm.GoLevelDBBackend, s.workingDir)
	}

	// set application
	s.app = app.NewDnServiceApp(logger, db, vmConfig, s.invariantCheckPeriod, config.AppRestrictions{})

	return s
}
