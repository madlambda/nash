package goline

func findLastWord(t []rune, start int) int {
	for start > 0 && t[start-1] == ' ' {
		start--
	}
	for start > 0 && t[start-1] != ' ' {
		start--
	}
	return start
}

func findNextWord(t []rune, end, max int) int {
	for end < max && t[end] == ' ' {
		end++
	}
	for end < max && t[end] != ' ' {
		end++
	}
	return end
}

func Finish(l *GoLine) (bool, error) {
	return true, nil
}

func UserTerminated(l *GoLine) (bool, error) {
	return true, UserTerminatedError
}

func Backspace(l *GoLine) (bool, error) {
	if l.Len > 0 && l.Position > 0 {
		l.CurLine = append(l.CurLine[:l.Position-1], l.CurLine[l.Position:]...)
		l.Len--
		l.Position--
		l.CurLine[l.Len] = 0
	}
	return false, nil
}

func Tab(l *GoLine) (bool, error) {
	prev := l.CurLine[:l.Position]
	next := l.CurLine[l.Position:]

	l.CurLine = make([]rune, MAX_LINE)

	i := 0

	for i = 0; i < len(prev); i++ {
		l.CurLine[i] = prev[i]
	}

	l.CurLine[i] = ' '
	i++
	l.CurLine[i] = ' '
	i++
	l.CurLine[i] = ' '
	i++
	l.CurLine[i] = ' '

	i++
	j := 0

	for j = 0; j < len(next)-4; j++ {
		l.CurLine[i+j] = next[j]
	}

	l.Len += 1
	l.Position += 1
	l.CurLine[l.Len] = 0

	return false, nil
}

func MoveBackOneWord(l *GoLine) (bool, error) {
	l.Position = findLastWord(l.CurLine, l.Position)
	return false, nil
}

func MoveForwardOneWord(l *GoLine) (bool, error) {
	l.Position = findNextWord(l.CurLine, l.Position, l.Len)
	return false, nil
}

func MoveLeft(l *GoLine) (bool, error) {
	if l.Position > 0 {
		l.Position--
	}
	return false, nil
}

func MoveRight(l *GoLine) (bool, error) {
	if l.Position != l.Len {
		l.Position++
	}
	return false, nil
}

func DeleteLine(l *GoLine) (bool, error) {
	l.CurLine = make([]rune, MAX_LINE)
	l.Position = 0
	l.Len = 0
	return false, nil
}

func DeleteRestofLine(l *GoLine) (bool, error) {
	copy(l.CurLine, l.CurLine[:l.Position])
	l.Len = l.Position
	return false, nil
}

func DeleteLastWord(l *GoLine) (bool, error) {
	prev_position := l.Position
	l.Position = findLastWord(l.CurLine, l.Position)
	copy(l.CurLine, append(l.CurLine[:l.Position], l.CurLine[prev_position:l.Len]...))
	l.Len -= prev_position - l.Position
	return false, nil
}

func DeleteNextWord(l *GoLine) (bool, error) {
	end := findNextWord(l.CurLine, l.Position, l.Len)
	copy(l.CurLine, append(l.CurLine[:l.Position], l.CurLine[end:l.Len]...))
	l.Len -= end - l.Position
	return false, nil
}

func DeleteCurrentChar(l *GoLine) (bool, error) {
	if l.Position < l.Len {
		l.CurLine = append(l.CurLine[:l.Position], l.CurLine[l.Position+1:]...)
		l.Len--
	}
	return false, nil
}

func SwapWithPreviousChar(l *GoLine) (bool, error) {
	if l.Position > 0 && l.Position <= l.Len {
		x := l.Position
		if l.Position == l.Len {
			x--
		}
		l.CurLine[x], l.CurLine[x-1] = l.CurLine[x-1], l.CurLine[x]
		if l.Position != l.Len {
			l.Position++
		}
	}
	return false, nil
}

func MoveStartofLine(l *GoLine) (bool, error) {
	l.Position = 0
	return false, nil
}

func MoveEndofLine(l *GoLine) (bool, error) {
	l.Position = l.Len
	return false, nil
}

func ClearScreen(l *GoLine) (bool, error) {
	l.ClearScreen()
	return false, nil
}
