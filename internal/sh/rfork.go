// +build !linux,!plan9

//
package sh

func (sh *Shell) executeRfork(rfork *RforkNode) error {
	return newError("rfork only supported on Linux and Plan9")
}
