package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ID sdk.Uint

func (id ID) uint() sdk.Uint {
	return sdk.Uint(id)
}

func (id ID) UInt64() uint64 {
	return id.uint().Uint64()
}

func (id ID) Valid() error {
	if id.uint().BigInt() == nil {
		return fmt.Errorf("nil")
	}

	return nil
}

func (id ID) Equal(id2 ID) bool {
	return id.uint().Equal(id2.uint())
}

func (id ID) LT(id2 ID) bool {
	return id.uint().LT(id2.uint())
}

func (id ID) GT(id2 ID) bool {
	return id.uint().GT(id2.uint())
}

func (id ID) Incr() ID {
	return ID(id.uint().Incr())
}

func (id ID) Decr() ID {
	return ID(id.uint().Decr())
}

func (id ID) String() string {
	return id.uint().String()
}

func (id ID) MarshalAmino() (string, error) {
	return sdk.Uint(id).MarshalAmino()
}

func (id *ID) UnmarshalAmino(text string) error {
	var u sdk.Uint
	err := u.UnmarshalAmino(text)
	if err != nil {
		return err
	}
	*id = ID(u)

	return nil
}

func (id ID) MarshalJSON() ([]byte, error) {
	return sdk.Uint(id).MarshalJSON()
}

func (id *ID) UnmarshalJSON(data []byte) error {
	return (*sdk.Uint)(id).UnmarshalJSON(data)
}

func NewIDFromUint64(id uint64) ID {
	return ID(sdk.NewUint(id))
}

func NewIDFromString(str string) (ID, error) {
	if str == "" {
		return ID{}, fmt.Errorf("empty")
	}

	return ID(sdk.NewUintFromString(str)), nil
}
