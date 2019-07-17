package types

import (
    "github.com/cosmos/cosmos-sdk/types"
    "fmt"
    "crypto/sha256"
    "encoding/hex"
)

type Destroy struct {
    ID      types.Int        `json:"id"`
    ChainID string           `json:"chainID"`
    Symbol  string           `json:"symbol"`
    Amount  types.Int        `json:"amount"`
    Spender types.AccAddress `json:"spender"`
    TxHash  string           `json:"tx_hash"`
}

func NewDestroy(id types.Int, chainID string, symbol string, amount types.Int, spender types.AccAddress, txBytes []byte) Destroy {
    hash := sha256.Sum256(txBytes)

    return Destroy{
        ID:      id,
        ChainID: chainID,
        Symbol:  symbol,
        Amount:  amount,
        Spender: spender,
        TxHash:  hex.EncodeToString(hash[:]),
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

