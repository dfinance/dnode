// +build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// GenesisIssue validation.
func TestCurrencies_GenesisIssue_Valid(t *testing.T) {
	issue := GenesisIssue{
		Issue: NewIssue(sdk.NewCoin("eth", sdk.ZeroInt()), sdk.AccAddress("addr1")),
		ID:    "",
	}

	// fail: invalid issue
	{
		require.Error(t, issue.Valid())
	}
	// fail: id empty
	{
		issue.Issue = NewIssue(sdk.NewCoin("eth", sdk.OneInt()), sdk.AccAddress("addr1"))
		require.Error(t, issue.Valid())
	}
	// ok
	{
		issue.ID = "id"
		require.NoError(t, issue.Valid())
	}
}

// Genesis validation.
func TestCurrencies_Genesis_Valid(t *testing.T) {
	coin := sdk.NewCoin("eth", sdk.OneInt())
	addr := sdk.AccAddress("addr1")
	pgPayee, pgChainID := "payee", "chainID"
	timestamp, txHash := int64(1), []byte("hash")

	// fail: invalid issue
	{
		state := GenesisState{
			Issues: []GenesisIssue{
				{
					Issue: NewIssue(coin, addr),
					ID:    "",
				},
			},
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: duplicated issues
	{
		state := GenesisState{
			Issues: []GenesisIssue{
				{
					Issue: NewIssue(coin, addr),
					ID:    "1",
				},
				{
					Issue: NewIssue(coin, addr),
					ID:    "2",
				},
				{
					Issue: NewIssue(coin, addr),
					ID:    "1",
				},
			},
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: invalid withdraw
	{
		state := GenesisState{
			Withdraws: Withdraws{Withdraw{}},
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: duplicated withdraws
	{
		state := GenesisState{
			Withdraws: Withdraws{
				NewWithdraw(dnTypes.NewIDFromUint64(1), coin, addr, pgPayee, pgChainID, timestamp, txHash),
				NewWithdraw(dnTypes.NewIDFromUint64(2), coin, addr, pgPayee, pgChainID, timestamp, txHash),
				NewWithdraw(dnTypes.NewIDFromUint64(1), coin, addr, pgPayee, pgChainID, timestamp, txHash),
			},
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: lastWithdrawID invalid
	{
		id := dnTypes.ID{}
		state := GenesisState{
			Withdraws: Withdraws{
				NewWithdraw(dnTypes.NewIDFromUint64(1), coin, addr, pgPayee, pgChainID, timestamp, txHash),
			},
			LastWithdrawID: &id,
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: lastWithdrawID not nil without withdraws
	{
		id := dnTypes.NewZeroID()
		state := GenesisState{
			LastWithdrawID: &id,
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: lastWithdrawID nil with withdraws
	{
		state := GenesisState{
			Withdraws: Withdraws{
				NewWithdraw(dnTypes.NewIDFromUint64(1), coin, addr, pgPayee, pgChainID, timestamp, txHash),
			},
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// fail: lastWithdrawID mismatch
	{
		id := dnTypes.NewIDFromUint64(4)
		state := GenesisState{
			Withdraws: Withdraws{
				NewWithdraw(dnTypes.NewIDFromUint64(1), coin, addr, pgPayee, pgChainID, timestamp, txHash),
				NewWithdraw(dnTypes.NewIDFromUint64(2), coin, addr, pgPayee, pgChainID, timestamp, txHash),
				NewWithdraw(dnTypes.NewIDFromUint64(3), coin, addr, pgPayee, pgChainID, timestamp, txHash),
			},
			LastWithdrawID: &id,
		}
		require.Error(t, state.Validate(time.Time{}))
	}
	// ok
	{
		id := dnTypes.NewIDFromUint64(3)
		state := GenesisState{
			Issues: []GenesisIssue{
				{
					Issue: NewIssue(coin, addr),
					ID:    "1",
				},
				{
					Issue: NewIssue(coin, addr),
					ID:    "2",
				},
			},
			Withdraws: Withdraws{
				NewWithdraw(dnTypes.NewIDFromUint64(1), coin, addr, pgPayee, pgChainID, timestamp, txHash),
				NewWithdraw(dnTypes.NewIDFromUint64(2), coin, addr, pgPayee, pgChainID, timestamp, txHash),
				NewWithdraw(dnTypes.NewIDFromUint64(3), coin, addr, pgPayee, pgChainID, timestamp, txHash),
			},
			LastWithdrawID: &id,
		}
		require.NoError(t, state.Validate(time.Time{}))
	}
}
