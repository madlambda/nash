fn test_println() {
        got, status <= ./cmd/nash/nash ./stdlib/fmt_example.sh "hello %s" "world"
        if $status != "0" {
                exit("1")
        }

        expected = "hello world"
        if $got != $expected {
                print("expected [%s] got [%s]\n", $expected, $got)
                exit("1")
        }
}

test_println()
