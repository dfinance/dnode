package types

import (
    "github.com/cosmos/cosmos-sdk/types"
    "fmt"
)

type Destroy struct {
    ID      types.Int        `json:"id"`
    ChainID string           `json:"chainID"`
    Symbol  string           `json:"symbol"`
    Amount  types.Int        `json:"amount"`
    Spender types.AccAddress `json:"spender"`
}

func NewDestroy(id types.Int, chainID string, symbol string, amount types.Int, spender types.AccAddress) Destroy {
    return Destroy{
        ID:      id,
        ChainID: chainID,
        Symbol:  symbol,
        Amount:  amount,
        Spender: spender,
    }
}

func (destroy Destroy) String() string {
    return fmt.Sprintf("Destroy: \n" +
        "\tChainID: %s\n" +
        "\tID:      %s\n" +
        "\tSymbol:  %s\n" +
        "\tAmount:  %s\n" +
        "\tSpender: %s\n",
        destroy.ChainID,
        destroy.ID, destroy.Symbol,
        destroy.Amount, destroy.Spender)
}

