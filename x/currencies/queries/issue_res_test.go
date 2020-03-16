package queries

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/currencies/types"
)

func TestQueryIssueRes_String(t *testing.T) {
	target := QueryIssueRes{Issue: types.Issue{
		Symbol:    "test",
		Amount:    sdk.Int{},
		Recipient: nil,
	}}

	require.Equal(t, target.Issue.String(), target.String())
}
