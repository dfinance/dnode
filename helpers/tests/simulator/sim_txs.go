package simulator

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/distribution"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/cmd/config"
)

// GenTxAdvanced returns a signed Tx with a single msg.
func (s *Simulator) GenTxAdvanced(msg sdk.Msg, accNumber, accSeq uint64, pubKey crypto.PubKey, prvKey secp256k1.PrivKeySecp256k1) auth.StdTx {
	memo := fmt.Sprintf("memo_%s", msg.Type())
	msgs := []sdk.Msg{msg}

	fee := auth.StdFee{
		Amount: sdk.NewCoins(s.defFee),
		Gas:    s.defGas,
	}

	signBytes, err := prvKey.Sign(
		auth.StdSignBytes(
			s.chainID, accNumber, accSeq, fee, msgs, memo,
		),
	)
	require.NoError(s.t, err)

	signature := auth.StdSignature{
		PubKey:    pubKey,
		Signature: signBytes,
	}

	return auth.NewStdTx(msgs, fee, []auth.StdSignature{signature}, memo)
}

// GenTx queries account for newer seqNumber and generates a signed Tx.
func (s *Simulator) GenTx(msg sdk.Msg, simAcc *SimAccount) auth.StdTx {
	require.NotNil(s.t, simAcc)

	acc := s.QueryAuthAccount(simAcc.Address)
	require.NotNil(s.t, acc)

	return s.GenTxAdvanced(msg, acc.GetAccountNumber(), acc.GetSequence(), simAcc.PublicKey, simAcc.PrivateKey)
}

// DeliverTx delivers Tx and parses the result.
func (s *Simulator) DeliverTx(tx auth.StdTx, responseValue interface{}) {
	s.beginBlock()

	_, res, err := s.app.Deliver(tx)
	require.NoError(s.t, err)

	s.endBlock()

	if responseValue != nil {
		s.cdc.MustUnmarshalJSON(res.Data, responseValue)
	}
}

// TxStakingCreateValidator creates a new validator operated by simAcc with min self delegation.
func (s *Simulator) TxStakingCreateValidator(simAcc *SimAccount, commissions staking.CommissionRates) {
	require.NotNil(s.t, simAcc)

	selfDelegation := sdk.NewCoin(config.MainDenom, s.minSelfDelegationLvl)
	msg := staking.NewMsgCreateValidator(
		simAcc.Address.Bytes(),
		simAcc.PublicKey,
		selfDelegation,
		staking.NewDescription(simAcc.Address.String(), "", "", "", ""),
		commissions,
		s.minSelfDelegationLvl,
	)
	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxStakingDelegate delegates amount from delegator to validator.
func (s *Simulator) TxStakingDelegate(simAcc *SimAccount, validator *staking.Validator, amount sdk.Coin) {
	require.NotNil(s.t, simAcc)
	require.NotNil(s.t, validator)

	msg := staking.MsgDelegate{
		DelegatorAddress: simAcc.Address,
		ValidatorAddress: validator.OperatorAddress,
		Amount:           amount,
	}
	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxStakingRedelegate redelegates amount from one validator to other.
func (s *Simulator) TxStakingRedelegate(simAcc *SimAccount, valSrc, valDst sdk.ValAddress, amount sdk.Coin) {
	require.NotNil(s.t, simAcc)

	msg := staking.MsgBeginRedelegate{
		DelegatorAddress:    simAcc.Address,
		ValidatorSrcAddress: valSrc,
		ValidatorDstAddress: valDst,
		Amount:              amount,
	}
	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxStakingUndelegate undelegates amount from validator.
func (s *Simulator) TxStakingUndelegate(simAcc *SimAccount, validatorAddr sdk.ValAddress, amount sdk.Coin) {
	require.NotNil(s.t, simAcc)

	msg := staking.MsgUndelegate{
		DelegatorAddress: simAcc.Address,
		ValidatorAddress: validatorAddr,
		Amount:           amount,
	}
	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxDistributionReward taking reward.
func (s *Simulator) TxDistributionReward(simAcc *SimAccount, validatorAddr sdk.ValAddress) {
	require.NotNil(s.t, simAcc)

	msg := distribution.MsgWithdrawDelegatorReward{
		DelegatorAddress: simAcc.Address,
		ValidatorAddress: validatorAddr,
	}

	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}
