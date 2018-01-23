package strings_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash/stdbin/strings"
)

// TODO
// Test fatal error on input io.Reader

func TestStrings(t *testing.T) {

	tcases := []testcase{
		testcase{
			name:        "UTF-8With2Bytes",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("λ"), bin...)
			},
			output: []string{"λ"},
		},
		testcase{
			name:        "UTF-8With3Bytes",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("€"), bin...)
			},
			output: []string{"€"},
		},
		testcase{
			name:        "UTF-8With4Bytes",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("𐍈"), bin...)
			},
			output: []string{"𐍈"},
		},
		testcase{
			name:        "NonASCIIWordHasOneLessCharThanMin",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("λ"), bin...)
			},
			output: []string{},
		},
		testcase{
			name:        "NonASCIIWordHasMinWordSize",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("λλ"), bin...)
			},
			output: []string{"λλ"},
		},
		testcase{
			name:        "WordHasOneLessCharThanMin",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("k"), bin...)
			},
			output: []string{},
		},
		testcase{
			name:        "WordHasMinWordSize",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("kz"), bin...)
			},
			output: []string{"kz"},
		},
		testcase{
			name:        "WordHasOneMoreCharThanMinWordSize",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("ktz"), bin...)
			},
			output: []string{"ktz"},
		},
		testcase{
			name:        "StartingWithOneChar",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append([]byte("k"), bin...)
			},
			output: []string{"k"},
		},
		testcase{
			name:        "EndWithOneChar",
			minWordSize: 1,
			input: func() []byte {
				bin := newBinary(64)
				return append(bin, []byte("k")...)
			},
			output: []string{"k"},
		},
		testcase{
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
		testcase{
			name:        "StartingWithText",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnBeggining"
				bin := newBinary(64)
				return append([]byte(expected), bin...)
			},
			output: []string{"textOnBeggining"},
		},
		testcase{
			name:        "TextOnMiddle",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnMiddle"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"textOnMiddle"},
		},
		testcase{
			name:        "NonASCIITextOnMiddle",
			minWordSize: 1,
			input: func() []byte {
				expected := "λλλ"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"λλλ"},
		},
		testcase{
			name:        "ASCIIAndNonASCII",
			minWordSize: 1,
			input: func() []byte {
				expected := "(define (λ (x) (+ x a)))"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"(define (λ (x) (+ x a)))"},
		},
		testcase{
			name:        "TextOnEnd",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnEnd"
				bin := newBinary(64)
				return append(bin, append([]byte(expected), bin...)...)
			},
			output: []string{"textOnEnd"},
		},
		testcase{
			name:        "JustText",
			minWordSize: 1,
			input: func() []byte {
				return []byte("justtext")
			},
			output: []string{"justtext"},
		},
		testcase{
			name:        "JustBinary",
			minWordSize: 1,
			input: func() []byte {
				return newBinary(64)
			},
			output: []string{},
		},
		testcase{
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
		testcase{
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
		testcase{
			name:        "WordsAreNotAccumulativeBetweenBinData",
			minWordSize: 2,
			input: func() []byte {
				bin := newBinary(64)
				t := append([]byte("k"), bin...)
				return append(t, byte('t'))
			},
			output: []string{},
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

			if tcase.fail {
				if scanner.Err() == nil {
					t.Fatal("expected error, got nil")
				}
			}
		})
	}
}

type testcase struct {
	name        string
	input       func() []byte
	output      []string
	fail        bool
	minWordSize uint
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
