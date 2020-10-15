//+build integ

package app

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/compiler_grpc"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
)

const swapModuleSrc = `
module Swap {
    use 0x1::Dfinance;
    use 0x1::Account;
    use 0x1::Signer;

    // The resource of module which contains swap parameters.
    resource struct T<Offered, Expected>{
        offered: Dfinance::T<Offered>,
        price: u128,
    }

    // Create a swap deal with two coin pairs: Offered and Expected.
    public fun create<Offered, Expected>(sender: &signer, offered: Dfinance::T<Offered>, price: u128) {
        let sender_addr = Signer::address_of(sender);

        assert(!exists_at<Offered, Expected>(sender_addr), 101);

        move_to<T<Offered, Expected>>(
            sender,
            T<Offered, Expected> {
                offered: offered,
                price
            }
        );
    }

    // Get the price of the swap deal.
    public fun get_price<Offered, Expected>(seller: address): u128 acquires T {
        let offer = borrow_global<T<Offered, Expected>>(seller);
        offer.price
    }

    // Change price before swap happens.
    public fun change_price<Offered, Expected>(sender: &signer, new_price: u128) acquires T {
        let offer = borrow_global_mut<T<Offered, Expected>>(Signer::address_of(sender));
        offer.price = new_price;
    }

    // Swap coins and deposit them to accounts: both creator and buyer.
    public fun swap<Offered, Expected>(sender: &signer, seller: address, exp: Dfinance::T<Expected>) acquires T {
       let T<Offered, Expected> { offered, price } = move_from<T<Offered, Expected>>(seller);
       let exp_value = Dfinance::value<Expected>(&exp);

       assert(exp_value == price, 102);
       Account::deposit(sender, seller, exp);
       Account::deposit_to_sender(sender, offered);
    }

    // Check if the swap pair already exists for the account.
    public fun exists_at<Offered, Expected>(addr: address): bool {
        exists<T<Offered, Expected>>(addr)
    }
}
`

const createSwapScriptSrcFmt = `
script {
    use {{sender}}::Swap;
    use 0x1::XFI;
    use 0x1::Coins;
    use 0x1::Account;

    fun main(sender: &signer, amount: u128, price: u128) {
        let xfi = Account::withdraw_from_sender(sender, amount);

        // Deposit XFI coins in exchange to BTC.
        Swap::create<XFI::T, Coins::BTC>(sender, xfi, price);
    }
}
`

const swapSwapScriptSrcFmt = `
script {
    use {{sender}}::Swap;
    use 0x1::XFI;
    use 0x1::Coins;
    use 0x1::Account;

    fun main(sender: &signer, seller: address, price: u128) {
        let btc = Account::withdraw_from_sender(sender, price);

        // Deposit BTC to swap coins.
        Swap::swap<XFI::T, Coins::BTC>(sender, seller, btc);
    }
}
`

// Test checks Swap Move module without crisis module panic (checks vmauth <-> ccstorage integration).
// 1. Issue BTCs to client2
// 2. Create Swap to exchange client1 XFIs for BTCs (client1's XFIs are locked within Move module)
// 3. Execute Swap transferring BTCs to client1 and XFIs to client2
// 4. Verify balances are updated
func TestIntegApp_Crisis(t *testing.T) {
	app, dvmAddr, appStop := NewTestDnAppDVM(t, log.AllowInfoWith("module", "x/crisis"))
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(3, GenDefCoins(t))
	CheckSetGenesisDVM(t, app, genAccs)

	skipBlock := func() {
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	client1Idx, client2Idx := uint(0), uint(1)
	client1Addr, client2Addr := genAccs[client1Idx].Address, genAccs[client2Idx].Address
	client1PrivKey, client2PrivKey := genPrivKeys[client1Idx], genPrivKeys[client2Idx]
	client1LibraAddr := common_vm.Bech32ToLibra(client1Addr)

	verboseSuppliesDiff := func(diffs []string) []string {
		for i := 0; i < len(diffs); i++ {
			diffs[i] = strings.ReplaceAll(diffs[i], client1Addr.String(), "client1")
			diffs[i] = strings.ReplaceAll(diffs[i], client2Addr.String(), "client2")
		}
		return diffs
	}

	getXfiBtcAccCoins := func(addr sdk.AccAddress) (sdk.Coin, sdk.Coin) {
		xfiCoin := sdk.NewCoin(defaults.MainDenom, sdk.ZeroInt())
		btcCoin := sdk.NewCoin("btc", sdk.ZeroInt())
		acc := GetAccountCheckTx(app, addr)
		for _, coin := range acc.GetCoins() {
			switch coin.Denom {
			case xfiCoin.Denom:
				xfiCoin = coin
			case btcCoin.Denom:
				btcCoin = coin
			}
		}
		return xfiCoin, btcCoin
	}

	// compile and deploy module
	{
		// compile
		byteCode, compileErr := vm_client.Compile(dvmAddr, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: swapModuleSrc,
					Name: "swapModuleSrc",
				},
			},
			Address: client1LibraAddr,
		})

		require.NoError(t, compileErr)
		require.Len(t, byteCode, 1)

		// deploy using helper func
		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), client1PrivKey
		deployMsg := vm.MsgDeployModule{
			Signer: client1Addr,
			Module: byteCode[0].ByteCode,
		}
		tx := GenTx([]sdk.Msg{deployMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// issue 1.0 btc to client2
	{
		// supplies before the Tx
		suppliesBefore := GetAllSupplies(t, app, GetContext(app, true))

		amount, _ := sdk.NewIntFromString("100000000")
		coin := sdk.NewCoin("btc", amount)
		IssueCurrency(t, app, coin, "1", "issue1", client2Idx, genAccs, genPrivKeys, true)

		// supplies after the Tx
		suppliesDiff := suppliesBefore.GetDiffString(GetAllSupplies(t, app, GetContext(app, true)))
		t.Logf(">> Issue 1.0 btc to client2, supply diff:\n%s", strings.Join(verboseSuppliesDiff(suppliesDiff), "\n"))
	}

	// client1 offers 1.0 XFI for 0.5 BTC
	offerAmount, _ := sdk.NewIntFromString("1000000000000000000")
	priceAmount, _ := sdk.NewIntFromString("50000000")

	// save client1 balances before Swap lock
	client1XfiBeforeLock, _ := getXfiBtcAccCoins(client1Addr)

	// compile and execute create swap script
	{
		// supplies before the Tx
		suppliesBefore := GetAllSupplies(t, app, GetContext(app, true))

		// compile
		createSwapScriptSrc := strings.ReplaceAll(createSwapScriptSrcFmt, "{{sender}}", client1Addr.String())
		byteCode, compileErr := vm_client.Compile(dvmAddr, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: createSwapScriptSrc,
					Name: "createSwapScriptSrc",
				},
			},
			Address: client1LibraAddr,
		})
		require.NoError(t, compileErr)
		require.Len(t, byteCode, 1)

		// prepare execute Tx
		swapAmountArg, amountArgErr := vm_client.NewU128ScriptArg(offerAmount.String())
		require.NoError(t, amountArgErr)
		swapPriceArg, priceArgErr := vm_client.NewU128ScriptArg(priceAmount.String())
		require.NoError(t, priceArgErr)

		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), client1PrivKey
		executeMsg := vm.MsgExecuteScript{
			Signer: client1Addr,
			Script: byteCode[0].ByteCode,
			Args:   []vm.ScriptArg{swapAmountArg, swapPriceArg},
		}
		tx := GenTx([]sdk.Msg{executeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)

		// execute script manually (without triggering EndBLocker as crisis module would panic)
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		// supplies after the Tx (within current DeliverTx context)
		suppliesDiff := suppliesBefore.GetDiffString(GetAllSupplies(t, app, GetContext(app, false)))
		t.Logf(">> Swap create, supply diff:\n%s", strings.Join(verboseSuppliesDiff(suppliesDiff), "\n"))

		// crisis module should panic here if its invariant is enalbed
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		// check crisis invariants
		skipBlock()
	}

	// check client1 balance after Swap lock
	{
		client1XfiAfterLock, _ := getXfiBtcAccCoins(client1Addr)
		t.Logf("client1 before/after Swap lock: %s / %s", client1XfiBeforeLock, client1XfiAfterLock)

		// calc expected amount including fee
		expectedAmount := client1XfiBeforeLock.Amount
		expectedAmount = expectedAmount.Sub(offerAmount)
		expectedAmount = expectedAmount.Sub(sdk.OneInt())
		require.True(t, client1XfiAfterLock.Amount.Equal(expectedAmount))
	}

	// save client balances before swap execution
	client1XfiBeforeExecution, client1BtcBeforeExecution := getXfiBtcAccCoins(client1Addr)
	client2XfiBeforeExecution, client2BtcBeforeExecution := getXfiBtcAccCoins(client2Addr)

	// compile and execute swap execute script
	{
		suppliesBefore := GetAllSupplies(t, app, GetContext(app, true))

		createSwapScriptSrc := strings.ReplaceAll(swapSwapScriptSrcFmt, "{{sender}}", client1Addr.String())
		byteCode, compileErr := vm_client.Compile(dvmAddr, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: createSwapScriptSrc,
					Name: "createSwapScriptSrc",
				},
			},
			Address: client1LibraAddr,
		})
		require.NoError(t, compileErr)
		require.Len(t, byteCode, 1)

		sellerAddrArg, sellerArgErr := vm_client.NewAddressScriptArg(client1Addr.String())
		require.NoError(t, sellerArgErr)
		swapPriceArg, priceArgErr := vm_client.NewU128ScriptArg(priceAmount.String())
		require.NoError(t, priceArgErr)

		senderAcc, senderPrivKey := GetAccountCheckTx(app, client2Addr), client2PrivKey
		executeMsg := vm.MsgExecuteScript{
			Signer: client2Addr,
			Script: byteCode[0].ByteCode,
			Args:   []vm.ScriptArg{sellerAddrArg, swapPriceArg},
		}
		tx := GenTx([]sdk.Msg{executeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		suppliesDiff := suppliesBefore.GetDiffString(GetAllSupplies(t, app, GetContext(app, true)))
		t.Logf(">> Swap execute, supply diff:\n%s", strings.Join(verboseSuppliesDiff(suppliesDiff), "\n"))

		// check crisis invariants
		skipBlock()
	}

	// check balances after Swap execution
	{
		client1XfiAfterExecution, client1BtcAfterExecution := getXfiBtcAccCoins(client1Addr)
		client2XfiAfterExecution, client2BtcAfterExecution := getXfiBtcAccCoins(client2Addr)

		t.Logf("client1 before/after Swap execution: %s / %s, %s / %s", client1XfiBeforeExecution, client1XfiAfterExecution, client1BtcBeforeExecution, client1BtcAfterExecution)
		t.Logf("client2 before/after Swap execution: %s / %s, %s / %s", client2XfiBeforeExecution, client2XfiAfterExecution, client2BtcBeforeExecution, client2BtcAfterExecution)

		// client1
		{
			// xfi
			require.True(t, client1XfiAfterExecution.IsEqual(client1XfiBeforeExecution))
			// btc
			expectedBtcAmount := client1BtcBeforeExecution.Amount
			expectedBtcAmount = expectedBtcAmount.Add(priceAmount)
			require.True(t, client1BtcAfterExecution.Amount.Equal(expectedBtcAmount))
		}
		// client2
		{
			// xfi (including fee)
			expectedXfiAmount := client2XfiBeforeExecution.Amount
			expectedXfiAmount = expectedXfiAmount.Add(offerAmount)
			expectedXfiAmount = expectedXfiAmount.Sub(sdk.OneInt())
			require.True(t, client2XfiAfterExecution.Amount.Equal(expectedXfiAmount))
			// btc
			expectedBtcAmount := client2BtcBeforeExecution.Amount
			expectedBtcAmount = expectedBtcAmount.Sub(priceAmount)
			require.True(t, client2BtcAfterExecution.Amount.Equal(expectedBtcAmount))
		}
	}
}
