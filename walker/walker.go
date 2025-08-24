package walker

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/decatile/treecmp/dispatcher"
)

func Walk(dirA string, dirB string, options dispatcher.Options) error {
	dispatcher := dispatcher.Run(options)
	err := walk(dirA, dirB, dispatcher)
	err2 := dispatcher.Close()
	if err != nil {
		return err
	}
	return err2
}

func walk(dirA string, dirB string, dispatcher dispatcher.Emitter) error {
	entriesA, err := os.ReadDir(dirA)
	if err != nil {
		return err
	}
	entriesB, err := os.ReadDir(dirB)
	if err != nil {
		return err
	}
	if len(entriesA) != len(entriesB) {
		return fmt.Errorf(
			"directories differ in entries amount: '%s' -> '%d', when '%s' -> '%d'",
			dirA, len(entriesA), dirB, len(entriesB),
		)
	}
	slices.SortFunc(entriesA, dirComparer)
	slices.SortFunc(entriesB, dirComparer)
	for i := range len(entriesA) {
		infoA, err := entriesA[i].Info()
		if err != nil {
			return err
		}
		infoB, err := entriesB[i].Info()
		if err != nil {
			return err
		}
		if infoA.Name() != infoB.Name() {
			entryType := "file"
			if infoA.IsDir() {
				entryType = "directory"
			}
			return fmt.Errorf(
				"expected %s '%s' at '%s'",
				entryType, infoA.Name(), dirB,
			)
		}
		filepathA := filepath.Join(dirA, infoA.Name())
		filepathB := filepath.Join(dirB, infoB.Name())
		if infoA.IsDir() {
			err = walk(filepathA, filepathB, dispatcher)
			if err != nil {
				return err
			}
		} else {
			if infoA.Size() != infoB.Size() {
				return fmt.Errorf(
					"files differ in size: '%s/%s' -> %dB, when '%s/%s' -> %dB",
					dirA, infoA.Name(), infoA.Size(),
					dirB, infoB.Name(), infoB.Size(),
				)
			}
			if !dispatcher.Emit(filepathA, filepathB) {
				break
			}
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
