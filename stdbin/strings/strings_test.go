package strings_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/NeowayLabs/nash/stdbin/strings"
)

// TODO: Test first byte is rune start but also a EOF, should not search for rest of the rune

func TestStrings(t *testing.T) {

	type testcase struct {
		name        string
		input       func() []byte
		output      []string
		minWordSize uint
	}

	tcases := []testcase{
		{
			name:        "UTF-8With2Bytes",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("λ"), bin...)
			},
			output: []string{"λ"},
		},
		{
			name:        "UTF-8With3Bytes",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("€"), bin...)
			},
			output: []string{"€"},
		},
		{
			name:        "UTF-8With4Bytes",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("𐍈"), bin...)
			},
			output: []string{"𐍈"},
		},
		{
			name:        "NonASCIIWordHasOneLessCharThanMin",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("λ"), bin...)
			},
			output: []string{},
		},
		{
			name:        "NonASCIIWordHasMinWordSize",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("λλ"), bin...)
			},
			output: []string{"λλ"},
		},
		{
			name:        "WordHasOneLessCharThanMin",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("k"), bin...)
			},
			output: []string{},
		},
		{
			name:        "WordHasMinWordSize",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("kz"), bin...)
			},
			output: []string{"kz"},
		},
		{
			name:        "WordHasOneMoreCharThanMinWordSize",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("ktz"), bin...)
			},
			output: []string{"ktz"},
		},
		{
			name:        "StartingWithOneChar",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("k"), bin...)
			},
			output: []string{"k"},
		},
		{
			name:        "EndWithOneChar",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append(bin, []byte("k")...)
			},
			output: []string{"k"},
		},
		{
			name:        "OneCharInTheMiddle",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				t := append(bin, []byte("k")...)
				t = append(t, bin...)
				return t
			},
			output: []string{"k"},
		},
		{
			name:        "StartingWithText",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnBeggining"
				bin := newBinary(64)
				return append([]byte(expected), bin...)
			},
			output: []string{"textOnBeggining"},
		},
		{
			name:        "TextOnMiddle",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnMiddle"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"textOnMiddle"},
		},
		{
			name:        "NonASCIITextOnMiddle",
			minWordSize: 1,
			input: func() []byte {
				expected := "λλλ"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"λλλ"},
		},
		{
			name:        "ASCIIAndNonASCII",
			minWordSize: 1,
			input: func() []byte {
				expected := "(define (λ (x) (+ x a)))"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"(define (λ (x) (+ x a)))"},
		},
		{
			name:        "TextOnEnd",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnEnd"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"textOnEnd"},
		},
		{
			name:        "JustText",
			minWordSize: 1,
			input: func() []byte {
				return []byte("justtext")
			},
			output: []string{"justtext"},
		},
		{
			name:        "JustBinary",
			minWordSize: 1,
			input: func() []byte {
				return newBinary(64)
			},
			output: []string{},
		},
		{
			name:        "TextSeparatedByBinary",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				text := []byte("text")
				t := []byte{}
				t = append(t, bin...)
				t = append(t, text...)
				t = append(t, bin...)
				t = append(t, text...)
				return t
			},
			output: []string{"text", "text"},
		},
		{
			name:        "NonASCIITextSeparatedByBinary",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				text := []byte("awesomeλ=)")
				t := []byte{}
				t = append(t, bin...)
				t = append(t, text...)
				t = append(t, bin...)
				t = append(t, text...)
				return t
			},
			output: []string{"awesomeλ=)", "awesomeλ=)"},
		},
		{
			name:        "WordsAreNotAccumulativeBetweenBinData",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				t := append([]byte("k"), bin...)
				return append(t, byte('t'))
			},
			output: []string{},
		},
		{
			name:        "ASCIISeparatedByByteThatLooksLikeUTF",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte{
					'n',
					runestart,
					'k',
				}, bin...)
			},
			output: []string{"n", "k"},
		},
		{
			name:        "ASCIIAfterPossibleFirstByteOfUTF",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte{
					runestart,
					'k',
				}, bin...)
			},
			output: []string{"k"},
		},
		{
			name:        "ASCIIAfterPossibleSecondByteOfUTF",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte{
					byte(0xE2),
					byte(0x82),
					'k',
				}, bin...)
			},
			output: []string{"k"},
		},
		{
			name:        "ASCIIAfterPossibleThirdByteOfUTF",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte{
					byte(0xF0),
					byte(0x90),
					byte(0x8D),
					'k',
				}, bin...)
			},
			output: []string{"k"},
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			input := tcase.input()
			scanner := strings.Do(bytes.NewBuffer(input), tcase.minWordSize)

			lines := []string{}
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			if len(lines) != len(tcase.output) {
				t.Fatalf("wanted[%s] got[%s]", tcase.output, lines)
			}

			for i, want := range tcase.output {
				got := lines[i]
				if want != got {
					t.Errorf("unexpected line at[%d]", i)
					t.Errorf("wanted[%s] got[%s]", want, got)
					t.Errorf("wantedLines[%s] gotLines[%s]", tcase.output, lines)
				}
			}

			if scanner.Err() != nil {
				t.Fatal("unexpected error[%s]", scanner.Err())
			}
		})
	}
}

func TestStringsReadErrorOnFirstByte(t *testing.T) {
	var minWordSize uint = 1
	scanner := strings.Do(newFakeReader(func(d []byte) (int, error) {
		return 0, errors.New("fake injected error")
	}), minWordSize)
	assertScannerFails(t, scanner, 0)
}

func TestStringsReadErrorOnSecondByte(t *testing.T) {
	var minWordSize uint = 1
	sentFirstByte := false
	scanner := strings.Do(newFakeReader(func(d []byte) (int, error) {
		if sentFirstByte {
			return 0, errors.New("fake injected error")
		}
		sentFirstByte = true
		return 1, nil
	}), minWordSize)
	assertScannerFails(t, scanner, 1)
}

func TestStringsReadErrorAfterValidUTF9StartingByte(t *testing.T) {
	var minWordSize uint = 1
	sentFirstByte := false
	scanner := strings.Do(newFakeReader(func(d []byte) (int, error) {
		if sentFirstByte {
			return 0, errors.New("fake injected error")
		}
		sentFirstByte = true
		d[0] = runestart
		return 1, nil
	}), minWordSize)
	assertScannerFails(t, scanner, 1)
}

func TestStringsReadCanReturnEOFWithData(t *testing.T) {
	var minWordSize uint = 1
	want := byte('k')

	scanner := strings.Do(newFakeReader(func(d []byte) (int, error) {
		if len(d) == 0 {
			t.Fatal("empty data on Read operation")
		}
		d[0] = want
		return 1, io.EOF
	}), minWordSize)

	if !scanner.Scan() {
		t.Fatal("unexpected Scan failure")
	}
	got := scanner.Text()
	if string(want) != got {
		t.Fatalf("want[%s] != got[%s]", string(want), got)
	}
}

func TestStringReadErrors(t *testing.T) {

	type readresult struct {
		data byte
		err  error
	}

	type errorcase struct {
		name    string
		results []readresult
	}

	errorcases := []errorcase{
		{
			name: "NoReadAfterError",
			results: []readresult{
				{
					err: errors.New("injected error"),
				},
			},
		},
		{
			name: "NoReadAfterEOF",
			results: []readresult{
				{
					err: io.EOF,
				},
			},
		},
		{
			name: "NoReadAfterErrorSearchingForUTF",
			results: []readresult{
				{data: runestart},
				{err: errors.New("surprise moderfocker")},
			},
		},
	}

	for _, tcase := range errorcases {
		t.Run(tcase.name, func(t *testing.T) {
			var minWordSize uint = 1
			var fatalerr error

			calls := 0

			scanner := strings.Do(newFakeReader(func(d []byte) (int, error) {
				if fatalerr != nil {
					return 0, fatalerr
				}
				calls += 1

				if calls > len(tcase.results) {
					fatalerr = fmt.Errorf(
						"expected [%d] reads but got[%d]",
						len(tcase.results),
						calls,
					)
					return 0, fatalerr
				}
				written := 0
				result := tcase.results[calls-1]
				if result.data != 0 && len(d) > 0 {
					d[0] = result.data
					written = 1
				}
				return written, result.err
			}), minWordSize)

			for scanner.Scan() {
			}

			if fatalerr != nil {
				t.Fatal(fatalerr.Error())
			}
			if calls != len(tcase.results) {
				t.Fatalf("expected [%d] Read calls, got[%d]", len(tcase.results), calls)
			}
		})
	}
}

const runestart byte = 0xC2

type FakeReader struct {
	read func([]byte) (int, error)
}

func (f *FakeReader) Read(d []byte) (int, error) {
	if f.read == nil {
		return 0, fmt.Errorf("FakeReader has no Read implementation")
	}
	return f.read(d)
}

func newFakeReader(read func([]byte) (int, error)) *FakeReader {
	return &FakeReader{read: read}
}

func assertScannerFails(t *testing.T, scanner *bufio.Scanner, expectedIter uint) {
	var iterations uint
	for scanner.Scan() {
		iterations += 1
	}

	if iterations != expectedIter {
		t.Fatalf("expected[%d] Scan calls, got [%d]", expectedIter, iterations)
	}

	if scanner.Err() == nil {
		t.Fatal("expected failure on scanner, got none")
	}
}

func newBinary(size uint) []byte {
	// WHY: Starting with the most significant bit as 1 helps to test
	// UTF-8 corner cases. Don't change this without providing
	// testing for this. Not the best way to do this (not explicit)
	// but it is what we have for today =).
	bin := make([]byte, size)
	for i := 0; i < int(size); i++ {
		bin[i] = 0xFF
	}
	return bin
}
