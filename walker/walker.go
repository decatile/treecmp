package walker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decatile/treecmp/files"
)

type walker struct {
	dir1     string
	dir2     string
	sink     chan<- error
	comparer files.Comparer
}

func (w walker) cd(dir1, dir2 string) walker {
	w.dir1 = dir1
	w.dir2 = dir2
	return w
}

func (w walker) walk() walkResult {
	entries1, err := os.ReadDir(w.dir1)
	if err != nil {
		w.sink <- fmt.Errorf("failed to read directory %s: %w", w.dir1, err)
		return resultFail
	}
	entries2, err := os.ReadDir(w.dir2)
	if err != nil {
		w.sink <- fmt.Errorf("failed to read directory %s: %w", w.dir2, err)
		return resultFail
	}
	if len(entries1) != len(entries2) {
		w.sink <- fmt.Errorf(
			"directories differ in entries amount:\n%s -> %d\n%s -> %d",
			w.dir1, len(entries1), w.dir2, len(entries2),
		)
		return resultFail
	}
	entriesMap2 := make(map[string]os.DirEntry)
	entriesNotFound1 := make([]string, 0)
	for _, entry2 := range entries2 {
		entriesMap2[entry2.Name()] = entry2
	}
	res := resultOk
	for _, entry1 := range entries1 {
		entry2, ok := entriesMap2[entry1.Name()]
		if ok {
			delete(entriesMap2, entry1.Name())
			info1, err := entry1.Info()
			if err != nil {
				w.sink <- fmt.Errorf("failed to read file info: %w", err)
				res = resultFail
			}
			info2, err := entry2.Info()
			if err != nil {
				w.sink <- fmt.Errorf("failed to read file info: %w", err)
				res = resultFail
			}
			switch w.walkFile(info1, info2) {
			case resultOk:
			case resultFail:
				res = resultFail
			case resultCancel:
				return resultCancel
			}
		} else {
			entriesNotFound1 = append(entriesNotFound1, entry1.Name())
		}
	}
	if len(entriesNotFound1) > 0 {
		entriesNotFound2 := make([]string, 0, len(entriesMap2))
		for name := range entriesMap2 {
			entriesNotFound2 = append(entriesNotFound2, name)
		}
		w.sink <- fmt.Errorf(
			"not matching files in directories %s and %s:\n+ %s\n- %s",
			w.dir1, w.dir2,
			strings.Join(entriesNotFound1, "\n+ "),
			strings.Join(entriesNotFound2, "\n- "),
		)
		res = resultFail
	}
	return res
}

func (w *walker) walkFile(info1 os.FileInfo, info2 os.FileInfo) walkResult {
	if info1.Name() != info2.Name() {
		entryType := "file"
		if info1.IsDir() {
			entryType = "directory"
		}
		w.sink <- fmt.Errorf(
			"expected %s %s at %s",
			entryType, info1.Name(), w.dir2,
		)
		return resultFail
	}
	filepathA := filepath.Join(w.dir1, info1.Name())
	filepathB := filepath.Join(w.dir2, info2.Name())
	if info1.IsDir() {
		return w.cd(filepathA, filepathB).walk()
	} else {
		if info1.Size() != info2.Size() {
			w.sink <- fmt.Errorf(
				"files differ in size:\n* %s/%s -> %dB\n* %s/%s -> %dB",
				w.dir1, info1.Name(), info1.Size(),
				w.dir2, info2.Name(), info2.Size(),
			)
			return resultFail
		}
		err := w.comparer.Compare(filepathA, filepathB)
		switch err {
		case nil:
			return resultOk
		case context.Canceled, context.DeadlineExceeded:
			return resultCancel
		default:
			return resultFail
		}
	}
}

func Walk(ctx context.Context, dir1, dir2 string, concurrency int) error {
	errs := make(chan error)
	defer close(errs)
	go func() {
		for err := range errs {
			fmt.Println(err)
		}
	}()
	comparer := files.NewComparer(ctx, concurrency, errs)
	res := walker{dir1, dir2, errs, comparer}.walk()
	comparer.Close()
	switch res {
	case resultFail:
		return errors.New("one or more errors occurred")
	case resultCancel:
		return errors.New("interrupted")
	}
	return nil
}
