package goline

import "strings"

type History struct {
	history  [][]rune
	curIndex int
}

func (h *History) PreviousHistory(l *GoLine) (bool, error) {
	if h.curIndex > 0 {
		line := h.history[h.curIndex-1]
		l.CurLine = line[:MAX_LINE]
		l.Position = len(line)
		l.Len = len(line)
		h.curIndex--
	}
	return false, nil
}

func (h *History) NextHistory(l *GoLine) (bool, error) {
	if h.curIndex < len(h.history)-1 {
		line := h.history[h.curIndex+1]
		l.CurLine = line[:MAX_LINE]
		l.Position = len(line)
		l.Len = len(line)
		h.curIndex++
	} else if h.curIndex == len(h.history)-1 {
		return DeleteLine(l)
	}
	return false, nil
}

func (h *History) AddLine(line []rune) {
	if strings.Trim(string(line), " ") != "" {
		h.history = append(h.history, line)
		h.curIndex = len(h.history)
	}
}

func (h *History) HistoryFinish(l *GoLine) (bool, error) {
	h.AddLine(l.CurLine[:l.Len])
	return Finish(l)
}

func SetupHistory(l *GoLine) {
	h := History{}

	l.AddHandler(CHAR_CTRLP, h.PreviousHistory)
	l.AddHandler(CHAR_CTRLN, h.NextHistory)
	l.AddHandler(ESCAPE_UP, h.PreviousHistory)
	l.AddHandler(ESCAPE_DOWN, h.NextHistory)

	// Overwrite any previous definition
	l.AddHandler(CHAR_ENTER, h.HistoryFinish)
}
