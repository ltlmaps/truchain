package story

import (
	app "github.com/TruStory/truchain/types"
	"github.com/TruStory/truchain/x/category"
	sdk "github.com/cosmos/cosmos-sdk/types"
	amino "github.com/tendermint/go-amino"
)

// ReadKeeper defines a module interface that facilitates read only access
// to truchain data
type ReadKeeper interface {
	app.ReadKeeper

	GetChallengedStoriesWithCategory(
		ctx sdk.Context,
		catID int64) (stories []Story, err sdk.Error)
	GetCoinName(ctx sdk.Context, id int64) (name string, err sdk.Error)
	GetFeedByCategory(
		ctx sdk.Context,
		catID int64) (stories []Story, err sdk.Error)
	GetStoriesByCategory(ctx sdk.Context, catID int64) (stories []Story, err sdk.Error)
	GetStory(ctx sdk.Context, storyID int64) (Story, sdk.Error)
}

// WriteKeeper defines a module interface that facilities read/write access
type WriteKeeper interface {
	ReadKeeper

	NewStory(
		ctx sdk.Context, body string, categoryID int64, creator sdk.AccAddress, kind Kind) (int64, sdk.Error)
	StartChallenge(ctx sdk.Context, storyID int64) sdk.Error
	UpdateStory(ctx sdk.Context, story Story)
}

// Keeper data type storing keys to the key-value store
type Keeper struct {
	app.Keeper

	categoryKeeper category.ReadKeeper
}

// NewKeeper creates a new keeper with write and read access
func NewKeeper(
	storeKey sdk.StoreKey,
	categoryKeeper category.ReadKeeper,
	codec *amino.Codec) Keeper {

	return Keeper{app.NewKeeper(codec, storeKey), categoryKeeper}
}

// ============================================================================

// StartChallenge records challenging a story
func (k Keeper) StartChallenge(ctx sdk.Context, storyID int64) sdk.Error {
	// get story
	story, err := k.GetStory(ctx, storyID)
	if err != nil {
		return err
	}
	// update story state
	story.State = Challenged
	k.UpdateStory(ctx, story)

	// add story to challenged list
	k.appendStoriesList(
		ctx, storyIDsByCategoryKey(k, story.CategoryID, story.Timestamp, true), story)

	return nil
}

// NewStory adds a story to the key-value store
func (k Keeper) NewStory(
	ctx sdk.Context,
	body string,
	categoryID int64,
	creator sdk.AccAddress,
	kind Kind) (int64, sdk.Error) {

	_, err := k.categoryKeeper.GetCategory(ctx, categoryID)
	if err != nil {
		return 0, category.ErrInvalidCategory(categoryID)
	}

	story := Story{
		ID:         k.GetNextID(ctx),
		Body:       body,
		CategoryID: categoryID,
		Creator:    creator,
		GameID:     0,
		State:      Created,
		Kind:       kind,
		Timestamp:  app.NewTimestamp(ctx.BlockHeader()),
	}

	k.setStory(ctx, story)
	k.appendStoriesList(
		ctx, storyIDsByCategoryKey(k, categoryID, story.Timestamp, false), story)

	return story.ID, nil
}

// GetCoinName returns the name of the category coin for the story
func (k Keeper) GetCoinName(ctx sdk.Context, id int64) (name string, err sdk.Error) {
	story, err := k.GetStory(ctx, id)
	if err != nil {
		return
	}
	cat, err := k.categoryKeeper.GetCategory(ctx, story.CategoryID)
	if err != nil {
		return
	}

	return cat.CoinName(), nil
}

// GetStory gets the story with the given id from the key-value store
func (k Keeper) GetStory(
	ctx sdk.Context, storyID int64) (story Story, err sdk.Error) {

	store := k.GetStore(ctx)
	val := store.Get(k.GetIDKey(storyID))
	if val == nil {
		return story, ErrStoryNotFound(storyID)
	}
	k.GetCodec().MustUnmarshalBinary(val, &story)

	return
}

// GetStoriesByCategory gets the stories for a given category id
func (k Keeper) GetStoriesByCategory(
	ctx sdk.Context, catID int64) (stories []Story, err sdk.Error) {

	return k.storiesByCategory(
		ctx, storyIDsByCategorySubspaceKey(k, catID, false), catID)
}

// GetChallengedStoriesWithCategory gets all challenged stories for a category
func (k Keeper) GetChallengedStoriesWithCategory(
	ctx sdk.Context, catID int64) (stories []Story, err sdk.Error) {

	return k.storiesByCategory(
		ctx, storyIDsByCategorySubspaceKey(k, catID, true), catID)
}

// GetFeedByCategory gets stories ordered by challenged stories first
func (k Keeper) GetFeedByCategory(
	ctx sdk.Context,
	catID int64) (stories []Story, err sdk.Error) {

	// get all story ids by category
	storyIDs, err := k.storyIDsByCategory(
		ctx, storyIDsByCategorySubspaceKey(k, catID, false), catID)
	if err != nil {
		return
	}

	// get all challenged story ids by category
	challengedStoryIDs, err := k.storyIDsByCategory(
		ctx, storyIDsByCategorySubspaceKey(k, catID, true), catID)
	if err != nil {
		return
	}

	// make a list of all unchallenged story ids
	var unchallengedStoryIDs []int64
	for _, sid := range storyIDs {
		isMatch := false
		for _, cid := range challengedStoryIDs {
			isMatch = sid == cid
			if isMatch {
				break
			}
		}
		if !isMatch {
			unchallengedStoryIDs = append(unchallengedStoryIDs, sid)
		}
	}

	// concat challenged story ids with unchallenged story ids
	feedIDs := append(challengedStoryIDs, unchallengedStoryIDs...)

	return k.storiesByID(ctx, feedIDs)
}

// UpdateStory updates an existing story in the store
func (k Keeper) UpdateStory(ctx sdk.Context, story Story) {
	newStory := Story{
		story.ID,
		story.Body,
		story.CategoryID,
		story.Creator,
		story.GameID,
		story.State,
		story.Kind,
		story.Timestamp.Update(ctx.BlockHeader()),
	}

	k.setStory(ctx, newStory)
}

// ============================================================================

func (k Keeper) appendStoriesList(
	ctx sdk.Context, key []byte, story Story) {

	// get stories store
	store := k.GetStore(ctx)

	// marshal story id to list
	store.Set(
		key,
		k.GetCodec().MustMarshalBinary(story.ID))
}

// setStory saves a `Story` type to the KVStore
func (k Keeper) setStory(ctx sdk.Context, story Story) {
	store := k.GetStore(ctx)
	store.Set(
		k.GetIDKey(story.ID),
		k.GetCodec().MustMarshalBinary(story))
}

func (k Keeper) storiesByCategory(
	ctx sdk.Context,
	prefix []byte,
	catID int64) (stories []Story, err sdk.Error) {

	storyIDs, err := k.storyIDsByCategory(ctx, prefix, catID)
	if err != nil {
		return
	}

	if len(storyIDs) == 0 {
		return stories, ErrStoriesWithCategoryNotFound(catID)
	}

	return k.storiesByID(ctx, storyIDs)
}

func (k Keeper) storiesByID(
	ctx sdk.Context, storyIDs []int64) (stories []Story, err sdk.Error) {

	for _, storyID := range storyIDs {
		story, err := k.GetStory(ctx, storyID)
		if err != nil {
			return stories, ErrStoryNotFound(storyID)
		}
		stories = append(stories, story)
	}

	return
}

func (k Keeper) storyIDsByCategory(
	ctx sdk.Context, prefix []byte, catID int64) (storyIDs []int64, err sdk.Error) {

	store := k.GetStore(ctx)

	// iterate over subspace, creating a list of stories
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var storyID int64
		k.GetCodec().MustUnmarshalBinary(iter.Value(), &storyID)
		storyIDs = append(storyIDs, storyID)
	}

	return storyIDs, nil
}
