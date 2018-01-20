package strings_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash/stdbin/strings"
)

// TODO

// Test fatal error on input io.Reader

// Test word size is relative to rune not bytes (practice some chinese)

//func TestMinTextSizeIsAdjustable(t *testing.T) {
//}

func TestStrings(t *testing.T) {

	tcases := []testcase{
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
	minWordSize int
}

func newBinary(size uint) []byte {
	// TODO: Not the most awesome random binary =/
	bin := make([]byte, size)
	for i := 0; i < int(size); i++ {
		bin[i] = 0xFF
	}
	return bin
}
