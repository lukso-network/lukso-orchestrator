// Package filters specifies utilities for building a set of data attribute
// filters to be used when filtering data through database queries in practice.
// For example, one can specify a filter query for data by start epoch + end epoch + shard
// for attestations, build a filter as follows, and respond to it accordingly:
//
//   f := filters.NewFilter().SetStartEpoch(3).SetEndEpoch(5)
//   for k, v := range f.Filters() {
//       switch k {
//       case filters.StartEpoch:
//          // Verify data matches filter criteria...
//       case filters.EndEpoch:
//          // Verify data matches filter criteria...
//       }
//   }
package filters

import types "github.com/prysmaticlabs/eth2-types"

// FilterType defines an enum which is used as the keys in a map that tracks
// set attribute filters for data as part of the `FilterQuery` struct type.
type FilterType uint8

const (
	// StartEpoch is used for range filters of objects by their epoch (inclusive).
	StartEpoch FilterType = iota
	// EndEpoch is used for range filters of objects by their epoch (inclusive).
	EndEpoch
)

// QueryFilter defines a generic interface for type-asserting
// specific filters to use in querying DB objects.
type QueryFilter struct {
	queries map[FilterType]interface{}
}

// NewFilter instantiates a new QueryFilter type used to build filters for
// certain eth2 data types by attribute.
func NewFilter() *QueryFilter {
	return &QueryFilter{
		queries: make(map[FilterType]interface{}),
	}
}

// Filters returns and underlying map of FilterType to interface{}, giving us
// a copy of the currently set filters which can then be iterated over and type
// asserted for use anywhere.
func (q *QueryFilter) Filters() map[FilterType]interface{} {
	return q.queries
}

// SetStartEpoch enables filtering by the StartEpoch attribute of an object (inclusive).
func (q *QueryFilter) SetStartEpoch(val types.Epoch) *QueryFilter {
	q.queries[StartEpoch] = val
	return q
}

// SetEndEpoch enables filtering by the EndEpoch attribute of an object (inclusive).
func (q *QueryFilter) SetEndEpoch(val types.Epoch) *QueryFilter {
	q.queries[EndEpoch] = val
	return q
}
