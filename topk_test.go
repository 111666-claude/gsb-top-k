package topk

import "testing"

func TestTopLengthIsAtMostK(t *testing.T) {
selector := New(2)
for index := 0; index < 5; index++ {
selector.Push("k", int64(index))
}
if got := len(selector.Top()); got != 2 {
t.Fatalf("应该返回 2 个：%d", got)
}
}

func TestEmptySelector(t *testing.T) {
if got := len(New(3).Top()); got != 0 {
t.Fatalf("空选择器应返回空：%d", got)
}
}

func TestFewerItemsThanK(t *testing.T) {
selector := New(10)
selector.Push("a", 1)
if got := len(selector.Top()); got != 1 {
t.Fatalf("不足 k 个就返回全部：%d", got)
}
}

func TestHeldCountsPushed(t *testing.T) {
selector := New(1)
selector.Push("a", 1)
selector.Push("b", 2)
if selector.Held() != 2 {
t.Fatalf("Held 应该等于喂进来的条数：%d", selector.Held())
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
