package types

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VoteForTally is a convinience wrapper to reduct redundant lookup cost
type VoteForTally struct {
	ExchangeRateVote
	Power int64
}

// NewVoteForTally returns a new VoteForTally instance
func NewVoteForTally(vote ExchangeRateVote, power int64) VoteForTally {
	return VoteForTally{
		vote,
		power,
	}
}

type CrossExchangeRates []CrossExchangeRate

type CrossExchangeRate struct {
	Denom1            string  `json:"denom1"`              // Ticker name of target first terra currency
	Denom2            string  `json:"denom2"`              // Ticker name of target second terra currency
	CrossExchangeRate sdk.Dec `json:"cross_exchange_rate"` // CrossExchangeRate of Luna in target fiat currency
}

func GetDenomOrderAsc(denom1, denom2 string) (string, string) {
	if denom1 > denom2 {
		return denom2, denom1
	}
	return denom1, denom2
}

func NewCrossExchangeRate(denom1, denom2 string, exchangeRate sdk.Dec) CrossExchangeRate {
	// swap ascending order for deterministic kv indexing
	denom1, denom2 = GetDenomOrderAsc(denom1, denom2)
	//if denom1 > denom2 {
	//	denom1, denom2 = denom2, denom1
	//}
	return CrossExchangeRate{
		denom1,
		denom2,
		exchangeRate,
	}
}
// Todo ordering CrossExchangeRates on get

func (cer CrossExchangeRate) DenomPair() string {
	return cer.Denom1 + "_" + cer.Denom2
}

// ExchangeRateBallot is a convinience wrapper around a ExchangeRateVote slice
type ExchangeRateBallot []VoteForTally

// Power returns the total amount of voting power in the ballot
func (pb ExchangeRateBallot) Power() int64 {
	totalPower := int64(0)
	for _, vote := range pb {
		totalPower += vote.Power
	}

	return totalPower
}

// WeightedMedian returns the median weighted by the power of the ExchangeRateVote.
func (pb ExchangeRateBallot) WeightedMedian() sdk.Dec {
	totalPower := pb.Power()
	if pb.Len() > 0 {
		if !sort.IsSorted(pb) {
			sort.Sort(pb)
		}

		pivot := int64(0)
		for _, v := range pb {
			votePower := v.Power

			pivot += votePower
			if pivot >= (totalPower / 2) {
				return v.ExchangeRate
			}
		}
	}
	return sdk.ZeroDec()
}

// StandardDeviation returns the standard deviation by the power of the ExchangeRateVote.
func (pb ExchangeRateBallot) StandardDeviation() (standardDeviation sdk.Dec) {
	if len(pb) == 0 {
		return sdk.ZeroDec()
	}

	median := pb.WeightedMedian()

	sum := sdk.ZeroDec()
	for _, v := range pb {
		deviation := v.ExchangeRate.Sub(median)
		sum = sum.Add(deviation.Mul(deviation))
	}

	variance := sum.QuoInt64(int64(len(pb)))

	floatNum, _ := strconv.ParseFloat(variance.String(), 64)
	floatNum = math.Sqrt(floatNum)
	standardDeviation, _ = sdk.NewDecFromStr(fmt.Sprintf("%f", floatNum))

	return
}

// Len implements sort.Interface
func (pb ExchangeRateBallot) Len() int {
	return len(pb)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (pb ExchangeRateBallot) Less(i, j int) bool {
	return pb[i].ExchangeRate.LTE(pb[j].ExchangeRate)
}

// Swap implements sort.Interface.
func (pb ExchangeRateBallot) Swap(i, j int) {
	pb[i], pb[j] = pb[j], pb[i]
}

// Calculates the median and returns it. Sets the set of voters to be rewarded, i.e. voted within
// a reasonable spread from the weighted median to the store
func Tally(pb ExchangeRateBallot, rewardBand sdk.Dec) (weightedMedian sdk.Dec, ballotWinners []Claim) {
	if !sort.IsSorted(pb) {
		sort.Sort(pb)
	}

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
			ballotWinners = append(ballotWinners, Claim{
				Recipient: vote.Voter,
				Weight:    vote.Power,
			})
		}

	}

	return
}

// Calculates the median for cross exchange rate and returns it.
func TallyCrossRate(voteMap map[string]ExchangeRateBallot, voteTargets map[string]sdk.Dec) (crossExchangeRates CrossExchangeRates) {
	crossRateMapByVali := make(map[string]map[string]sdk.Dec)
	crossRateMapByDenom := make(map[string]map[string]ExchangeRateBallot)

	for denom, ballot := range voteMap {
		if _, exists := voteTargets[denom]; !exists {
			continue
		}
		for _, vote := range ballot {
			// Filter ballot winners & abstain voters
			if vote.ExchangeRate.IsPositive() {
				for k, v := range crossRateMapByVali[vote.Voter.String()]{
					var crossRate sdk.Dec
					denom1, denom2 := GetDenomOrderAsc(denom, k)
					fmt.Println(v, vote.ExchangeRate)
					if denom > k {
						crossRate = v.Quo(vote.ExchangeRate) // TODO: need to check precision
					} else {
						crossRate = vote.ExchangeRate.Quo(v) // TODO: need to check precision
					}
					fmt.Println(v, vote.ExchangeRate, crossRate)
					if crossRate.IsPositive() {
						if _, ok := crossRateMapByDenom[denom1]; !ok {
							crossRateMapByDenom[denom1] = make(map[string]ExchangeRateBallot)
						}
						//crossRateMapByVali[vote.Voter.String()][denom1] = true
						crossRateMapByDenom[denom1][denom2] = append(crossRateMapByDenom[denom1][denom2],
							NewVoteForTally(
								NewExchangeRateVote(crossRate, denom1+"_"+denom2, vote.Voter),
								vote.Power,
							),
						)
					}
				}
				if _, ok := crossRateMapByVali[vote.Voter.String()]; !ok {
					crossRateMapByVali[vote.Voter.String()] = make(map[string]sdk.Dec)
				}
				crossRateMapByVali[vote.Voter.String()][denom] = vote.ExchangeRate  // for only existing flag tmp
			}
		}
	}

	for denom1, denom2List := range crossRateMapByDenom {
		for denom2, erb := range denom2List {
			wm := erb.WeightedMedian()
			// TODO: validation
			if wm.IsPositive(){
				crossExchangeRates = append(crossExchangeRates, NewCrossExchangeRate(denom1, denom2, wm))
			}
		}
	}
	fmt.Println(crossRateMapByDenom, crossExchangeRates)
	return
}