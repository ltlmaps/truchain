package challenge

import (
	store "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewResponseEndBlock is called at the end of every block tick
func (k Keeper) NewResponseEndBlock(ctx sdk.Context) sdk.Tags {
	gameQueue := store.NewQueue(k.GetCodec(), ctx.KVStore(k.gameQueueKey))
	err := checkExpiredGames(ctx, k, gameQueue)
	if err != nil {
		panic(err)
	}

	return sdk.EmptyTags()
}

// ============================================================================

// checkExpiredGames checks each validation game to see if it has expired.
// It calls itself recursively until all games have been processed.
func checkExpiredGames(ctx sdk.Context, k Keeper, q store.Queue) sdk.Error {
	// check the head of the queue
	var gameID int64
	if err := q.Peek(&gameID); err != nil {
		return nil
	}

	// retrieve game from kvstore
	game, err := k.gameKeeper.Get(ctx, gameID)
	if err != nil {
		return err
	}

	// all remaining games expire at a later time
	if game.ExpiresTime.After(ctx.BlockHeader().Time) {
		// terminate recursion
		return nil
	}

	// remove expired game from queue
	q.Pop()

	// return funds if game hasn't started
	if !game.Started() {
		if err = returnFunds(ctx, k, game.ID); err != nil {
			return err
		}

		// update story state to reflect expired game
		err = k.storyKeeper.EndChallenge(ctx, game.StoryID)
		if err != nil {
			return err
		}
	}

	return checkExpiredGames(ctx, k, q)
}

// returnFunds iterates through the challenge keyspace and returns funds
func returnFunds(ctx sdk.Context, k Keeper, gameID int64) sdk.Error {
	store := k.GetStore(ctx)

	// builds prefix of from "game:id:5:challenges:user:"
	prefix := k.challengeByGameIDSubspace(ctx, gameID)

	// iterates through keyspace to find all challenges on a game
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var challengeID int64
		k.GetCodec().MustUnmarshalBinary(iter.Value(), &challengeID)
		challenge, err := k.Get(ctx, challengeID)
		if err != nil {
			return err
		}

		// return funds
		_, _, err = k.bankKeeper.AddCoins(
			ctx, challenge.Creator, sdk.Coins{challenge.Amount})
		if err != nil {
			return err
		}
	}

	return nil
}
