package strings

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

func Do(input io.Reader, minTextSize uint) *bufio.Scanner {
	outputReader, outputWriter := io.Pipe()
	go searchstrings(input, minTextSize, outputWriter)
	return bufio.NewScanner(outputReader)
}

func searchstrings(input io.Reader, minTextSize uint, output *io.PipeWriter) {

	newline := byte('\n')
	searcher := wordSearcher{minTextSize: minTextSize}

	write := func(data []byte) error {
		data = append(data, newline)
		n, err := output.Write(data)
		if n != len(data) {
			return fmt.Errorf(
				"expected to write[%d] wrote[%d]",
				len(data),
				n,
			)
		}
		return err
	}

	handleIOError := func(err error) bool {
		if err != nil {
			var finalwriteerr error

			if text, ok := searcher.flushBuffer(); ok {
				finalwriteerr = write(text)
			}
			if err == io.EOF {
				if finalwriteerr == nil {
					output.Close()
				} else {
					output.CloseWithError(fmt.Errorf(
						"error[%s] writing last data",
						finalwriteerr,
					))
				}
			} else {
				output.CloseWithError(err)
			}
			return true
		}
		return false
	}

	data := make([]byte, 1)
	for {
		// WHY: Don't see the point of checking N when reading a single byte
		n, err := input.Read(data)

		if n <= 0 {
			if handleIOError(err) {
				return
			}
			// WHY:
			// Implementations of Read are discouraged from
			// returning a zero byte count with a nil error,
			// except when len(p) == 0.
			// Callers should treat a return of 0 and nil as
			// indicating that nothing happened; in particular it
			// does not indicate EOF.
			continue
		}

		if text, ok := searcher.next(data[0]); ok {
			err = write(text)
		}

		if handleIOError(err) {
			return
		}
	}
}

type byteType int

type wordSearcher struct {
	buffer         []byte
	possibleRune   []byte
	waitingForRune bool
	minTextSize    uint
}

const (
	binaryType byteType = iota
	asciiType
	runeStartType
)

func (w *wordSearcher) next(b byte) ([]byte, bool) {
	if w.waitingForRune {
		return w.nextRune(b)
	}
	return w.nextASCII(b)
}

func (w *wordSearcher) nextRune(b byte) ([]byte, bool) {

	const maxUTFSize = 4

	if word := string([]byte{b}); utf8.ValidString(word) {
		w.resetRuneSearch()
		text, ok := w.flushBuffer()
		w.writeOnBuffer(b)
		return text, ok
	}

	if utf8.RuneStart(b) {
		// TODO: write test to exercise flush of previous text on this
		// case since what looked like a rune was actually binary data.
		w.resetRuneSearch()
		w.startRuneSearch(b)
		return nil, false
	}

	w.writeOnPossibleRune(b)
	if utf8.ValidString(string(w.possibleRune)) {
		w.writeOnBuffer(w.possibleRune...)
		w.resetRuneSearch()
		return nil, false
	}

	if len(w.possibleRune) == maxUTFSize {
		w.resetRuneSearch()
		return w.flushBuffer()
	}

	return nil, false
}

func (w *wordSearcher) resetRuneSearch() {
	w.waitingForRune = false
	w.possibleRune = nil
}

func (w *wordSearcher) nextASCII(b byte) ([]byte, bool) {
	switch bytetype(b) {
	case binaryType:
		{
			return w.flushBuffer()
		}
	case asciiType:
		{
			w.writeOnBuffer(b)
		}
	case runeStartType:
		{
			w.startRuneSearch(b)
		}
	}
	return nil, false
}

func (w *wordSearcher) startRuneSearch(b byte) {
	w.waitingForRune = true
	w.writeOnPossibleRune(b)
}

func (w *wordSearcher) writeOnBuffer(b ...byte) {
	w.buffer = append(w.buffer, b...)
}

func (w *wordSearcher) writeOnPossibleRune(b byte) {
	w.possibleRune = append(w.possibleRune, b)
}

func (w *wordSearcher) bufferLenInRunes() uint {
	return uint(len([]rune(string(w.buffer))))
}

func (w *wordSearcher) flushBuffer() ([]byte, bool) {
	if len(w.buffer) == 0 {
		return nil, false
	}
	if w.bufferLenInRunes() < w.minTextSize {
		w.buffer = nil
		return nil, false
	}
	b := w.buffer
	w.buffer = nil
	return b, true
}

func bytetype(b byte) byteType {
	if word := string([]byte{b}); utf8.ValidString(word) {
		return asciiType
	}
	if utf8.RuneStart(b) {
		return runeStartType
	}
	return binaryType
}
