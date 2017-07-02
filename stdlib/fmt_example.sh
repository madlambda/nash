import fmt

if len($ARGS) == "2" {
        fmt_println($ARGS[1])
        exit("0")
}

fmt_println($ARGS[1], $ARGS[2])
