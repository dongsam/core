package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Defines the prefix of each query path
const (
	QueryParameters                  = "parameters"
	QueryExchangeRate                = "exchangeRate"
	QueryExchangeRates               = "exchangeRates"
	QueryCrossExchangeRate           = "crossExchangeRate"
	QueryCrossExchangeRates          = "crossExchangeRates"
	QueryCrossExchangeRatesWithDenom = "crossExchangeRatesWithDenom"
	QueryActives                     = "actives"
	QueryPrevotes                    = "prevotes"
	QueryVotes                       = "votes"
	QueryFeederDelegation            = "feederDelegation"
	QueryMissCounter                 = "missCounter"
	QueryAggregatePrevote            = "aggregatePrevote"
	QueryAggregateVote               = "aggregateVote"
	QueryVoteTargets                 = "voteTargets"
	QueryTobinTax                    = "tobinTax"
	QueryTobinTaxes                  = "tobinTaxes"
)

// QueryExchangeRateParams defines the params for the following queries:
// - 'custom/oracle/exchange_rate'
type QueryExchangeRateParams struct {
	Denom string
}

// NewQueryExchangeRateParams returns params for exchange_rate query
func NewQueryExchangeRateParams(denom string) QueryExchangeRateParams {
	return QueryExchangeRateParams{denom}
}

// QueryCrossExchangeRatesWithDenomParams defines the params for the following queries:
// - 'custom/oracle/crossExchangeRatesWithDenom'
type QueryCrossExchangeRatesWithDenomParams struct {
	Denom string
}

// NewQueryCrossExchangeRatesWithDenomParams returns params for cross_exchange_rate query
func NewQueryCrossExchangeRatesWithDenomParams(denom string) QueryCrossExchangeRatesWithDenomParams {
	return QueryCrossExchangeRatesWithDenomParams{denom}
}

// QueryCrossExchangeRateParams defines the params for the following queries:
// - 'custom/oracle/cross_exchange_rate'
type QueryCrossExchangeRateParams struct {
	Denom1 string
	Denom2 string
}

// NewQueryCrossExchangeRateParams returns params for cross_exchange_rate query
func NewQueryCrossExchangeRateParams(denom1, denom2 string) QueryCrossExchangeRateParams {
	return QueryCrossExchangeRateParams{denom1, denom2}
}

// QueryPrevotesParams defines the params for the following queries:
// - 'custom/oracle/prevotes'
type QueryPrevotesParams struct {
	Voter sdk.ValAddress
	Denom string
}

// NewQueryPrevotesParams returns params for exchange_rate prevotes query
func NewQueryPrevotesParams(voter sdk.ValAddress, denom string) QueryPrevotesParams {
	return QueryPrevotesParams{voter, denom}
}

// QueryVotesParams defines the params for the following queries:
// - 'custom/oracle/votes'
type QueryVotesParams struct {
	Voter sdk.ValAddress
	Denom string
}

// NewQueryVotesParams returns params for exchange_rate votes query
func NewQueryVotesParams(voter sdk.ValAddress, denom string) QueryVotesParams {
	return QueryVotesParams{voter, denom}
}

// QueryFeederDelegationParams defeins the params for the following queries:
// - 'custom/oracle/feederDelegation'
type QueryFeederDelegationParams struct {
	Validator sdk.ValAddress
}

// NewQueryFeederDelegationParams returns params for feeder delegation query
func NewQueryFeederDelegationParams(validator sdk.ValAddress) QueryFeederDelegationParams {
	return QueryFeederDelegationParams{validator}
}

// QueryMissCounterParams defines the params for the following queries:
// - 'custom/oracle/missCounter'
type QueryMissCounterParams struct {
	Validator sdk.ValAddress
}

// NewQueryMissCounterParams returns params for feeder delegation query
func NewQueryMissCounterParams(validator sdk.ValAddress) QueryMissCounterParams {
	return QueryMissCounterParams{validator}
}

// QueryAggregatePrevoteParams defines the params for the following queries:
// - 'custom/oracle/aggregatePrevote'
type QueryAggregatePrevoteParams struct {
	Validator sdk.ValAddress
}

// NewQueryAggregatePrevoteParams returns params for feeder delegation query
func NewQueryAggregatePrevoteParams(validator sdk.ValAddress) QueryAggregatePrevoteParams {
	return QueryAggregatePrevoteParams{validator}
}

// QueryAggregateVoteParams defines the params for the following queries:
// - 'custom/oracle/aggregateVote'
type QueryAggregateVoteParams struct {
	Validator sdk.ValAddress
}

// NewQueryAggregateVoteParams returns params for feeder delegation query
func NewQueryAggregateVoteParams(validator sdk.ValAddress) QueryAggregateVoteParams {
	return QueryAggregateVoteParams{validator}
}

// QueryTobinTaxParams defines the params for the following queries:
// - 'custom/oracle/tobinTax'
type QueryTobinTaxParams struct {
	Denom string
}

// NewQueryTobinTaxParams returns params for tobin tax query
func NewQueryTobinTaxParams(denom string) QueryTobinTaxParams {
	return QueryTobinTaxParams{denom}
}
