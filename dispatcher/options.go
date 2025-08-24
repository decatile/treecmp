package dispatcher

import "context"

type Options struct {
	MetadataOnly bool
	Threads      int
	QueueSize    int
	Context      context.Context
}
