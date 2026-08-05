package fsx

import "unicode"

// naturalCompare compares two strings the way a person would: runs of
// digits are compared numerically rather than character-by-character, so
// "file9" sorts before "file10". Ties fall back to a plain byte comparison
// of the remaining text.
func naturalCompare(a, b string) int {
	return naturalCompareCase(a, b, true)
}

// naturalCompareFold is naturalCompare but case-insensitive.
func naturalCompareFold(a, b string) int {
	return naturalCompareCase(a, b, false)
}

func naturalCompareCase(a, b string, caseSensitive bool) int {
	ra, rb := []rune(a), []rune(b)
	i, j := 0, 0

	for i < len(ra) && j < len(rb) {
		ca, cb := ra[i], rb[j]

		if unicode.IsDigit(ca) && unicode.IsDigit(cb) {
			starti, startj := i, j
			for i < len(ra) && unicode.IsDigit(ra[i]) {
				i++
			}
			for j < len(rb) && unicode.IsDigit(rb[j]) {
				j++
			}

			numA := trimLeadingZeros(string(ra[starti:i]))
			numB := trimLeadingZeros(string(rb[startj:j]))

			if len(numA) != len(numB) {
				if len(numA) < len(numB) {
					return -1
				}
				return 1
			}
			if numA != numB {
				if numA < numB {
					return -1
				}
				return 1
			}
			continue
		}

		if !caseSensitive {
			ca = unicode.ToLower(ca)
			cb = unicode.ToLower(cb)
		}

		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
		i++
		j++
	}

	switch {
	case len(ra)-i < len(rb)-j:
		return -1
	case len(ra)-i > len(rb)-j:
		return 1
	default:
		return 0
	}
}

func trimLeadingZeros(s string) string {
	i := 0
	for i < len(s)-1 && s[i] == '0' {
		i++
	}
	return s[i:]
}
