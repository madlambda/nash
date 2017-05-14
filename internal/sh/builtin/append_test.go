package builtin_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash/internal/sh"
)

type testcase struct {
	name           string
	code           string
	expectedErr    string
	expectedStdout string
	expectedStderr string
}

func testAppend(t *testing.T, tc testcase) {
	sh, err := sh.NewShell()
	if err != nil {
		t.Fatal(err)
	}

	var (
		outb, errb bytes.Buffer
	)
	sh.SetStdout(&outb)
	sh.SetStderr(&errb)

	err = sh.Exec(tc.name, tc.code)
	stdout := string(outb.Bytes())
	stderr := errb.String()

	if stdout != tc.expectedStdout {
		t.Errorf("String differs: '%s' != '%s'", tc.expectedStdout, stdout)
		return
	}
	if stderr != tc.expectedStderr {
		t.Errorf("String differs: '%s' != '%s'", tc.expectedStderr, stderr)
		return
	}

	if err != nil {
		if err.Error() != tc.expectedErr {
			t.Fatalf("Expected err '%s' but got '%s'", tc.expectedErr, err.Error())
		}
	} else if tc.expectedErr != "" {
		t.Fatalf("Expected err '%s' but err is nil", tc.expectedErr)
	}
}

func TestAppend(t *testing.T) {
	for _, tc := range []testcase{
		{
			name:        "no argument fails",
			code:        `append()`,
			expectedErr: "<interactive>:1:0: append expects at least two arguments",
		},
		{
			name:        "one argument fails",
			code:        `append("1")`,
			expectedErr: "<interactive>:1:0: append expects at least two arguments",
		},
		{
			name: "simple append",
			code: `a = ()
		 a <= append($a, "hello")
		 a <= append($a, "world")
		 echo -n $a...`,
			expectedErr:    "",
			expectedStdout: "hello world",
			expectedStderr: "",
		},
		{
			name: "append is for lists",
			code: `a = "something"
		 a <= append($a, "other")
		 echo -n $a...`,
			expectedErr: "<interactive>:2:8: append expects a " +
				"list as first argument, but a StringType was provided",
			expectedStdout: "",
			expectedStderr: "",
		},
		{
			name: "var args",
			code: `a <= append((), "1", "2", "3", "4", "5", "6")
				echo -n $a...`,
			expectedErr:    "",
			expectedStdout: "1 2 3 4 5 6",
			expectedStderr: "",
		},
		{
			name: "append of lists",
			code: `a <= append((), (), ())
				if len($a) != "2" {
					print("wrong")
				} else if len($a[0]) != "0" {
					print("wrong")
				} else if len($a[1]) != "0" {
					print("wrong")
				} else { print("ok") }`,
			expectedErr:    "",
			expectedStdout: "ok",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testAppend(t, tc)
		})
	}

}
