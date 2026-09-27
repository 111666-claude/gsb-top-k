package topk

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestTopLengthIsAtMostK(t *testing.T) {
	selector := New(2)
	for index := 0; index < 5; index++ {
		selector.Push("k", int64(index))
	}
	if got := len(selector.Top()); got != 2 {
		t.Fatalf("应该返回 2 个：%d", got)
	}
	top := selector.Top()
	if top[0].Score != 4 || top[1].Score != 3 {
		t.Fatalf("应取最大的两个（4, 3），实际：%v", top)
	}
}

func TestSelectsLargestScores(t *testing.T) {
	selector := New(3)
	scores := []int64{25, 90, 1, 88, 90, 42, 100}
	for index, score := range scores {
		selector.Push(fmt.Sprintf("k%02d", index), score)
	}
	got := selector.Top()
	want := []Item{{"k06", 100}, {"k01", 90}, {"k04", 90}}
	assertItems(t, got, want)
}

func TestTiesOrderedByKeyAscending(t *testing.T) {
	selector := New(3)
	selector.Push("c", 50)
	selector.Push("a", 50)
	selector.Push("b", 50)
	selector.Push("d", 10)
	got := selector.Top()
	want := []Item{{"a", 50}, {"b", 50}, {"c", 50}}
	assertItems(t, got, want)
}

func TestEmptySelector(t *testing.T) {
	top := New(3).Top()
	if len(top) != 0 {
		t.Fatalf("空选择器应返回空：%v", top)
	}
}

func TestKZeroAndNegative(t *testing.T) {
	for _, k := range []int{0, -1, -100} {
		selector := New(k)
		selector.Push("a", 1)
		selector.Push("b", 2)
		if got := selector.Top(); len(got) != 0 {
			t.Fatalf("k=%d 应返回空：%v", k, got)
		}
		if selector.Held() != 0 {
			t.Fatalf("k=%d 时 Held 应为 0：%d", k, selector.Held())
		}
	}
}

func TestFewerItemsThanK(t *testing.T) {
	selector := New(10)
	selector.Push("a", 1)
	selector.Push("b", 3)
	selector.Push("c", 2)
	got := selector.Top()
	want := []Item{{"b", 3}, {"c", 2}, {"a", 1}}
	assertItems(t, got, want)
}

func TestThresholdPruningHeldIsBoundedByK(t *testing.T) {
	selector := New(3)
	for index := 0; index < 100; index++ {
		selector.Push(fmt.Sprintf("k%03d", index), int64(index))
		// 不变量：任何时刻保留条目都不超过 K（README 给的上界是 ≤ 2K）。
		if selector.Held() > 3 {
			t.Fatalf("第 %d 条后 Held=%d 超过 K=3", index, selector.Held())
		}
	}
	if selector.Held() != 3 {
		t.Fatalf("喂满 100 条后 Held 应等于 K=3：%d", selector.Held())
	}
	want := []Item{{"k099", 99}, {"k098", 98}, {"k097", 97}}
	assertItems(t, selector.Top(), want)
}

func TestComparedIsCounted(t *testing.T) {
	selector := New(1)
	selector.Push("a", 1)
	selector.Push("b", 2)
	if selector.Compared() == 0 {
		t.Fatal("比较次数应该被统计")
	}
}

// naiveTopK 是朴素参考实现：存下全部条目，按「分数降序、键升序」
// 稳定排序后取前 K。每调用一次排序比较器记一次比较。
func naiveTopK(items []Item, k int) ([]Item, int) {
	if k <= 0 || len(items) == 0 {
		return []Item{}, 0
	}
	sorted := append([]Item{}, items...)
	compared := 0
	sort.SliceStable(sorted, func(i, j int) bool {
		compared++
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score > sorted[j].Score
		}
		return sorted[i].Key < sorted[j].Key
	})
	if len(sorted) > k {
		sorted = sorted[:k]
	}
	return sorted, compared
}

func TestMatchesNaiveReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20260927))
	// 固定样例：空、不足 K、大量同分，用于覆盖边界。
	fixed := [][]Item{
		{},
		{{"a", 1}},
		{{"c", 7}, {"a", 7}, {"b", 7}},
		{{"a", 1}, {"b", 1}, {"c", 1}, {"d", 1}, {"e", 1}},
	}
	for groupIndex, items := range fixed {
		for _, k := range []int{1, 3, 10} {
			checkAgainstNaive(t, groupIndex, k, items)
		}
	}
	// 1000 组随机数据：随机 N、随机 K、少量可能重复的键与分数。
	for group := 0; group < 1000; group++ {
		n := rng.Intn(200)
		k := rng.Intn(12) // 0..11，覆盖 K=0 与 K>N
		items := make([]Item, n)
		for index := range items {
			items[index] = Item{
				Key:   fmt.Sprintf("k%03d", rng.Intn(n/2+1)),
				Score: int64(rng.Intn(20)),
			}
		}
		checkAgainstNaive(t, group, k, items)
	}
}

func checkAgainstNaive(t *testing.T, group, k int, items []Item) {
	t.Helper()
	selector := New(k)
	for _, item := range items {
		selector.Push(item.Key, item.Score)
	}
	got := selector.Top()
	want, _ := naiveTopK(items, k)
	if len(got) != len(want) {
		t.Fatalf("组 %d k=%d n=%d：长度 %d != %d", group, k, len(items), len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("组 %d k=%d n=%d：第 %d 项 %+v != %+v",
				group, k, len(items), index, got[index], want[index])
		}
	}
	bound := k
	if len(items) < bound {
		bound = len(items)
	}
	if selector.Held() != bound {
		t.Fatalf("组 %d k=%d n=%d：Held=%d，应为 min(K,N)=%d",
			group, k, len(items), selector.Held(), bound)
	}
	// 重复调用 Top 必须得到完全相同的结果（可复现、不破坏状态）。
	assertItems(t, selector.Top(), got)
}

func TestComparedOrderNLogK(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	const n = 20000
	const k = 8
	items := make([]Item, n)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("k%06d", index), Score: rng.Int63()}
	}
	selector := New(k)
	for _, item := range items {
		selector.Push(item.Key, item.Score)
	}
	got := selector.Top()
	want, naiveCompared := naiveTopK(items, k)
	assertItems(t, got, want)

	heapCompared := selector.Compared()
	// 严格上界：每条一次阈值比较；堆调整每层至多两次比较、共 ⌈log₂(K+1)⌉ 层；
	// 前 K 条上浮每层一次比较也被该上界覆盖。
	upper := n * (1 + 2*ceilLog2(k+1))
	if heapCompared > upper {
		t.Fatalf("比较次数 %d 超过 N(1+logK) 上界 %d", heapCompared, upper)
	}
	// 量级对照：流式选择应明显小于朴素全排序（N log N）。
	if heapCompared >= naiveCompared/4 {
		t.Fatalf("比较次数量级不对：堆 %d vs 朴素 %d", heapCompared, naiveCompared)
	}
	t.Logf("N=%d K=%d：堆比较 %d（上界 %d），朴素排序比较 %d",
		n, k, heapCompared, upper, naiveCompared)
}

func TestScaleOneMillionK100(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过百万规模测试")
	}
	const n = 1_000_000
	const k = 100
	selector := New(k)
	state := uint64(20260927)
	for index := 0; index < n; index++ {
		state = state*6364136223846793005 + 1442695040888963407
		selector.Push(fmt.Sprintf("k%07d", index), int64(state>>40)%1_000_000)
	}
	if selector.Held() != k {
		t.Fatalf("百万条后 Held=%d，应为 K=%d", selector.Held(), k)
	}
	top := selector.Top()
	if len(top) != k {
		t.Fatalf("Top 长度应为 %d：%d", k, len(top))
	}
	for index := 1; index < len(top); index++ {
		if top[index-1].Score < top[index].Score {
			t.Fatalf("Top 未按分数降序：位置 %d", index)
		}
		if top[index-1].Score == top[index].Score && top[index-1].Key > top[index].Key {
			t.Fatalf("同分时未按键升序：位置 %d", index)
		}
	}
	upper := n * (1 + 2*ceilLog2(k+1))
	if selector.Compared() > upper {
		t.Fatalf("比较次数 %d 超过 N(1+logK) 上界 %d", selector.Compared(), upper)
	}
	t.Logf("N=%d K=%d：held=%d compared=%d（上界 %d）",
		n, k, selector.Held(), selector.Compared(), upper)
}

func ceilLog2(x int) int {
	bits := 0
	for power := 1; power < x; power <<= 1 {
		bits++
	}
	return bits
}

func assertItems(t *testing.T, got, want []Item) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("长度 %d != %d：got=%v want=%v", len(got), len(want), got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("第 %d 项 %+v != %+v（got=%v）", index, got[index], want[index], got)
		}
	}
}
