package cnt

import "testing"

func TestParseSimple(t *testing.T) {
	parser := NewParser("parser simple", `

            echo "hello world"
        `)

	tr, err := parser.Parse()

	if err != nil {
		t.Error(err)
		return
	}

	if tr == nil {
		t.Errorf("Failed to parse")
		return
	}
}
