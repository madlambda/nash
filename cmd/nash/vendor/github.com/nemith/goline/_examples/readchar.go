package main

import (
	"fmt"
	"github.com/nemith/goline"
	"syscall"
)

func main() {
	tty, _ := goline.NewTty(syscall.Stdin)

	tty.EnableRawMode()
	defer tty.DisableRawMode()

	for {
		c, _ := tty.ReadRune()
		switch c {
		case goline.CHAR_CTRLC:
			return
		default:
			tty.Write([]byte(fmt.Sprintf("Char: %c (%d) [0x%x]\r\n", c, c, c)))
		}
	}
}
