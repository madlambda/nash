package sh

import "testing"

func TestFunctionsClosures(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "simpleClosure",
			execStr: `
				fn func(a) {
					fn closure() {
						print($a)
					}
					return $closure
				}

				x <= func("1")
				y <= func("2")
				$x()
				$y()
			`,
			expectedStdout: "12",
		},
		{
			desc: "eachCallCreatesNewVar",
			execStr: `
				fn func() {
					a = ()
					fn add(elem) {
						a <= append($a, $elem)
						print("a:%s,",$a)
					}
					return $add
				}

				add <= func()
				$add("1")
				$add("3")
				$add("5")
			`,
			expectedStdout: "a:1,a:3,a:5,",
		},
		{
			desc: "adder example",
			execStr: `
fn makeAdder(x) {
    fn add(y) {
        ret <= expr $x "+" $y
        return $ret
    }
    return $add
}

add1 <= makeAdder("1")
add5 <= makeAdder("5")
add1000 <= makeAdder("1000")

print("%s\n", add5("5"))
print("%s\n", add5("10"))
print("%s\n", add1("10"))
print("%s\n", add1("2"))
print("%s\n", add1000("50"))
print("%s\n", add1000("100"))
print("%s\n", add1("10"))
`,
			expectedStdout: `10
15
11
3
1050
1100
11
`,
		},
		{
			desc: "logger",
			execStr: `fn getlogger(prefix) {
    fn log(fmt, args...) {
        print($prefix+$fmt+"\n", $args...)
    }

    return $log
}

info <= getlogger("[info] ")
error <= getlogger("[error] ")
warn <= getlogger("[warn] ")

$info("nuke initialized successfully")
$warn("temperature above anormal circunstances: %s°", "870")
$error("about to explode...")
`,
			expectedStdout: `[info] nuke initialized successfully
[warn] temperature above anormal circunstances: 870°
[error] about to explode...
`,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}

func TestFunctionsVariables(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "fn stored only as vars",
			execStr: `
				fn func(a) {
					echo -n "hello"
				}

				func = "teste"
				echo -n $func
				func()
			`,
			expectedStdout: "teste",
			expectedErr:    "<interactive>:8:4: Identifier 'func' is not a function",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}

func TestFunctionsStateless(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "functions have no shared state",
			execStr: `fn iter(first, last, func) {
   sequence <= seq $first $last
   range <= split($sequence, "\n")
   for i in $range {
       $func($i)
   }
}

fn create_vm(index) {
	echo "create_vm: "+$index
	iter("1", "3", $create_disk)
}

fn create_disk(index) {
	echo "create_disk: " + $index
}

iter("1", "2", $create_vm)
`,
			expectedStdout: `create_vm: 1
create_disk: 1
create_disk: 2
create_disk: 3
create_vm: 2
create_disk: 1
create_disk: 2
create_disk: 3
`,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}
