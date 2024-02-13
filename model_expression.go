package nextroute

import "sync/atomic"

var expressionIndex uint32

// NewModelExpressionIndex returns the next unique expression index.
func NewModelExpressionIndex() int {
	return int(atomic.AddUint32(&expressionIndex, 1) - 1)
}
