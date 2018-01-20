package strings_test

import (
	"bytes"
	stdstrings "strings"
	"testing"

	"github.com/NeowayLabs/nash/stdbin/strings"
)

// TODO:
//func TestBinaryEndingWithText(t *testing.T) {
//}

//func TestBinaryWithTextOnMiddle(t *testing.T) {
//}

//func TestMinTextSizeIsAdjustable(t *testing.T) {
//}

//func TestEachTextOccurenceIsANewLine(t *testing.T) {
//}

//func TestJustText(t *testing.T) {
//}

//func TestJustBinary(t *testing.T) {
//}

func TestBinaryWithText(t *testing.T) {

	tcases := []testcase{
		testcase{
			name:        "startingWithText",
			minWordSize: 1,
			input: func() []byte {
				expected := "textOnBeggining"
				bin := newBinary(512)
				return append([]byte(expected), bin...)
			},
			output: "textOnBeggining",
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

			assertStrings(t, tcase.output, stdstrings.Join(lines, "\n"))
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
	output      string
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

// TODO: Start an assert package on nash or use our other project ?
func assertTrue(t *testing.T, b bool, msg string) {
	t.Helper()
	if !b {
		t.Fatalf("want true, got false: %s", msg)
	}
}

func assertFalse(t *testing.T, b bool, msg string) {
	t.Helper()
	if b {
		t.Fatalf("want false, got true: %s", msg)
	}
}

func assertStrings(t *testing.T, want string, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("want[%s] != got[%s]", want, got)
	}
}
