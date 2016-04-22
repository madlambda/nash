// +build linux

package goline

import "syscall"

// Read and Write syscall operations on Linux platforms
const (
	ioctlReadTermios  = syscall.TCGETS
	ioctlWriteTermios = syscall.TCSETS
)
