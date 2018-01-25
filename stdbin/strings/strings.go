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
	buffer := []byte{}

	writeOnBuffer := func(word string) {
		buffer = append(buffer, []byte(word)...)
	}

	bufferLenInRunes := func() uint {
		return uint(len([]rune(string(buffer))))
	}

	flushBuffer := func() {
		if len(buffer) == 0 {
			return
		}
		if bufferLenInRunes() < minTextSize {
			buffer = nil
			return
		}
		buffer = append(buffer, newline)
		n, err := output.Write(buffer)
		if n != len(buffer) {
			output.CloseWithError(fmt.Errorf("strings:fatal wrote[%d] bytes wanted[%d]\n", n, len(buffer)))
			return
		}
		if err != nil {
			output.CloseWithError(fmt.Errorf("strings:fatal error[%s] writing data\n", err))
			return
		}
		buffer = nil
	}

	handleIOError := func(err error) bool {
		if err == io.EOF {
			flushBuffer()
			output.Close()
			return true
		}
		if err != nil {
			flushBuffer()
			output.CloseWithError(err)
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

		switch bytetype(data[0]) {
		case binaryType:
			{
				flushBuffer()
			}
		case asciiType:
			{
				writeOnBuffer(string(data[0]))
			}
		case runeStartType:
			{
				if word, flush, ok := searchNonASCII(input, data[0]); ok {
					if flush {
						flushBuffer()
					}
					writeOnBuffer(word)
				} else {
					flushBuffer()
				}
			}
		}

		if handleIOError(err) {
			return
		}
	}
}

type byteType int

const (
	binaryType byteType = iota
	asciiType
	runeStartType
)

func bytetype(b byte) byteType {
	if word := string([]byte{b}); utf8.ValidString(word) {
		return asciiType
	}
	if utf8.RuneStart(b) {
		return runeStartType
	}
	return binaryType
}

func searchNonASCII(input io.Reader, first byte) (string, bool, bool) {
	data := make([]byte, 1)
	buffer := []byte{first}
	// WHY: We already have the first byte, 3 missing
	missingCharsForUTF := 3

	for i := 0; i < missingCharsForUTF; i++ {
		// TODO: Test Read errors here
		input.Read(data)
		if word := string(data); utf8.ValidString(word) {
			// WHY: Valid ASCII range after something that looked
			// like a possible char outsize ASCII
			return word, true, true
		}
		buffer = append(buffer, data[0])
		possibleWord := string(buffer)
		if utf8.ValidString(possibleWord) {
			return possibleWord, false, true
		}
	}

	return "", false, false
}
