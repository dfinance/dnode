package simulator

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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

// CheckTx checks Tx and parses the result.
func (s *Simulator) CheckTx(tx auth.StdTx, responseValue interface{}) error {
	_, res, err := s.app.Check(tx)
	if err != nil {
		return err
	}

	if responseValue != nil {
		s.cdc.MustUnmarshalJSON(res.Data, responseValue)
	}

	return nil
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

// TxStakeCreateValidator creates a new validator operated by simAcc with min self delegation.
func (s *Simulator) TxStakeCreateValidator(simAcc *SimAccount, commissions staking.CommissionRates) {
	require.NotNil(s.t, simAcc)

	selfDelegation := sdk.NewCoin(s.stakingDenom, s.minSelfDelegationLvl)
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

// TxStakeDelegate delegates amount from delegator to validator.
func (s *Simulator) TxStakeDelegate(simAcc *SimAccount, simVal *SimValidator, amount sdk.Coin) (delegationsOverflow bool) {
	require.NotNil(s.t, simAcc)
	require.NotNil(s.t, simVal)

	msg := staking.MsgDelegate{
		DelegatorAddress: simAcc.Address,
		ValidatorAddress: simVal.GetAddress(),
		Amount:           amount,
	}
	tx := s.GenTx(msg, simAcc)

	if err := s.CheckTx(tx, nil); err != nil {
		if staking.ErrMaxDelegationsLimit.Is(err) {
			delegationsOverflow = true
			return
		}
		require.NoError(s.t, err)
	}

	s.DeliverTx(tx, nil)

	return
}

// TxStakeRedelegate redelegates amount from one validator to other.
func (s *Simulator) TxStakeRedelegate(simAcc *SimAccount, valSrc, valDst sdk.ValAddress, amount sdk.Coin) {
	require.NotNil(s.t, simAcc)

	msg := staking.MsgBeginRedelegate{
		DelegatorAddress:    simAcc.Address,
		ValidatorSrcAddress: valSrc,
		ValidatorDstAddress: valDst,
		Amount:              amount,
	}
	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxStakeUndelegate undelegates amount from validator.
func (s *Simulator) TxStakeUndelegate(simAcc *SimAccount, validatorAddr sdk.ValAddress, amount sdk.Coin) {
	require.NotNil(s.t, simAcc)

	msg := staking.MsgUndelegate{
		DelegatorAddress: simAcc.Address,
		ValidatorAddress: validatorAddr,
		Amount:           amount,
	}
	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxDistDelegatorRewards withdraws delegator rewards.
func (s *Simulator) TxDistDelegatorRewards(simAcc *SimAccount, validatorAddr sdk.ValAddress) {
	require.NotNil(s.t, simAcc)

	msg := distribution.MsgWithdrawDelegatorReward{
		DelegatorAddress: simAcc.Address,
		ValidatorAddress: validatorAddr,
	}

	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxDistValidatorCommission withdraws validator commission rewards.
func (s *Simulator) TxDistValidatorCommission(simAcc *SimAccount, validatorAddr sdk.ValAddress) (noCommission bool) {
	require.NotNil(s.t, simAcc)

	msg := distribution.MsgWithdrawValidatorCommission{
		ValidatorAddress: validatorAddr,
	}
	tx := s.GenTx(msg, simAcc)

	if err := s.CheckTx(tx, nil); err != nil {
		if distribution.ErrNoValidatorCommission.Is(err) {
			noCommission = true
			return
		}
		require.NoError(s.t, err)
	}

	s.DeliverTx(s.GenTx(msg, simAcc), nil)

	return
}

// TxDistLockRewards locks validator rewards.
func (s *Simulator) TxDistLockRewards(simAcc *SimAccount, validatorAddr sdk.ValAddress) {
	require.NotNil(s.t, simAcc)

	msg := distribution.MsgLockValidatorRewards{
		ValidatorAddress: validatorAddr,
		SenderAddress:    simAcc.Address,
	}

	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}

// TxDistDisableAutoRenewal disables validator rewards lock auto-renewal.
func (s *Simulator) TxDistDisableAutoRenewal(simAcc *SimAccount, validatorAddr sdk.ValAddress) {
	require.NotNil(s.t, simAcc)

	msg := distribution.MsgDisableLockedRewardsAutoRenewal{
		ValidatorAddress: validatorAddr,
		SenderAddress:    simAcc.Address,
	}

	s.DeliverTx(s.GenTx(msg, simAcc), nil)
}
