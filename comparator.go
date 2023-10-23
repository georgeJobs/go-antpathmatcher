package antpathmatcher

// @Author :George
// @File: comparator
// @Version: 1.0.0
// @Date 2023/10/10 12:14

type Comparator interface {
	Compare(string, string) int
}
