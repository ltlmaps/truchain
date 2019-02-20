package trubank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Category errors reserve 1200 ~ 1299.
const (
	DefaultCodespace sdk.CodespaceType = "tbank"

	CodeErrorAddingCoinsToUser     sdk.CodeType = 1201
	CodeErrorAddingCoinsToCategory sdk.CodeType = 1202
)

// ErrTransferringCoinsToUser throws an error when the category is invalid
func ErrTransferringCoinsToUser(creator sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrorAddingCoinsToUser, "Coins could not be added to the user "+fmt.Sprintf("%s", creator))
}

// ErrTransferringCoinsToCategory throws an error when a category msg is invalid
func ErrTransferringCoinsToCategory(id int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrorAddingCoinsToCategory, "Coins could not be added to category "+fmt.Sprintf("%d", id))
}
