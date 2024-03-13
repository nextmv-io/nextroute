// © 2019-present nextmv.io inc

package common

import (
	"container/heap"
)

// NSmallest returns the n-smallest items in the slice items using the
// function f to determine the value of each item. If n is greater than
// the length of items, all items are returned.
func NSmallest[T any](items []T, f func(T) float64, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(items) {
		return items
	}
	h := &minHeap[T]{}
	heap.Init(h)

	for _, item := range items {
		value := f(item)
		if h.Len() < n {
			h.Push(itemValue[T]{item, f(item)})
		} else if h.Peek().value > value {
			heap.Pop(h)
			heap.Push(h, itemValue[T]{
				item:  item,
				value: value,
			})
		}
	}
	result := make([]T, n)
	for i := 0; i < n; i++ {
		result[i] = heap.Pop(h).(itemValue[T]).item
	}
	return result
}

type itemValue[T any] struct {
	item  T
	value float64
}

type minHeap[T any] []itemValue[T]

func (h minHeap[T]) Len() int {
	return len(h)
}

func (h minHeap[T]) Less(i, j int) bool {
	return h[i].value < h[j].value
}

func (h minHeap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *minHeap[T]) Push(x interface{}) {
	*h = append(*h, x.(itemValue[T]))
}

func (h *minHeap[T]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *minHeap[T]) Peek() itemValue[T] {
	return (*h)[0]
}
