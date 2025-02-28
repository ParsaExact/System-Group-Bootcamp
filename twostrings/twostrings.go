package twostrings
import "unicode/utf8"
const (
    Invlaid   = -1
    None      = 0
    Equal     = 1
    Prefix    = 2
    Suffix    = 3
    Substring = 4
)

func prefix(s1 string, s2 string) bool {
	if len(s2) > len(s1) { 
		return false
    }
	for i := 0; i < len(s2); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func suffix(s1 string, s2 string) bool {
	return len(s1) >= len(s2) && s1[len(s1)-len(s2):] == s2
}

func contains(s1 string, s2 string) bool {
	for i := 0; i < len(s1); i++ {
		if prefix(s1[i:], s2) {
			return true
		}
	}
	return false
}

func Process(s1 string, s2 string) int {
	if utf8.RuneCountInString(s1) == 0 || utf8.RuneCountInString(s2) == 0 || utf8.RuneCountInString(s2) > utf8.RuneCountInString(s1) {
        return Invlaid
    }else if s1 == s2 {
		return Equal
	}else if prefix(s1, s2) {
		return Prefix
	}else if suffix(s1, s2) {
		return Suffix
	}else if contains(s1, s2) {
		return Substring
	}else {
		return None
	}
}
	
