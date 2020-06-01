package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-project/core/x/oracle/internal/types"
)

// OrganizeBallotByDenom collects all oracle votes for the period, categorized by the votes' denom parameter
func (k Keeper) OrganizeBallotByDenom(ctx sdk.Context) (votes map[string]types.ExchangeRateBallot) {
	votes = map[string]types.ExchangeRateBallot{}
	aggregateVoterMap := map[string]bool{}

	//params := k.GetParams(ctx)
	//
	//// TODO: make map [vali][denom1][denom2]
	//crossRateMapByVali := map[string]map[string]map[string]sdk.Dec{}

	// Organize aggregate votes
	aggregateHandler := func(vote types.AggregateExchangeRateVote) (stop bool) {
		validator := k.StakingKeeper.Validator(ctx, vote.Voter)

		// organize ballot only for the active validators
		if validator != nil && validator.IsBonded() && !validator.IsJailed() {
			aggregateVoterMap[string(validator.GetOperator().Bytes())] = true

			power := validator.GetConsensusPower()
			for _, tuple := range vote.ExchangeRateTuples {
				tmpPower := power
				if !tuple.ExchangeRate.IsPositive() {
					// Make the power of abstain vote zero
					tmpPower = 0
				}

				//// TODO: WIP
				//// TODO: map exists check, iterate already mapped denom
				//crossRateMapByVali[valoper][tuple.Denom] = ""
				votes[tuple.Denom] = append(votes[tuple.Denom],
					types.NewVoteForTally(
						types.NewExchangeRateVote(tuple.ExchangeRate, tuple.Denom, vote.Voter),
						tmpPower,
					),
				)
			}

		}

		return false
	}
	k.IterateAggregateExchangeRateVotes(ctx, aggregateHandler)

	// organize individual votes
	handler := func(vote types.ExchangeRateVote) (stop bool) {
		validator := k.StakingKeeper.Validator(ctx, vote.Voter)

		// organize ballot only for the active validators
		if validator != nil && validator.IsBonded() && !validator.IsJailed() {
			// block normal vote from the voter who did aggregate vote
			if _, ok := aggregateVoterMap[string(validator.GetOperator().Bytes())]; ok {
				return false
			}

			power := validator.GetConsensusPower()
			if !vote.ExchangeRate.IsPositive() {
				// Make the power of abstain vote zero
				power = 0
			}

			votes[vote.Denom] = append(votes[vote.Denom],
				types.NewVoteForTally(
					vote,
					power,
				),
			)
		}

		return false
	}
	k.IterateExchangeRateVotes(ctx, handler)

	//for valoper, denom1List := range crossRateMapByVali {
	//	for denom1, _ := range denom1List {
	//		delete(denom1List, denom1)
	//		for denom2, _ := range denom1List {
	//			tmp1, tmp2 := types.GetDenomOrderAsc(denom1, denom2)
	//			crossRateMapByVali[valoper][tmp1][tmp2] = tmp2
	//		}
	//		//for _, d := range params.Whitelist {
	//		//	d
	//		//}
	//	}
	//}
	//
	//// TODO: iterate votes, add crossRates
	//
	////if _, ok := crossRateMapByVali[valoper][tuple.Denom]; ok {
	////
	////} else {
	////	crossRateMapByVali[valoper][tuple.Denom] = ""
	////}
	////for _, d := range params.Whitelist {
	////	d
	////}
	return
}
