// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVM_StdlibUpdateProposal(t *testing.T) {
	// ok
	require.NoError(t, NewStdlibUpdateProposal(NewPlan(1), "http://github.com/repo", "tst", []byte{1}).ValidateBasic())

	// check plan validation
	require.Error(t, NewStdlibUpdateProposal(NewPlan(0), "http://github.com/repo", "tst", []byte{1}).ValidateBasic())

	// check parameters validation
	require.Error(t, NewStdlibUpdateProposal(NewPlan(1), "", "tst", []byte{1}).ValidateBasic())
	require.Error(t, NewStdlibUpdateProposal(NewPlan(1), "1://repo", "tst", []byte{1}).ValidateBasic())
	require.Error(t, NewStdlibUpdateProposal(NewPlan(1), "http://github.com/repo", "", []byte{1}).ValidateBasic())
	require.Error(t, NewStdlibUpdateProposal(NewPlan(1), "http://github.com/repo", "tst", nil).ValidateBasic())
}
