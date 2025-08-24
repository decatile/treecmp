package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/decatile/treecmp/dispatcher"
	"github.com/decatile/treecmp/walker"
)

const Help = `treecmp - compare directories recursively

-f, --failfast    exit after first error          (default: false)
-m, --metadata    check only metadata             (default: false)
-t, --threads     num routines                    (default: %d)
-q, --queue       size of task queue for routines (default: 1)
`

func main() {
	var (
		failfast bool
		metadata bool
		threads  int
		queue    int
	)

	flag.CommandLine.Usage = func() { fmt.Printf(Help, runtime.NumCPU()) }

	flag.BoolVar(&failfast, "f", false, "")
	flag.BoolVar(&failfast, "failfast", false, "")

	flag.BoolVar(&metadata, "m", false, "")
	flag.BoolVar(&metadata, "metadata", false, "")

	flag.IntVar(&threads, "t", runtime.NumCPU(), "")
	flag.IntVar(&threads, "threads", runtime.NumCPU(), "")

	flag.IntVar(&queue, "q", 1, "")
	flag.IntVar(&queue, "queue", 1, "")

	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("2 arguments (directory roots for comparison) required")
		os.Exit(1)
	}

	treeA := flag.Arg(0)
	treeB := flag.Arg(1)

	ensureRoot(treeA)
	ensureRoot(treeB)

	treeA = filepath.Clean(treeA)
	treeB = filepath.Clean(treeB)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	err := walker.Walk(treeA, treeB, dispatcher.Options{
		MetadataOnly: metadata,
		Failfast:     failfast,
		Threads:      threads,
		QueueSize:    queue,
		Context:      ctx,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ensureRoot(path string) {
	if path == "" {
		fmt.Println("root is an empty string")
		os.Exit(1)
	}
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		fmt.Printf("'%s' is not an accessible directory\n", path)
		os.Exit(1)
	}
}
