
fn run_example(args...) {
        got, status <= ./cmd/nash/nash ./stdlib/fmt_example.sh $args
        return $got, $status
}

fn assert_success(expected, got, status) {
        if $status != "0" {
                exit("1")
        }
        if $got != $expected {
                print("expected [%s] got [%s]\n", $expected, $got)
                exit("1")
        }
}

fn test_println_format() {
        got, status <= run_example("hello %s", "world")

        assert_success("hello world", $got, $status)
}

fn test_println() {
        expected = "pazu"
        got, status <= run_example($expected)

        assert_success($expected, $got, $status)
}

test_println_format()
test_println()
