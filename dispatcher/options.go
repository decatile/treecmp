package dispatcher

import "context"

type Options struct {
	Failfast     bool
	MetadataOnly bool
	Threads      int
	QueueSize    int
	Context      context.Context
}
