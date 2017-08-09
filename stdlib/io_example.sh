import io

if len($ARGS) == "2" {
        io_println($ARGS[1])
        exit("0")
}

io_println($ARGS[1], $ARGS[2])
