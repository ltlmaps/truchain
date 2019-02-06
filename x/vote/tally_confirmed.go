package vote

import (
	"fmt"

	app "github.com/TruStory/truchain/types"
	"github.com/TruStory/truchain/x/backing"
	"github.com/TruStory/truchain/x/challenge"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// calculate reward pool for a confirmed story
func confirmedPool(
	ctx sdk.Context, falseVotes []app.Voter, pool *sdk.Coin) (err sdk.Error) {

	for _, vote := range falseVotes {
		switch v := vote.(type) {

		case backing.Backing:
			// slash inflationary rewards and add to pool
			*pool = (*pool).Plus(v.Interest)

		case challenge.Challenge:
			// add challenge amount to reward pool
			*pool = (*pool).Plus(v.Amount())

		case TokenVote:
			// add vote fee to reward pool
			*pool = (*pool).Plus(v.Amount())

		default:
			if err = ErrInvalidVote(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func distributeRewardsConfirmed(
	ctx sdk.Context,
	bankKeeper bank.Keeper,
	votes poll,
	pool sdk.Coin,
	denom string) (err sdk.Error) {

	logger := ctx.Logger().With("module", "vote")
	logger.Info(fmt.Sprintf("reward pool (confirmed): %s", pool))

	// determine pool share per voter
	voterRewardAmount := voterRewardAmount(pool, voterCount(votes.trueVotes))
	logger.Info(fmt.Sprintf("token voter reward amount: %s", voterRewardAmount))

	// distribute reward to winners
	for _, vote := range votes.trueVotes {
		switch v := vote.(type) {

		case backing.Backing:
			// keep backing as is

		case TokenVote:
			// get back original staked amount in trustake
			_, _, err = bankKeeper.AddCoins(ctx, v.Creator(), sdk.Coins{v.Amount()})
			if err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("giving back origin amount: %s", v.Amount()))

			// calculate reward, an equal portion of the reward pool
			rewardCoin := sdk.NewCoin(pool.Denom, voterRewardAmount)

			// remove reward amount from pool
			pool = pool.Minus(rewardCoin)

			// distribute reward in cred
			cred := app.NewCategoryCoin(denom, rewardCoin)
			_, _, err = bankKeeper.AddCoins(ctx, v.Creator(), sdk.Coins{cred})

		default:
			if err = ErrInvalidVote(v); err != nil {
				return err
			}
		}
	}

	// slash losers
	for _, vote := range votes.falseVotes {
		switch v := vote.(type) {

		// backer who changed their implicit TRUE vote to FALSE
		case backing.Backing:
			// return backing because we are nice people
			_, _, err = bankKeeper.AddCoins(ctx, v.Creator(), sdk.Coins{v.Amount()})

		case challenge.Challenge:
			// do nothing
			// don't get their stake back

		case TokenVote:
			// do nothing
			// don't get their stake back

		default:
			err = ErrInvalidVote(v)
		}

		if err != nil {
			return err
		}
	}

	// TODO [shanev]: Remove after fixing https://github.com/TruStory/truchain/issues/314
	// err = checkForEmptyPool(pool)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// count voters
func voterCount(winners []app.Voter) (voterCount int64) {
	for _, voter := range winners {
		if _, ok := voter.(TokenVote); ok {
			voterCount = voterCount + 1
		}
	}

	return voterCount
}

func voterRewardAmount(pool sdk.Coin, voterCount int64) sdk.Int {
	// check for no token voters
	// prevent division by zero errors
	if voterCount == 0 {
		return sdk.NewInt(0)
	}

	poolDec := sdk.NewDecFromInt(pool.Amount)
	voterCountInt := sdk.NewInt(voterCount)

	return poolDec.
		QuoInt(voterCountInt).
		RoundInt()
}
