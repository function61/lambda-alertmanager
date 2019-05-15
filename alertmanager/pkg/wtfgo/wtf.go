package wtfgo

// wtfgo: Go, why the fuck do we need to write these ourselves??!

// https://stackoverflow.com/questions/28718682/how-to-get-a-substring-from-a-string-of-runes-in-golang/56129287#56129287
// https://stackoverflow.com/questions/12311033/extracting-substrings-in-go/56129336#56129336
func Substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

func Truncate(input string, to int) string {
	// we could optimize and do len(input) here but it operates on bytes (not runes), so
	// we could have some unexpected behaviour with input getting suffixed ... when no
	// truncation actually happens
	truncated := Substr(input, 0, to)

	if truncated != input {
		return truncated + "..."
	}

	return input
}

// https://stackoverflow.com/questions/27516387/what-is-the-correct-way-to-find-the-min-between-two-integers-in-go
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
