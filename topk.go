// Package topk 做流式 top-K 选择。
package topk

import "sort"

// Item 是带分数的条目。
type Item struct {
	Key   string
	Score int64
}

// Selector 是流式选择器。缺陷：把全部条目存下来，最后才排序切片。
type Selector struct {
	k        int
	items    []Item
	compared int
}

// New 建一个选 K 个的 selector。
func New(k int) *Selector { return &Selector{k: k} }

// Push 喂一个条目。缺陷：全部保留，没有阈值剪枝。
func (s *Selector) Push(key string, score int64) {
	s.items = append(s.items, Item{Key: key, Score: score})
}

// Top 返回前 k 名。缺陷：排序方向反了（按分数升序取前 k）。
func (s *Selector) Top() []Item {
	sorted := append([]Item{}, s.items...)
	sort.SliceStable(sorted, func(i, j int) bool {
		s.compared++
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score < sorted[j].Score
		}
		return sorted[i].Key < sorted[j].Key
	})
	if len(sorted) > s.k {
		return sorted[:s.k]
	}
	return sorted
}

// Held 返回当前保留的条目数（统计用）。
func (s *Selector) Held() int { return len(s.items) }

// Compared 返回比较次数（统计用）。
func (s *Selector) Compared() int { return s.compared }
