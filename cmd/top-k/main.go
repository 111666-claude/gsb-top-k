// Command top-k 用固定种子喂一批条目并打印前 K 名。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/topk"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("top-k", flag.ContinueOnError)
	flags.SetOutput(stderr)
	k := flags.Int("k", 5, "要选出的条目数")
	count := flags.Int("generate", 1000, "生成多少个条目")
	held := flags.Bool("held", false, "额外打印当前保留的条目数")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	selector := topk.New(*k)
	state := uint64(20260927)
	for index := 0; index < *count; index++ {
		state = state*6364136223846793005 + 1442695040888963407
		score := int64(state>>50) % 1000
		selector.Push(fmt.Sprintf("k%05d", index), score)
	}
	for _, item := range selector.Top() {
		fmt.Fprintf(stdout, "%s\t%d\n", item.Key, item.Score)
	}
	fmt.Fprintf(stdout, "k=%d generate=%d compared=%d"+"\n", *k, *count, selector.Compared())
	if *held {
		fmt.Fprintf(stdout, "held=%d\n", selector.Held())
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
