// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math/rand"
)

// Alias is an interface that allows for sampling from a discrete
// distribution in O(1) time.
type Alias interface {
	Sample(rng *rand.Rand) int
}

// NewAlias creates a new Alias from the given weights.
// The weights must be positive and at least one weight must be given.
// The weights are normalized to sum to 1.
// NewAlias([]float64{1, 2, 3}) will return an Alias that will
// return 0 with probability 1/6, 1 with probability 1/3 and 2 with
// probability 1/2.
func NewAlias(weights []float64) (Alias, error) {
	n := len(weights)

	if n < 1 {
		return nil, fmt.Errorf(
			"at least one weight is required",
		)
	}

	if int(uint32(n)) != n {
		return nil, fmt.Errorf(
			"too many weights, max is %d",
			1<<32-1,
		)
	}

	sum := 0.0

	for idx, weight := range weights {
		if weight <= 0 {
			return nil,
				fmt.Errorf("a weight at index %v is non-positive %v",
					idx,
					weight,
				)
		}
		sum += weight
	}

	alias := aliasImpl{
		table: make([]int32PieceImpl, n),
	}

	twins := make([]float64PieceImpl, n)

	smallTop := -1
	largeBottom := n

	multiplier := float64(n) / sum
	for i, weight := range weights {
		weight *= multiplier
		if weight >= 1 {
			largeBottom--
			twins[largeBottom] = float64PieceImpl{
				weight,
				uint32(i),
			}
		} else {
			smallTop++
			twins[smallTop] = float64PieceImpl{
				weight,
				uint32(i),
			}
		}
	}
	for smallTop >= 0 && largeBottom < n {
		l := twins[smallTop]
		smallTop--

		t := twins[largeBottom]
		largeBottom++

		alias.table[l.alias].probability = uint32(l.probability * (1<<31 - 1))
		alias.table[l.alias].alias = t.alias

		t.probability = (t.probability + l.probability) - 1

		if t.probability < 1 {
			smallTop++
			twins[smallTop] = t
		} else {
			largeBottom--
			twins[largeBottom] = t
		}
	}
	for i := n - 1; i >= largeBottom; i-- {
		alias.table[twins[i].alias].probability = 1<<31 - 1
	}
	for i := 0; i <= smallTop; i++ {
		alias.table[twins[i].alias].probability = 1<<31 - 1
	}
	return &alias, nil
}

func (a *aliasImpl) Sample(random *rand.Rand) int {
	ri := uint32(random.Int31())
	w := ri % uint32(len(a.table))
	if ri > a.table[w].probability {
		return int(a.table[w].alias)
	}
	return int(w)
}

type aliasImpl struct {
	table []int32PieceImpl
}

type float64PieceImpl struct {
	probability float64
	alias       uint32
}

type int32PieceImpl struct {
	probability uint32
	alias       uint32
}
