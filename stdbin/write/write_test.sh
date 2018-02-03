import "./test.sh"

setenv PATH = "./stdbin/write:/bin"

# FIXME: we need our mktemp
var nonExistentFile = "./here-be-dragons"

fn clean() {
    _, _ <= rm -f $nonExistentFile
}

clean()

var out, err, status <= write $nonExistentFile "hello"
assert("", $out, "standard out isnt empty")
assert("", $err, "standard err isnt empty")
assert("0", $status, "status is not success")

var content, status <= cat $nonExistentFile
assert("0", $status, "status is not success")
assert("hello", $content, "file content is wrong")

clean()
