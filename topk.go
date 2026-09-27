// Package topk 做流式 top-K 选择。
package topk

import (
	"container/heap"
	"sort"
)

// Item 是带分数的条目。
type Item struct {
	Key   string
	Score int64
}

// better 报告 a 的排名是否高于 b：分数高者优先，同分按键的字典序小者优先。
func better(a, b Item) bool {
	if a.Score != b.Score {
		return a.Score > b.Score
	}
	return a.Key < b.Key
}

// itemHeap 是小顶堆，堆顶是当前保留条目里排名最低（第 K 名）的条目。
type itemHeap struct {
	items    []Item
	compared *int
}

func (h itemHeap) Len() int { return len(h.items) }

// Less 把排名更低的条目放在堆顶。
func (h itemHeap) Less(i, j int) bool {
	*h.compared++
	return better(h.items[j], h.items[i])
}

func (h itemHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *itemHeap) Push(x any) { h.items = append(h.items, x.(Item)) }

func (h *itemHeap) Pop() any {
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[:n-1]
	return item
}

// Selector 是流式选择器：只保留一个大小不超过 K 的堆，内存 O(K)。
type Selector struct {
	k        int
	heap     itemHeap
	compared int
}

// New 建一个选 K 个的 selector。K <= 0 时什么都选不出。
func New(k int) *Selector {
	if k < 0 {
		k = 0
	}
	selector := &Selector{k: k}
	selector.heap.compared = &selector.compared
	return selector
}

// Push 喂一个条目。先和当前第 K 名比，不更好就直接丢，不进堆。
func (s *Selector) Push(key string, score int64) {
	if s.k <= 0 {
		return
	}
	item := Item{Key: key, Score: score}
	if s.heap.Len() < s.k {
		heap.Push(&s.heap, item)
		return
	}
	s.compared++
	if better(item, s.heap.items[0]) {
		heap.Pop(&s.heap)
		heap.Push(&s.heap, item)
	}
}

// Top 返回前 K 名，按「分数降序、键升序」排列。
func (s *Selector) Top() []Item {
	top := append([]Item{}, s.heap.items...)
	sort.Slice(top, func(i, j int) bool {
		s.compared++
		return better(top[i], top[j])
	})
	return top
}

// Held 返回当前保留的条目数（统计用），最多为 K。
func (s *Selector) Held() int { return s.heap.Len() }

// Compared 返回比较次数（统计用），总量级 O(N log K)。
func (s *Selector) Compared() int { return s.compared }
