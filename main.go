package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/decatile/treecmp/walker"
)

const Help = `Usage: treecmp [flags] dir1 dir2
Flags: -t, --threads   num routines (default: %d)
`

func main() {
	var tasks int

	flag.CommandLine.Usage = func() { fmt.Printf(Help, runtime.NumCPU()) }
	flag.IntVar(&tasks, "t", runtime.NumCPU(), "")
	flag.IntVar(&tasks, "threads", runtime.NumCPU(), "")
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

	err := walker.Walk(ctx, treeA, treeB, tasks)
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
