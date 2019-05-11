package story

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// KeyExpireDuration is store's key for expire duration
var (
	KeyExpireDuration = []byte("expireDuration")
	KeyMinStoryLength = []byte("minStoryLength")
	KeyMaxStoryLength = []byte("maxStoryLength")
)

// Params holds parameters for a story
type Params struct {
	ExpireDuration time.Duration `json:"expire_duration"`
	MinStoryLength int           `json:"min_story_length"`
	MaxStoryLength int           `json:"max_story_length"`
}

// DefaultParams is the story params for testing
func DefaultParams() Params {
	return Params{
		ExpireDuration: 24 * time.Hour,
		MinStoryLength: 25,
		MaxStoryLength: 350,
	}
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyExpireDuration, Value: &p.ExpireDuration},
		{Key: KeyMinStoryLength, Value: &p.MinStoryLength},
		{Key: KeyMaxStoryLength, Value: &p.MaxStoryLength},
	}
}

// ParamKeyTable for story module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// GetParams gets the genesis params for the story
func (k Keeper) GetParams(ctx sdk.Context) Params {
	var paramSet Params
	k.paramStore.GetParamSet(ctx, &paramSet)
	return paramSet
}

// SetParams sets the params for the story
func (k Keeper) SetParams(ctx sdk.Context, params Params) {
	logger := ctx.Logger().With("module", "story")
	k.paramStore.SetParamSet(ctx, &params)
	logger.Info(fmt.Sprintf("Loaded story params: %+v", params))
}

func (k Keeper) minStoryLength(ctx sdk.Context) (res int) {
	k.paramStore.Get(ctx, KeyMinStoryLength, &res)
	return
}

func (k Keeper) expireDuration(ctx sdk.Context) (res time.Duration) {
	k.paramStore.Get(ctx, KeyExpireDuration, &res)
	return
}
