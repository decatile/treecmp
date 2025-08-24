package dispatcher

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const cmpBufferPartLen = 4096

type bufferedComparer [cmpBufferPartLen * 2]byte

func newBufferedComparer() bufferedComparer {
	return [cmpBufferPartLen * 2]byte{}
}

func (buf *bufferedComparer) DispatchCompareRequest(info compareInfo) error {
	fileA, err := os.Open(info.a)
	if err != nil {
		return err
	}
	defer fileA.Close()
	fileB, err := os.Open(info.b)
	if err != nil {
		return err
	}
	defer fileB.Close()
	readerA := bufio.NewReader(fileA)
	readerB := bufio.NewReader(fileB)
	for {
		nA, err := io.ReadFull(readerA, buf[:cmpBufferPartLen])
		if err == io.EOF {
			return nil
		}
		if err != nil && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read '%s'", info.a)
		}
		nB, err := io.ReadFull(readerB, buf[cmpBufferPartLen:])
		if err != nil && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read '%s'", info.b)
		}
		if nA != nB {
			return fmt.Errorf("despite metadata, files '%s' and '%s' differ in size", info.a, info.b)
		}
		if !bytes.Equal(buf[0:nA], buf[cmpBufferPartLen:cmpBufferPartLen+nB]) {
			return fmt.Errorf("files '%s' and '%s' differ", info.a, info.b)
		}
	}
}
