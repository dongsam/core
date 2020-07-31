package oracle

import (
	"github.com/terra-project/core/x/oracle/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"

	core "github.com/terra-project/core/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k Keeper) {
	params := k.GetParams(ctx)

	// Not yet time for a tally
	if !core.IsPeriodLastBlock(ctx, params.VotePeriod) {
		return
	}

	// Build valid votes counter and winner map over all validators in active set
	validVotesCounterMap := make(map[string]int)
	winnerMap := make(map[string]types.Claim)
	k.StakingKeeper.IterateValidators(ctx, func(_ int64, validator exported.ValidatorI) bool {

		// Exclude not bonded validator or jailed validators from tallying
		if validator.IsBonded() && !validator.IsJailed() {

			// NOTE: we directly stringify byte to string to prevent unnecessary bech32fy works
			valAddr := validator.GetOperator()
			validVotesCounterMap[string(valAddr)] = 0
			winnerMap[string(valAddr)] = types.NewClaim(0, valAddr)
		}

		return false
	})

	// Denom-TobinTax map
	voteTargets := make(map[string]sdk.Dec)
	k.IterateTobinTaxes(ctx, func(denom string, tobinTax sdk.Dec) bool {
		voteTargets[denom] = tobinTax
		return false
	})

	// Clear all exchange rates
	k.IterateLunaExchangeRates(ctx, func(denom string, _ sdk.Dec) (stop bool) {
		k.DeleteLunaExchangeRate(ctx, denom)
		return false
	})

	// Organize votes to ballot by denom
	// NOTE: **Filter out inactive or jailed validators**
	// NOTE: **Make abstain votes to have zero vote power**
	voteMap := k.OrganizeBallotByDenom(ctx)

	// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
	LargestBallotPower := int64(0)
	var referenceTerra string
	// TODO: check ballot len or power?
	for denom, ballot := range voteMap {
		if _, exists := voteTargets[denom]; !exists {
			continue
		}
		ballotPower := ballot.Power()

		if !k.BallotIsPassing(ctx, ballot) {
			delete(voteTargets, denom)
			continue
		}

		if LargestBallotPower > ballotPower {
			referenceTerra = denom
			LargestBallotPower = ballotPower
		} else if LargestBallotPower == ballotPower {
			// TODO: check alphabetical order
			if referenceTerra > denom {
				referenceTerra = denom
			}
		}
	}

	ballot, _ := voteMap[referenceTerra]
	ballotMedian := k.Tally(ballot)
	k.SetLunaExchangeRate(ctx, referenceTerra, ballotMedian)

	// TODO: ballot winning claims for not reference terra after fixed
	// Collect claims of ballot winners
	ballotWinningClaims := k.GetBallotWinners(ballot, params.RewardBand, ballotMedian)
	for _, ballotWinningClaim := range ballotWinningClaims {

		// NOTE: we directly stringify byte to string to prevent unnecessary bech32fy works
		key := string(ballotWinningClaim.Recipient)

		// Update claim
		prevClaim := winnerMap[key]
		prevClaim.Weight += ballotWinningClaim.Weight
		winnerMap[key] = prevClaim

		// Increase valid votes counter
		validVotesCounterMap[key]++
	}

	// Emit abci events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeExchangeRateUpdate,
			sdk.NewAttribute(types.AttributeKeyDenom, referenceTerra),
			sdk.NewAttribute(types.AttributeKeyExchangeRate, ballotMedian.String()),
		),
	)

	crossExchangeRates := k.TallyCrossRate(ctx, voteMap, voteTargets, referenceTerra)
	for _, cer := range crossExchangeRates {
		if cer.Denom1 == referenceTerra {

		} else if cer.Denom2 == referenceTerra {

		} else {
			// TODO: panic
		}
	}
	//---------------------------
	// Do miss counting & slashing
	voteTargetsLen := len(voteTargets)
	for operatorAddrByteStr, count := range validVotesCounterMap {
		// Skip abstain & valid voters
		if count == voteTargetsLen {
			continue
		}

		// Increase miss counter
		operator := sdk.ValAddress(operatorAddrByteStr) // error never occur
		k.SetMissCounter(ctx, operator, k.GetMissCounter(ctx, operator)+1)
	}

	// Do slash who did miss voting over threshold and
	// reset miss counters of all validators at the last block of slash window
	if core.IsPeriodLastBlock(ctx, params.SlashWindow) {
		SlashAndResetMissCounters(ctx, k)
	}

	// Distribute rewards to ballot winners
	k.RewardBallotWinners(ctx, winnerMap)

	// Clear the ballot
	clearBallots(ctx, k, params.VotePeriod)

	// Update vote targets and tobin tax
	applyWhitelist(ctx, k, params.Whitelist, voteTargets)

	return
}

// clearBallots clears all tallied prevotes and votes from the store
func clearBallots(ctx sdk.Context, k Keeper, votePeriod int64) {
	// Clear all prevotes
	k.IterateExchangeRatePrevotes(ctx, func(prevote types.ExchangeRatePrevote) (stop bool) {
		if ctx.BlockHeight() > prevote.SubmitBlock+votePeriod {
			k.DeleteExchangeRatePrevote(ctx, prevote)
		}

		return false
	})

	// Clear all votes
	k.IterateExchangeRateVotes(ctx, func(vote types.ExchangeRateVote) (stop bool) {
		k.DeleteExchangeRateVote(ctx, vote)
		return false
	})

	// Clear all aggregate prevotes
	k.IterateAggregateExchangeRatePrevotes(ctx, func(aggregatePrevote types.AggregateExchangeRatePrevote) (stop bool) {
		if ctx.BlockHeight() > aggregatePrevote.SubmitBlock+votePeriod {
			k.DeleteAggregateExchangeRatePrevote(ctx, aggregatePrevote)
		}

		return false
	})

	// Clear all aggregate votes
	k.IterateAggregateExchangeRateVotes(ctx, func(vote types.AggregateExchangeRateVote) (stop bool) {
		k.DeleteAggregateExchangeRateVote(ctx, vote)
		return false
	})
}

// applyWhitelist update vote target denom list and set tobin tax with params whitelist
func applyWhitelist(ctx sdk.Context, k Keeper, whitelist types.DenomList, voteTargets map[string]sdk.Dec) {

	// check is there any update in whitelist params
	updateRequired := false
	if len(voteTargets) != len(whitelist) {
		updateRequired = true
	} else {
		for _, item := range whitelist {
			if tobinTax, ok := voteTargets[item.Name]; !ok || !tobinTax.Equal(item.TobinTax) {
				updateRequired = true
				break
			}
		}
	}

	if updateRequired {
		k.ClearTobinTaxes(ctx)

		for _, item := range whitelist {
			k.SetTobinTax(ctx, item.Name, item.TobinTax)
		}
	}
}
