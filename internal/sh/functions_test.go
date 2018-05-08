package sh_test

import "testing"

func TestFunctionsClosures(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "simpleClosure",
			code: `
				fn func(a) {
					fn closure() {
						print($a)
					}
					return $closure
				}

				var x <= func("1")
				var y <= func("2")
				$x()
				$y()
			`,
			expectedStdout: "12",
		},
		{
			desc: "eachCallCreatesNewVar",
			code: `
				fn func() {
					var a = ()
					fn add(elem) {
						a <= append($a, $elem)
						print("a:%s,",$a)
					}
					return $add
				}

				var add <= func()
				$add("1")
				$add("3")
				$add("5")
			`,
			expectedStdout: "a:1,a:1 3,a:1 3 5,",
		},
		{
			desc: "adder example",
			code: `
fn makeAdder(x) {
    fn add(y) {
        var ret <= expr $x "+" $y
        return $ret
    }
    return $add
}

var add1 <= makeAdder("1")
var add5 <= makeAdder("5")
var add1000 <= makeAdder("1000")

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
			code: `fn getlogger(prefix) {
    fn log(fmt, args...) {
        print($prefix+$fmt+"\n", $args...)
    }

    return $log
}

var info <= getlogger("[info] ")
var error <= getlogger("[error] ")
var warn <= getlogger("[warn] ")

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
			code: `
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
			code: `fn iter(first, last, func) {
   var sequence <= seq $first $last
   var range <= split($sequence, "\n")
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
