# common test routines

fn fatal(msg) {
    print($msg)
    exit("1")
}

fn assert(expected, got, desc) {
    if $expected != $got {
        fatal(format("%s: FAILED. Expected[%s] but got[%s]\n", $desc, $expected, $got))
    }
}