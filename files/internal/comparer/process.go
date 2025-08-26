package comparer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

const bufferSize = 4096

func process(ctx context.Context, task Task) error {
	file1, err := os.Open(task.File1)
	if err != nil {
		return err
	}
	defer file1.Close()
	file2, err := os.Open(task.File2)
	if err != nil {
		return err
	}
	defer file2.Close()
	var buf1, buf2 [bufferSize]byte
	for {
		if ctx.Err() != nil {
			return nil
		}
		n1, err := file1.Read(buf1[:])
		expectEOF := err == io.EOF
		if !expectEOF && err != nil {
			return fmt.Errorf("failed to read file %s: %w", task.File1, err)
		}
		n2, err := file2.Read(buf2[:])
		if expectEOF {
			if err != io.EOF {
				return fmt.Errorf(
					"despite metadata, files differ in size:\n* %s\n* %s",
					task.File1,
					task.File2,
				)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", task.File2, err)
		}
		if n1 != n2 {
			return fmt.Errorf(
				"despite metadata, files differ in size:\n* %s\n* %s",
				task.File1,
				task.File2,
			)
		}
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return fmt.Errorf(
				"files differ:\n* %s\n* %s",
				task.File1,
				task.File2,
			)
		}
	}
}
