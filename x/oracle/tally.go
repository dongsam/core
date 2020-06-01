package oracle

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-project/core/x/oracle/internal/types"
)

// Calculates the median and returns it. Sets the set of voters to be rewarded, i.e. voted within
// a reasonable spread from the weighted median to the store
func tally(ctx sdk.Context, pb types.ExchangeRateBallot, rewardBand sdk.Dec) (weightedMedian sdk.Dec, ballotWinners []types.Claim) {
	if !sort.IsSorted(pb) {
		sort.Sort(pb)
	}
	// TODO: fix
	weightedMedian = pb.WeightedMedian()
	standardDeviation := pb.StandardDeviation()
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))

	if standardDeviation.GT(rewardSpread) {
		rewardSpread = standardDeviation
	}

	for _, vote := range pb {
		// Filter ballot winners & abstain voters
		if (vote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
			vote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))) ||
			!vote.ExchangeRate.IsPositive() {

			// Abstain votes have zero vote power
			ballotWinners = append(ballotWinners, types.Claim{
				Recipient: vote.Voter,
				Weight:    vote.Power,
			})
		}

	}

	return
}

// Calculates the median and returns it. Sets the set of voters to be rewarded, i.e. voted within
// a reasonable spread from the weighted median to the store
func tallyCrossRate(ctx sdk.Context, voteMap map[string]types.ExchangeRateBallot, voteTargets map[string]sdk.Dec) types.CrossExchangeRates {
	crossRateMapByVali := map[string]map[string]map[string]sdk.Dec{}

	for denom, ballot := range voteMap {
		if _, exists := voteTargets[denom]; !exists {
			continue
		}
		for _, vote := range ballot {
			// Filter ballot winners & abstain voters
			if vote.ExchangeRate.IsPositive() {
				for target, voteTargets
				crossRateMapByVali[vote.Voter.String()][denom][denom] = vote.ExchangeRate
			}
		}
	}
	// If denom is not in the voteTargets, or the ballot for it has failed, then skip

	//if !sort.IsSorted(pb) {
	//	sort.Sort(pb)
	//}
	//weightedMedian = pb.WeightedMedian()
	//standardDeviation := pb.StandardDeviation()
	//rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))
	//
	//if standardDeviation.GT(rewardSpread) {
	//	rewardSpread = standardDeviation
	//}
	//
	//for _, vote := range pb {
	//	// Filter ballot winners & abstain voters
	//	if (vote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
	//		vote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))) ||
	//		!vote.ExchangeRate.IsPositive() {
	//
	//		// Abstain votes have zero vote power
	//		ballotWinners = append(ballotWinners, types.Claim{
	//			Recipient: vote.Voter,
	//			Weight:    vote.Power,
	//		})
	//	}
	//
	//}

	return
}


// ballot for the asset is passing the threshold amount of voting power
func ballotIsPassing(ctx sdk.Context, ballot types.ExchangeRateBallot, k Keeper) bool {
	totalBondedPower := sdk.TokensToConsensusPower(k.StakingKeeper.TotalBondedTokens(ctx))
	voteThreshold := k.VoteThreshold(ctx)
	thresholdVotes := voteThreshold.MulInt64(totalBondedPower).RoundInt()
	ballotPower := sdk.NewInt(ballot.Power())
	return ballotPower.GTE(thresholdVotes)
}
