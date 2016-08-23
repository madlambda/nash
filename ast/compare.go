package ast

func cmpCommon(expected, value Node) bool {
	if expected == value {
		return true
	}

	if expected.Position() != value.Position() {
		debug("Nodes positions arent equal... expected %d, but %d", expected.Position(), value.Position())
		return false
	}

	return true
}
