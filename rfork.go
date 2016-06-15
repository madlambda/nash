// +build !linux,!plan9

//
package nash

func (sh *Shell) executeRfork(rfork *RforkNode) error {
	return newError("rfork only supported on Linux and Plan9")
}
