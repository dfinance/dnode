// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPOA_Genesis_Validate(t *testing.T) {
	t.Parallel()

	// ok
	{
		state := GenesisState{
			Parameters: Params{
				MaxValidators: DefaultMaxValidators,
				MinValidators: DefaultMinValidators,
			},
			Validators: Validators{
				NewValidator(sdk.AccAddress("addr1"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b71"),
				NewValidator(sdk.AccAddress("addr2"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b72"),
				NewValidator(sdk.AccAddress("addr3"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b73"),
			},
		}
		require.NoError(t, state.Validate(false))
	}

	// fail: invalid params
	{
		state := GenesisState{
			Parameters: Params{
				MaxValidators: DefaultMaxValidators + 1,
				MinValidators: DefaultMinValidators,
			},
			Validators: Validators{
				NewValidator(sdk.AccAddress("addr1"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b71"),
				NewValidator(sdk.AccAddress("addr2"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b72"),
				NewValidator(sdk.AccAddress("addr3"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b73"),
			},
		}
		require.Error(t, state.Validate(false))
	}

	// fail: invalid validator
	{
		state := GenesisState{
			Parameters: Params{
				MaxValidators: DefaultMaxValidators,
				MinValidators: DefaultMinValidators,
			},
			Validators: Validators{
				NewValidator(sdk.AccAddress("addr1"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7"),
				NewValidator(sdk.AccAddress("addr2"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b72"),
				NewValidator(sdk.AccAddress("addr3"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b73"),
			},
		}
		require.Error(t, state.Validate(false))
	}

	// fail: validator count < min
	{
		state := GenesisState{
			Parameters: Params{
				MaxValidators: DefaultMaxValidators,
				MinValidators: DefaultMinValidators,
			},
			Validators: Validators{},
		}
		require.Error(t, state.Validate(false))
		require.NoError(t, state.Validate(true))
	}

	// fail: validator count > max
	{
		state := GenesisState{
			Parameters: Params{
				MaxValidators: DefaultMaxValidators,
				MinValidators: DefaultMinValidators,
			},
			Validators: Validators{
				NewValidator(sdk.AccAddress("addr1"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b71"),
				NewValidator(sdk.AccAddress("addr2"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b72"),
				NewValidator(sdk.AccAddress("addr3"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b73"),
				NewValidator(sdk.AccAddress("addr4"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b74"),
				NewValidator(sdk.AccAddress("addr5"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b75"),
				NewValidator(sdk.AccAddress("addr6"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b76"),
				NewValidator(sdk.AccAddress("addr7"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b77"),
				NewValidator(sdk.AccAddress("addr8"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b78"),
				NewValidator(sdk.AccAddress("addr9"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b79"),
				NewValidator(sdk.AccAddress("addr10"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7A"),
				NewValidator(sdk.AccAddress("addr11"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7B"),
				NewValidator(sdk.AccAddress("addr12"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7C"),
			},
		}
		require.Error(t, state.Validate(false))
		require.NoError(t, state.Validate(true))
	}

	// fail: duplicate validators
	{
		state := GenesisState{
			Parameters: Params{
				MaxValidators: DefaultMaxValidators,
				MinValidators: DefaultMinValidators,
			},
			Validators: Validators{
				NewValidator(sdk.AccAddress("addr1"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b71"),
				NewValidator(sdk.AccAddress("addr2"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b72"),
				NewValidator(sdk.AccAddress("addr1"), "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b73"),
			},
		}
		require.Error(t, state.Validate(false))
	}
}
