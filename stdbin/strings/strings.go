package strings

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

func Do(input io.Reader, minTextSize int) *bufio.Scanner {
	outputReader, outputWriter := io.Pipe()
	go searchstrings(input, minTextSize, outputWriter)
	return bufio.NewScanner(outputReader)
}

func searchstrings(input io.Reader, minTextSize int, output *io.PipeWriter) {

	// TODO: This still don't cover utf-8 corner cases (possibly a lot of them)
	// Also not respecting minTextSize...yet =)

	newline := []byte("\n")
	buffer := []byte{}

	writeOnBuffer := func(d []byte) {
		buffer = append(buffer, d...)
	}

	flushBuffer := func() {
		if len(buffer) == 0 {
			return
		}
		buffer = append(buffer, newline...)
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
			writeOnBuffer(data)
		} else {
			flushBuffer()
		}
	}
}
