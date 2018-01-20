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
	data := make([]byte, 1)
	for {
		_, err := input.Read(data)
		if err == io.EOF {
			output.Close()
			return
		}
		if err != nil {
			output.CloseWithError(err)
			return
		}

		if word := string(data); utf8.ValidString(word) {
			_, err = output.Write(data)
			if err != nil {
				output.CloseWithError(fmt.Errorf("strings:fatal error[%s] writing data\n", err))
				return
			}
		}
	}
}
