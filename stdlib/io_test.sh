
fn run_example(args...) {
        var got, status <= ./cmd/nash/nash ./stdlib/io_example.sh $args
        return $got, $status
}

fn assert_success(expected, got, status) {
        if $status != "0" {
                print("expected success, but got status code: %s\n", $status)
                exit("1")
        }
        if $got != $expected {
                print("expected [%s] got [%s]\n", $expected, $got)
                exit("1")
        }
}

fn test_println_format() {
        var got, status <= run_example("hello %s", "world")

        assert_success("hello world", $got, $status)
}

fn test_println() {
        var expected = "pazu"
        var got, status <= run_example($expected)

        assert_success($expected, $got, $status)
}

test_println_format()
test_println()
