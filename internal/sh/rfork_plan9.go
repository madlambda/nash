// +build plan9

package sh

import (
	"fmt"
	"syscall"
)

func (sh *Shell) executeRfork(rfork *RforkNode) error {
	return newError("Sorry. Plan9 rfork not implemented yet.")
}

// getflags converts to Plan9 flags
func getflags(flags string) (uintptr, error) {
	var (
		pflags uintptr
	)

	for i := 0; i < len(flags); i++ {
		switch flags[i] {
		case 'n':
			pflags |= syscall.RFNAMEG
		case 'N':
			pflags |= syscall.RFCNAMEG
		case 'e':
			pflags |= syscall.RFENVG
		case 'E':
			pflags |= syscall.RFCENVG
		case 's':
			pflags |= syscall.RFNOTEG
		case 'f':
			pflags |= syscall.RFFDG
		case 'F':
			pflags |= syscall.RFCFDG
		default:
			return 0, fmt.Errorf("Wrong rfork flag: %c", flags[i])
		}
	}

	if pflags == 0 {
		return 0, fmt.Errorf("Rfork requires some flag")
	}

	return pflags, nil
}
