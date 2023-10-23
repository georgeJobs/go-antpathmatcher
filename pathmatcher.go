package antpathmatcher

// @Author :George
// @File: pathmatcher
// @Version: 1.0.0
// @Date 2023/10/10 12:08

type PathMatcher interface {
	IsPattern(string) bool
	Match(string, string) bool
	MatchStart(string, string) bool
	ExtractPathWithinPattern(string, string) string
	ExtractUriTemplateVariables(string, string) map[string]string
	GetPatternComparator(string) Comparator
	Combine(string, string) string
}
