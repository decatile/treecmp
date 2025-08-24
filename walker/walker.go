package walker

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/decatile/treecmp/dispatcher"
)

type walker struct {
	dirA       string
	dirB       string
	failfast   bool
	dispatcher dispatcher.Emitter
}

func (w walker) cd(dirA, dirB string) walker {
	w.dirA = dirA
	w.dirB = dirB
	return w
}

func (w walker) walk() error {
	entriesA, err := os.ReadDir(w.dirA)
	if err != nil {
		return err
	}
	entriesB, err := os.ReadDir(w.dirB)
	if err != nil {
		return err
	}
	if len(entriesA) != len(entriesB) {
		return fmt.Errorf(
			"directories differ in entries amount: '%s' -> '%d', when '%s' -> '%d'",
			w.dirA, len(entriesA), w.dirB, len(entriesB),
		)
	}
	slices.SortFunc(entriesA, dirComparer)
	if w.failfast {
		slices.SortFunc(entriesB, dirComparer)
		for i, entryA := range entriesA {
			infoA, err := entryA.Info()
			if err != nil {
				return err
			}
			infoB, err := entriesB[i].Info()
			if err != nil {
				return err
			}
			err = w.walkFile(infoA, infoB)
			if err == context.Canceled {
				break
			}
			if err != nil {
				return err
			}
		}
	} else {
		errs := make([]error, 0)
		entriesMapB := make(map[string]os.DirEntry)
		entriesNotFoundA := make([]string, 0)
		for _, entry := range entriesB {
			entriesMapB[entry.Name()] = entry
		}
		for _, entryA := range entriesA {
			entryB, ok := entriesMapB[entryA.Name()]
			if ok {
				delete(entriesMapB, entryA.Name())
				infoA, err := entryA.Info()
				if err != nil {
					errs = append(errs, err)
					continue
				}
				infoB, err := entryB.Info()
				if err != nil {
					errs = append(errs, err)
					continue
				}
				err = w.walkFile(infoA, infoB)
				if err == context.Canceled {
					break
				}
				if err != nil {
					errs = append(errs, err)
				}
			} else {
				entriesNotFoundA = append(entriesNotFoundA, entryA.Name())
			}
		}
		if len(entriesNotFoundA) > 0 {
			entriesNotFoundB := make([]string, 0, len(entriesMapB))
			for name := range entriesMapB {
				entriesNotFoundB = append(entriesNotFoundB, name)
			}
			errs = append(errs, fmt.Errorf(
				"not matching files in directories '%s' and '%s':\n+ %s\n- %s",
				w.dirA, w.dirB,
				strings.Join(entriesNotFoundA, "\n+ "),
				strings.Join(entriesNotFoundB, "\n- "),
			))
		}
		return errors.Join(errs...)
	}
	return nil
}

func (w *walker) walkFile(infoA os.FileInfo, infoB os.FileInfo) error {
	if infoA.Name() != infoB.Name() {
		entryType := "file"
		if infoA.IsDir() {
			entryType = "directory"
		}
		return fmt.Errorf(
			"expected %s '%s' at '%s'",
			entryType, infoA.Name(), w.dirB,
		)
	}
	filepathA := filepath.Join(w.dirA, infoA.Name())
	filepathB := filepath.Join(w.dirB, infoB.Name())
	if infoA.IsDir() {
		err := w.cd(filepathA, filepathB).walk()
		if err != nil {
			return err
		}
	} else {
		if infoA.Size() != infoB.Size() {
			return fmt.Errorf(
				"files differ in size: '%s/%s' -> %dB, when '%s/%s' -> %dB",
				w.dirA, infoA.Name(), infoA.Size(),
				w.dirB, infoB.Name(), infoB.Size(),
			)
		}
		if !w.dispatcher.Emit(filepathA, filepathB) {
			return context.Canceled
		}
	}
	return nil
}

func dirComparer(a os.DirEntry, b os.DirEntry) int {
	if a.IsDir() == b.IsDir() {
		return cmp.Compare(a.Name(), b.Name())
	}
	if a.IsDir() {
		return -1
	}
	return 1
}

func Walk(dirA string, dirB string, opt dispatcher.Options) error {
	dispatcher := dispatcher.Run(opt)
	err := walker{dirA, dirB, opt.Failfast, dispatcher}.walk()
	err2 := dispatcher.Close()
	if err != nil {
		return err
	}
	return err2
}
