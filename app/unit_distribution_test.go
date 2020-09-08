// +build unit

package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/cmd/config"
)

// Check disabled distribution transactions.
func TestDistribution_MessagesNotWorking(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genValidators, _, _, _ := CreateGenAccounts(9, GenDefCoins(t))
	nodePrivKey := secp256k1.PrivKeySecp256k1(CheckSetGenesisMockVM(t, app, genValidators))
	nodePubKey := nodePrivKey.PubKey()
	nodeAddress := sdk.AccAddress(nodePubKey.Address())
	valAddress := sdk.ValAddress(nodePubKey.Address())

	// check withdraw rewards tx.
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nodeAddress), nodePrivKey
		// need delegator and validator address.
		withdrawRewardsMsg := distribution.NewMsgWithdrawDelegatorReward(nodeAddress, valAddress)
		tx := GenTx([]sdk.Msg{withdrawRewardsMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, errors.ErrUnknownRequest)
	}

	// check withdraw validator comission tx.
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nodeAddress), nodePrivKey
		withdrawComissionMsg := distribution.NewMsgWithdrawValidatorCommission(valAddress)
		tx := GenTx([]sdk.Msg{withdrawComissionMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, errors.ErrUnknownRequest)
	}

	// check fund community pool.
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nodeAddress), nodePrivKey
		withdrawComissionMsg := distribution.NewMsgFundPublicTreasuryPool(sdk.NewCoins(sdk.NewCoin(config.MainDenom, sdk.NewInt(1))), nodeAddress)
		tx := GenTx([]sdk.Msg{withdrawComissionMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, errors.ErrUnknownRequest)
	}

	// check set withdraw address.
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nodeAddress), nodePrivKey
		withdrawAddress := distribution.NewMsgSetWithdrawAddress(nodeAddress, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		tx := GenTx([]sdk.Msg{withdrawAddress}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, errors.ErrUnknownRequest)
	}
}
