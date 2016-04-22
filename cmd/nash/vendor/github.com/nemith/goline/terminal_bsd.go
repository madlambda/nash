// +build darwin freebsd netbsd openbsd

package goline

import "syscall"

// Read and Write syscall operations on BSD platforms
const (
	ioctlReadTermios  = syscall.TIOCGETA
	ioctlWriteTermios = syscall.TIOCSETA
)
