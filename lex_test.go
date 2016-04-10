package cnt

import "fmt"
import "testing"

func testTable(name, content string, expected []item, t *testing.T) {
	l := lex(name, content)

	if l == nil {
		t.Errorf("Failed to initialize lexer")
		return
	}

	if l.items == nil {
		t.Errorf("Failed to initialize lexer")
		return
	}

	result := make([]item, 0, 1024)

	for i := range l.items {
		result = append(result, i)
	}

	if len(result) != len(expected) {
		t.Errorf("Failed to parse commands, length differs %d != %d",
			len(result), len(expected))
		fmt.Printf("%+v\n", result)
		return
	}

	for i := 0; i < len(expected); i++ {
		if expected[i].typ != result[i].typ {
			t.Errorf("'%d' != '%d'", expected[i].typ, result[i].typ)
			fmt.Printf("Type: %d - %s\n", result[i].typ, result[i])
			return
		}

		if expected[i].val != result[i].val {
			t.Errorf("'%s' != '%s'", expected[i].val, result[i].val)
			return
		}
	}
}

func TestShebangOnly(t *testing.T) {
	expected := []item{
		item{
			typ: itemComment,
			val: "#!/bin/cnt",
		},
	}

	testTable("testShebangonly", "#!/bin/cnt\n", expected, t)
}

func TestSimpleCommand(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemString,
			val: "hello world",
		},
	}

	testTable("testSimpleCommand", `echo "hello world"`, expected, t)
}

func TestVariousCommands(t *testing.T) {
	content := `
            echo "hello world"
            mount -t proc proc /proc
        `

	expected := []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemString,
			val: "hello world",
		},
		item{
			typ: itemCommand,
			val: "mount",
		},
		item{
			typ: itemArg,
			val: "-t",
		},
		item{
			typ: itemArg,
			val: "proc",
		},
		item{
			typ: itemArg,
			val: "proc",
		},
		item{
			typ: itemArg,
			val: "/proc",
		},
	}

	testTable("testVariouscommands", content, expected, t)
}

func TestRfork(t *testing.T) {
	expected := []item{
		item{
			typ: itemRfork,
			val: "rfork",
		},
		item{
			typ: itemRforkFlags,
			val: "u",
		},
	}

	testTable("testRfork", "rfork u\n", expected, t)

	expected = []item{
		item{
			typ: itemRfork,
			val: "rfork",
		},
		item{
			typ: itemRforkFlags,
			val: "usnm",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemString,
			val: "inside namespace :)",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
	}

	testTable("testRforkWithBlock", `
            rfork usnm {
                echo "inside namespace :)"
            }
        `, expected, t)
}
