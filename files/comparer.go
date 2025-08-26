package files

import (
	"context"

	"github.com/decatile/treecmp/files/internal/comparer"
)

type Comparer interface {
	Close() error
	Compare(file1 string, file2 string) error
}

func NewComparer(ctx context.Context, numTasks int, sink chan<- error) Comparer {
	return comparer.New(ctx, numTasks, sink)
}
