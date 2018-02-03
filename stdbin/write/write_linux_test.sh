# linux tests of write command

import "./test.sh"

# this test uses only the write binary
setenv PATH = "./stdbin/write"

# (desc (out err status))
var tests = (
    ("standard out" ("/dev/stdout" "hello world" "" "0"))
    ("standard err" ("/dev/stderr" "" "hello world" "0"))
)

var outstr = "hello world"

for test in $tests {
    var desc = $test[0]
    var tc = $test[1]

    print("testing %s\n", $desc)

    var device = $tc[0]
    var expectedOut = $tc[1]
    var expectedErr = $tc[2]
    var expectedSts = $tc[3]

    var out, err, status <= write $device $outstr
    assert($expectedSts, $status, "status code")
    assert($expectedOut, $out, "standard output")
    assert($expectedErr, $err, "standard error")
}
