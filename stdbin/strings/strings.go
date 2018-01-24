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

	data := make([]byte, 1)
	for {
		// WHY: Don't see the point of checking N when reading a single byte
		_, err := input.Read(data)
		if err == io.EOF {
			flushBuffer()
			output.Close()
			return
		}
		if err != nil {
			flushBuffer()
			output.CloseWithError(err)
			return
		}

		if word := string(data); utf8.ValidString(word) {
			writeOnBuffer(word)
		} else if utf8.RuneStart(data[0]) {
			if word, ok := searchNonASCII(input, data[0]); ok {
				writeOnBuffer(word)
			} else {
				flushBuffer()
			}
		} else {
			flushBuffer()
		}
	}
}

func searchNonASCII(input io.Reader, first byte) (string, bool) {
	data := make([]byte, 1)
	buffer := []byte{first}
	// WHY: We already have the first byte, 3 missing
	missingCharsForUTF := 3

	for i := 0; i < missingCharsForUTF; i++ {
		// WHY: ignoring read errors here will cause us to
		// go back to the main search loop and eventually
		// will call Read again which will fail again.
		// Perhaps not a very good idea.
		input.Read(data)
		if word := string(data); utf8.ValidString(word) {
			// WHY: Valid ASCII range after something that looked
			// like a possible char outsize ASCII
			return word, true
		}
		buffer = append(buffer, data[0])
		possibleWord := string(buffer)
		if utf8.ValidString(possibleWord) {
			return possibleWord, true
		}
	}

	return "", false
}
