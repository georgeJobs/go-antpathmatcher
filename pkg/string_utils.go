package pkg

import (
	"strings"
	"unicode"
)

// @Author :George
// @File: string_utils
// @Version: 1.0.0
// @Date 2023/10/10 17:12

func TokenizeToStringArray(str, delimiters string, trimTokens, ignoreEmptyTokens bool) []string {
	if strings.TrimSpace(str) == "" {
		return []string{}
	} else {
		l := strings.FieldsFunc(str, func(a rune) bool {
			return strings.Contains(delimiters, string(a))
		})
		tokens := make([]string, 0)
		for k := range l {
			if trimTokens {
				l[k] = strings.TrimSpace(l[k])
			}
			if !ignoreEmptyTokens || l[k] != "" {
				tokens = append(tokens, l[k])
			}
		}
		return tokens
	}
}

func HasText(str string) bool {
	return strings.TrimSpace(str) != "" && containsText([]rune(str))
}

func containsText(str []rune) bool {
	for i := 0; i < len(str); i++ {
		if !unicode.IsSpace(str[i]) {
			return true
		}
	}
	return false
}
