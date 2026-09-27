package topk

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

// naiveTop 是朴素参考实现：存全部条目，按（分数降序、键升序）排序后取前 K。
// 返回结果与比较次数。
func naiveTop(items []Item, k int) ([]Item, int) {
	if k <= 0 {
		return nil, 0
	}
	sorted := append([]Item{}, items...)
	compared := 0
	sort.SliceStable(sorted, func(i, j int) bool {
		compared++
		return better(sorted[i], sorted[j])
	})
	if len(sorted) > k {
		sorted = sorted[:k]
	}
	return sorted, compared
}

func assertSameItems(t *testing.T, want, got []Item) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("长度不一致：want %d got %d（got=%v）", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 项不一致：want %+v got %+v", i, want[i], got[i])
		}
	}
}

func TestTopLengthIsAtMostK(t *testing.T) {
	selector := New(2)
	for index := 0; index < 5; index++ {
		selector.Push("k", int64(index))
	}
	if got := len(selector.Top()); got != 2 {
		t.Fatalf("应该返回 2 个：%d", got)
	}
}

func TestSelectsLargestScores(t *testing.T) {
	selector := New(3)
	for index := 0; index < 10; index++ {
		selector.Push(fmt.Sprintf("k%d", index), int64(index))
	}
	want := []Item{{"k9", 9}, {"k8", 8}, {"k7", 7}}
	assertSameItems(t, want, selector.Top())
}

func TestEmptySelector(t *testing.T) {
	if got := len(New(3).Top()); got != 0 {
		t.Fatalf("空选择器应返回空：%d", got)
	}
}

func TestNonPositiveK(t *testing.T) {
	for _, k := range []int{0, -1, -100} {
		selector := New(k)
		selector.Push("a", 1)
		selector.Push("b", 2)
		if got := len(selector.Top()); got != 0 {
			t.Fatalf("k=%d 应返回空：%d", k, got)
		}
		if selector.Held() != 0 {
			t.Fatalf("k=%d 不应保留条目：%d", k, selector.Held())
		}
	}
}

func TestFewerItemsThanK(t *testing.T) {
	selector := New(10)
	selector.Push("a", 1)
	selector.Push("b", 2)
	want := []Item{{"b", 2}, {"a", 1}}
	assertSameItems(t, want, selector.Top())
}

func TestTieBreakByKeyAscending(t *testing.T) {
	selector := New(3)
	selector.Push("c", 5)
	selector.Push("a", 5)
	selector.Push("b", 5)
	selector.Push("d", 5)
	want := []Item{{"a", 5}, {"b", 5}, {"c", 5}}
	assertSameItems(t, want, selector.Top())
}

func TestHeldIsBoundedByK(t *testing.T) {
	const k = 10
	selector := New(k)
	rng := rand.New(rand.NewSource(1))
	for index := 0; index < 100000; index++ {
		selector.Push(fmt.Sprintf("k%d", index), rng.Int63n(1000))
		if selector.Held() > k {
			t.Fatalf("Held 不应超过 K=%d：%d", k, selector.Held())
		}
	}
	if selector.Held() != k {
		t.Fatalf("喂了足够多的条目后 Held 应等于 K=%d：%d", k, selector.Held())
	}
}

func TestComparedIsNLogK(t *testing.T) {
	const (
		n = 200000
		k = 100
	)
	items := make([]Item, n)
	rng := rand.New(rand.NewSource(2))
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("k%d", index), Score: rng.Int63n(1000000)}
	}

	selector := New(k)
	for _, item := range items {
		selector.Push(item.Key, item.Score)
	}
	selector.Top()

	_, naiveCompared := naiveTop(items, k)
	if selector.Compared() >= naiveCompared {
		t.Fatalf("流式比较次数应小于朴素全排序：streaming=%d naive=%d",
			selector.Compared(), naiveCompared)
	}
	// 每次 Push 至多 1 次阈值比较加 2 次堆调整（各 log2(K) 次比较），
	// Top 再花 K log K；总量应远小于 N log N。
	bound := 3*int(math.Ceil(math.Log2(k+1)))*n/4 + k*int(math.Ceil(math.Log2(k+1)))
	if selector.Compared() > bound {
		t.Fatalf("比较次数应为 O(N log K)：got %d，bound %d", selector.Compared(), bound)
	}
	if naiveCompared < n {
		t.Fatalf("朴素实现比较次数应至少为 N 量级：%d", naiveCompared)
	}
}

func TestComparedIsCounted(t *testing.T) {
	selector := New(1)
	selector.Push("a", 1)
	selector.Push("b", 2)
	selector.Top()
	if selector.Compared() == 0 {
		t.Fatal("比较次数应该被统计")
	}
}

func TestDeterministicAcrossRuns(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	items := make([]Item, 5000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("k%d", rng.Intn(500)), Score: rng.Int63n(50)}
	}
	first := runTopK(items, 20)
	for run := 0; run < 5; run++ {
		assertSameItems(t, first, runTopK(items, 20))
	}
}

func runTopK(items []Item, k int) []Item {
	selector := New(k)
	for _, item := range items {
		selector.Push(item.Key, item.Score)
	}
	return selector.Top()
}

// TestMatchesNaiveReference 用样例与 1000 组随机数据逐项比对流式实现与朴素参考实现。
func TestMatchesNaiveReference(t *testing.T) {
	sample := []Item{
		{"a", 25}, {"b", 32}, {"c", 35}, {"d", 35}, {"e", 10}, {"f", 99},
	}
	cases := [][]Item{sample, nil, {{"x", 1}}}

	rng := rand.New(rand.NewSource(4))
	for len(cases) < 1000 {
		n := rng.Intn(200)
		items := make([]Item, n)
		for index := range items {
			// 键与分数都取自小值域，制造大量同分与同键。
			items[index] = Item{
				Key:   fmt.Sprintf("k%d", rng.Intn(30)),
				Score: int64(rng.Intn(20)),
			}
		}
		cases = append(cases, items)
	}

	for caseIndex, items := range cases {
		k := rng.Intn(50) - 2 // 覆盖 K <= 0 与 K > N
		want, _ := naiveTop(items, k)
		got := runTopK(items, k)
		if len(got) != len(want) {
			t.Fatalf("用例 %d（n=%d k=%d）长度不一致：want %d got %d",
				caseIndex, len(items), k, len(want), len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("用例 %d（n=%d k=%d）第 %d 项不一致：want %+v got %+v",
					caseIndex, len(items), k, i, want[i], got[i])
			}
		}
	}
}
