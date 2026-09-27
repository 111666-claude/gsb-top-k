// Package topk 做流式 top-K 选择。
package topk

import "sort"

// Item 是带分数的条目。
type Item struct {
	Key   string
	Score int64
}

// Selector 是流式选择器：内部只保留一个容量 K 的最小堆，
// 堆顶始终是当前第 K 名（最差的入围者），用作剪枝阈值。
// 内存 O(K)，每个条目最多做 O(log K) 次比较。
type Selector struct {
	k        int
	heap     []Item // 最小堆，长度不超过 k
	compared int
}

// New 建一个选 K 个的 selector。K <= 0 时任何输入都会被丢弃。
func New(k int) *Selector {
	if k <= 0 {
		k = 0
	}
	return &Selector{k: k}
}

// Push 喂一个条目：先和当前第 K 名（堆顶）比，低于阈值直接丢弃，
// 不进堆；堆未满或优于阈值时才插入并上浮，每次比较都计入 Compared。
func (s *Selector) Push(key string, score int64) {
	item := Item{Key: key, Score: score}
	// 堆已满：与阈值（第 K 名）比较，不优于它就直接剪枝。
	if len(s.heap) == s.k {
		if s.k == 0 || s.worse(item, s.heap[0]) {
			return
		}
		// 替换堆顶（被淘汰的旧第 K 名）并下沉。
		s.heap[0] = item
		s.siftDown(0)
		return
	}
	s.heap = append(s.heap, item)
	s.siftUp(len(s.heap) - 1)
}

// Top 返回前 K 名，按「分数降序、键升序」排列。
// 结果与「全量排序后取前 K」逐项一致；不修改选择器状态，可重复调用。
func (s *Selector) Top() []Item {
	top := append([]Item{}, s.heap...)
	sort.Slice(top, func(i, j int) bool {
		if top[i].Score != top[j].Score {
			return top[i].Score > top[j].Score
		}
		return top[i].Key < top[j].Key
	})
	return top
}

// Held 返回当前保留的条目数（统计用）。
func (s *Selector) Held() int { return len(s.heap) }

// Compared 返回比较次数（统计用）。
func (s *Selector) Compared() int { return s.compared }

// worse 判断 a 是否排在 b 后面（即 a 比 b 更差）：分数低更差，
// 同分则键的字典序大更差。堆顶是「最差的入围者」= 第 K 名。
func (s *Selector) worse(a, b Item) bool {
	s.compared++
	if a.Score != b.Score {
		return a.Score < b.Score
	}
	return a.Key > b.Key
}

func (s *Selector) siftUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		// 堆顶是最差条目：子节点不比父节点差则堆序成立。
		if !s.worse(s.heap[index], s.heap[parent]) {
			return
		}
		s.heap[parent], s.heap[index] = s.heap[index], s.heap[parent]
		index = parent
	}
}

func (s *Selector) siftDown(index int) {
	n := len(s.heap)
	for {
		left := index*2 + 1
		if left >= n {
			return
		}
		smallest := left
		if right := left + 1; right < n && s.worse(s.heap[right], s.heap[left]) {
			smallest = right
		}
		if !s.worse(s.heap[smallest], s.heap[index]) {
			return
		}
		s.heap[index], s.heap[smallest] = s.heap[smallest], s.heap[index]
		index = smallest
	}
}
