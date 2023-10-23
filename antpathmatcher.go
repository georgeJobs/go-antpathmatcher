package antpathmatcher

import (
	"bytes"
	"fmt"
	"github.com/georgeJobs/go-antpathmatcher/pkg"
	"gopkg.in/guregu/null.v3"
	"regexp"
	"strings"
)

// @Author :George
// @File: antpathmatcher
// @Version: 1.0.0
// @Date 2023/10/10 12:16

const DEFAULT_PATH_SEPARATOR = "/"
const CACHE_TURNOFF_THRESHOLD = 65536

var WILDCARD_CHARS = [3]byte{'*', '?', '{'}
var VARIABLE_PATTERN = regexp.MustCompile("\\{[^/]+?\\}")

//region AntPathMatcher

type AntPathMatcher struct {
	caseSensitive             bool
	trimTokens                bool
	cachePatterns             null.Bool
	pathSeparator             string
	pathSeparatorPatternCache *PathSeparatorPatternCache
	tokenizedPatternCache     pkg.MySyncMap
	stringMatcherCache        pkg.MySyncMap
}

func NewAntPathMatcher() *AntPathMatcher {
	return NewAntPathMatcherWithPathSeparator(DEFAULT_PATH_SEPARATOR)
}

func NewAntPathMatcherWithPathSeparator(pathSeparator string) *AntPathMatcher {
	if strings.TrimSpace(pathSeparator) == "" {
		pathSeparator = DEFAULT_PATH_SEPARATOR
	}
	return &AntPathMatcher{
		caseSensitive:             true,
		trimTokens:                false,
		pathSeparator:             pathSeparator,
		pathSeparatorPatternCache: NewPathSeparatorPatternCache(pathSeparator),
		tokenizedPatternCache:     pkg.MySyncMap{},
		stringMatcherCache:        pkg.MySyncMap{},
	}
}

func (a *AntPathMatcher) SetCachePatterns(cachePatterns bool) {
	a.cachePatterns.SetValid(cachePatterns)
}

func (a *AntPathMatcher) IsPattern(path string) bool {
	if strings.TrimSpace(path) == "" {
	} else {
		if strings.Contains(path, "*") || strings.Contains(path, "?") {
			return true
		}
		s := strings.IndexByte(path, '{')
		e := strings.IndexByte(path, '}')
		if s > -1 && e > -1 && s < e {
			return true
		}
	}
	return false
}
func (a *AntPathMatcher) Match(pattern, path string) bool {
	return a.doMatch(pattern, path, true, nil)
}
func (a *AntPathMatcher) MatchStart(pattern, path string) bool {
	return a.doMatch(pattern, path, false, nil)
}
func (a *AntPathMatcher) ExtractPathWithinPattern(pattern, path string) string {
	patternParts := pkg.TokenizeToStringArray(pattern, a.pathSeparator, a.trimTokens, true)
	pathParts := pkg.TokenizeToStringArray(path, a.pathSeparator, a.trimTokens, true)
	builder := bytes.NewBufferString("")
	pathStarted := false
	for segment := 0; segment < len(patternParts); segment++ {
		patternPart := patternParts[segment]
		if strings.IndexByte(patternPart, '*') > -1 || strings.IndexByte(patternPart, '?') > -1 {
			for ; segment < len(pathParts); segment++ {
				if pathStarted || segment == 0 && !strings.HasPrefix(pattern, a.pathSeparator) {
					builder.WriteString(a.pathSeparator)
				}
				builder.WriteString(pathParts[segment])
				pathStarted = true
			}
		}
	}
	return builder.String()
}
func (a *AntPathMatcher) ExtractUriTemplateVariables(pattern, path string) map[string]string {
	variables := make(map[string]string)
	result := a.doMatch(pattern, path, true, variables)
	if !result {
		panic("Pattern \"" + pattern + "\" is not a match for \"" + path + "\"")
	}
	return variables
}

func (a *AntPathMatcher) concat(path1, path2 string) string {
	path1EndsWithSeparator := strings.HasSuffix(path1, a.pathSeparator)
	path2StartsWithSeparator := strings.HasPrefix(path2, a.pathSeparator)

	if path1EndsWithSeparator && path2StartsWithSeparator {
		return path1 + path2[1:]
	} else if path1EndsWithSeparator || path2StartsWithSeparator {
		return path1 + path2
	} else {
		return path1 + a.pathSeparator + path2
	}
}

func (a *AntPathMatcher) GetPatternComparator(path string) Comparator {
	return NewAntPatternComparator(path)
}
func (a *AntPathMatcher) Combine(pattern1, pattern2 string) string {
	if !pkg.HasText(pattern1) && !pkg.HasText(pattern2) {
		return ""
	}
	if !pkg.HasText(pattern1) {
		return pattern2
	}
	if !pkg.HasText(pattern2) {
		return pattern1
	}

	pattern1ContainsUriVar := strings.IndexByte(pattern1, '{') != -1
	if pattern1 != pattern2 && !pattern1ContainsUriVar && a.Match(pattern1, pattern2) {
		// /* + /hotel -> /hotel ; "/*.*" + "/*.html" -> /*.html
		// However /user + /user -> /usr/user ; /{foo} + /bar -> /{foo}/bar
		return pattern2
	}

	// /hotels/* + /booking -> /hotels/booking
	// /hotels/* + booking -> /hotels/booking
	if strings.HasSuffix(pattern1, a.pathSeparatorPatternCache.endsOnWildcard) {
		return a.concat(pattern1[:len(pattern1)-2], pattern2)
	}
	// /hotels/** + /booking -> /hotels/**/booking
	// /hotels/** + booking -> /hotels/**/booking
	if strings.HasSuffix(pattern1, a.pathSeparatorPatternCache.endsOnDoubleWildcard) {
		return a.concat(pattern1, pattern2)
	}

	starDotPos1 := strings.Index(pattern1, "*.")
	if pattern1ContainsUriVar || starDotPos1 == -1 || a.pathSeparator == "." {
		// simply concatenate the two patterns
		return a.concat(pattern1, pattern2)
	}

	ext1 := pattern1[starDotPos1+1:]
	dotPos2 := strings.IndexByte(pattern2, '.')
	file2 := pattern2
	ext2 := ""
	if dotPos2 > -1 {
		file2 = pattern2[:dotPos2]
		ext2 = pattern2[dotPos2:]
	}

	ext1All := ext1 == ".*" || len(ext1) == 0
	ext2All := ext2 == ".*" || len(ext2) == 0

	if !ext1All && !ext2All {
		panic("Cannot combine patterns: " + pattern1 + " vs " + pattern2)
	}
	ext := ext1
	if ext1All {
		ext = ext2
	}
	return file2 + ext
}

func (a *AntPathMatcher) doMatch(pattern, path string, fullMatch bool, uriTemplateVariables map[string]string) bool {
	//todo path is null
	if strings.HasPrefix(path, a.pathSeparator) != strings.HasPrefix(pattern, a.pathSeparator) {
		return false
	}
	pattDirs := a.tokenizePattern(pattern)
	if fullMatch && a.caseSensitive && !a.isPotentialMatch(path, pattDirs) {
		return false
	}

	pathDirs := a.tokenizePath(path)
	pattIdxStart, pattIdxEnd, pathIdxStart, pathIdxEnd := 0, len(pattDirs)-1, 0, len(pathDirs)-1

	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
		pattDir := pattDirs[pattIdxStart]
		if "**" == pattDir {
			break
		}
		if !a.matchStrings(pattDir, pathDirs[pathIdxStart], uriTemplateVariables) {
			return false
		}
		pattIdxStart++
		pathIdxStart++
	}

	if pathIdxStart > pathIdxEnd {
		// Path is exhausted, only match if rest of pattern is * or **'s
		if pattIdxStart > pattIdxEnd {
			return strings.HasSuffix(pattern, a.pathSeparator) == strings.HasSuffix(path, a.pathSeparator)
		}
		if !fullMatch {
			return true
		}
		if pattIdxStart == pattIdxEnd && pattDirs[pattIdxStart] == "*" && strings.HasSuffix(path, a.pathSeparator) {
			return true
		}
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if pattDirs[i] != "**" {
				return false
			}
		}
		return true
	} else if pattIdxStart > pattIdxEnd {
		// String not exhausted, but pattern is. Failure.
		return false
	} else if !fullMatch && "**" == pattDirs[pattIdxStart] {
		// Path start definitely matches due to "**" part in pattern.
		return true
	}

	// up to last '**'
	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
		pattDir := pattDirs[pattIdxEnd]
		if pattDir == "**" {
			break
		}
		if !a.matchStrings(pattDir, pathDirs[pathIdxEnd], uriTemplateVariables) {
			return false
		}
		if pattIdxEnd == len(pattDirs)-1 && strings.HasSuffix(pattern, a.pathSeparator) != strings.HasSuffix(path, a.pathSeparator) {
			return false
		}
		pattIdxEnd--
		pathIdxEnd--
	}
	if pathIdxStart > pathIdxEnd {
		// String is exhausted
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if pattDirs[i] != "**" {
				return false
			}
		}
		return true
	}
	for pattIdxStart != pattIdxEnd && pathIdxStart <= pathIdxEnd {
		patIdxTmp := -1
		for i := pattIdxStart + 1; i <= pattIdxEnd; i++ {
			if pattDirs[i] == "**" {
				patIdxTmp = i
				break
			}
		}
		if patIdxTmp == pattIdxStart+1 {
			// '**/**' situation, so skip one
			pattIdxStart++
			continue
		}
		// Find the pattern between padIdxStart & padIdxTmp in str between
		// strIdxStart & strIdxEnd
		patLength, strLength, foundIdx := patIdxTmp-pattIdxStart-1, pathIdxEnd-pathIdxStart+1, -1
	strLoop:
		for i := 0; i < strLength-patLength; i++ {
			for j := 0; j < patLength; j++ {
				subPat, subStr := pattDirs[pattIdxStart+j+1], pathDirs[pathIdxStart+i+j]
				if !a.matchStrings(subPat, subStr, uriTemplateVariables) {
					continue strLoop
				}
			}
			foundIdx = pathIdxStart + i
			break
		}
		if foundIdx == -1 {
			return false
		}
		pattIdxStart = patIdxTmp
		pathIdxStart = foundIdx + patLength
	}
	for i := pattIdxStart; i <= pattIdxEnd; i++ {
		if pattDirs[i] != "**" {
			return false
		}
	}

	return true
}
func (a *AntPathMatcher) isPotentialMatch(path string, pattDirs []string) bool {
	if !a.trimTokens {
		pos := 0
		for k := range pattDirs {
			skipped := skipSeparator(pos, path, a.pathSeparator)
			pos += skipped
			skipped = skipSegment(path, pos, []byte(pattDirs[k]))
			if skipped < len(pattDirs[k]) {
				return skipped > 0 || (len(pattDirs[k]) > 0 && isWildcardChar([]byte(pattDirs[k])[0]))
			}
			pos += skipped
		}
	}
	return true
}

func skipSegment(path string, pos int, prefix []byte) int {
	skipped := 0
	for i := 0; i < len(prefix); i++ {
		if isWildcardChar(prefix[i]) {
			return skipped
		}
		currPos := pos + skipped
		if currPos >= len(path) {
			return 0
		}
		if prefix[i] == []byte(path)[currPos] {
			skipped++
		}
	}
	return skipped
}

func skipSeparator(pos int, path, separator string) int {
	skipped := 0
	for strings.HasPrefix(path[pos+skipped:], separator) {
		skipped += len(separator)
	}
	return skipped
}

func isWildcardChar(c byte) bool {
	for a := range WILDCARD_CHARS {
		if WILDCARD_CHARS[a] == c {
			return true
		}
	}
	return false
}

func (a *AntPathMatcher) tokenizePattern(pattern string) []string {
	tokenized := make([]string, 0)
	cachePatterns := a.cachePatterns
	if !cachePatterns.Valid || cachePatterns.Bool {
		tmp, ok := a.tokenizedPatternCache.Load(pattern)
		if ok {
			tokenized, _ = tmp.([]string)
		}
	}
	if tokenized == nil || len(tokenized) == 0 {
		tokenized = a.tokenizePath(pattern)
		if !a.cachePatterns.Valid && a.tokenizedPatternCache.Len() > CACHE_TURNOFF_THRESHOLD {
			// Try to adapt to the runtime situation that we're encountering:
			// There are obviously too many different patterns coming in here...
			// So let's turn off the cache since the patterns are unlikely to be reoccurring.
			a.DeactivatePatternCache()
			return tokenized
		}
		if !a.cachePatterns.Valid || cachePatterns.Bool {
			a.tokenizedPatternCache.Store(pattern, tokenized)
		}
	}
	return tokenized
}

func (a *AntPathMatcher) tokenizePath(path string) []string {
	return pkg.TokenizeToStringArray(path, a.pathSeparator, a.trimTokens, true)
}

func (a *AntPathMatcher) matchStrings(pattern, str string, uriTemplateVariables map[string]string) bool {
	return a.getStringMatcher(pattern).matchStrings(str, uriTemplateVariables)
}

func (a *AntPathMatcher) getStringMatcher(pattern string) *AntPathStringMatcher {
	var matcher *AntPathStringMatcher
	cachePatterns := a.cachePatterns
	if !cachePatterns.Valid || cachePatterns.Bool {
		tmp, ok := a.stringMatcherCache.Load(pattern)
		if ok {
			matcher = tmp.(*AntPathStringMatcher)
		}
	}
	if matcher == nil {
		matcher = NewAntPathStringMatcherWithCaseSensitive(pattern, a.caseSensitive)
		if !cachePatterns.Valid && a.stringMatcherCache.Len() >= CACHE_TURNOFF_THRESHOLD {
			// Try to adapt to the runtime situation that we're encountering:
			// There are obviously too many different patterns coming in here...
			// So let's turn off the cache since the patterns are unlikely to be reoccurring.
			a.DeactivatePatternCache()
			return matcher
		}
		if (!a.cachePatterns.Valid) || cachePatterns.Bool {
			a.stringMatcherCache.Store(pattern, matcher)
		}
	}
	return matcher
}

func (a *AntPathMatcher) DeactivatePatternCache() {
	a.cachePatterns.SetValid(false)
	a.tokenizedPatternCache = pkg.MySyncMap{}
	a.stringMatcherCache = pkg.MySyncMap{}
}

func (a *AntPathMatcher) SetPathSeparator(pathSeparator string) {
	if strings.TrimSpace(pathSeparator) == "" {
		pathSeparator = DEFAULT_PATH_SEPARATOR
	}
	a.pathSeparator = pathSeparator
	a.pathSeparatorPatternCache = NewPathSeparatorPatternCache(pathSeparator)
}

//endregion

//region PathSeparatorPatternCache

type PathSeparatorPatternCache struct {
	endsOnWildcard       string
	endsOnDoubleWildcard string
}

func NewPathSeparatorPatternCache(pathSeparator string) *PathSeparatorPatternCache {
	return &PathSeparatorPatternCache{
		endsOnWildcard:       pathSeparator + "*",
		endsOnDoubleWildcard: pathSeparator + "**",
	}
}

//endregion

//region AntPatternComparator

type AntPatternComparator struct {
	path string
}

func NewAntPatternComparator(path string) *AntPatternComparator {
	return &AntPatternComparator{path: path}
}

func (a *AntPatternComparator) Compare(pattern1, pattern2 string) int {
	info1 := NewPatternInfo(pattern1)
	info2 := NewPatternInfo(pattern2)
	if info1.isLeastSpecific() && info2.isLeastSpecific() {
		return 0
	} else if info1.isLeastSpecific() {
		return 1
	} else if info2.isLeastSpecific() {
		return -1
	}

	pattern1EqualsPath := pattern1 == a.path
	pattern2EqualsPath := pattern2 == a.path

	if pattern1EqualsPath && pattern2EqualsPath {
		return 0
	} else if pattern1EqualsPath {
		return -1
	} else if pattern2EqualsPath {
		return 1
	}

	if info1.prefixPattern && info2.prefixPattern {
		return info2.getLength() - info1.getLength()
	} else if info1.prefixPattern && info2.doubleWildcards == 0 {
		return 1
	} else if info2.prefixPattern && info1.doubleWildcards == 0 {
		return -1
	}

	if info1.getTotalCount() != info2.getTotalCount() {
		return info1.getTotalCount() - info2.getTotalCount()
	}

	if info1.getLength() != info2.getLength() {
		return info2.getLength() - info1.getLength()
	}

	if info1.singleWildcards < info2.singleWildcards {
		return -1
	} else if info1.singleWildcards > info2.singleWildcards {
		return 1
	}

	if info1.uriVars < info2.uriVars {
		return -1
	} else if info1.uriVars > info2.uriVars {
		return 1
	}
	return 0
}

//endregion

//region patternInfo

type patternInfo struct {
	pattern         string
	uriVars         int
	singleWildcards int
	doubleWildcards int
	catchAllPattern bool
	prefixPattern   bool
	length          int
}

func NewPatternInfo(pattern string) *patternInfo {
	p := &patternInfo{}
	p.pattern = pattern
	p.initCounters()
	p.catchAllPattern = pattern == "/**"
	p.prefixPattern = !p.catchAllPattern && strings.HasSuffix(pattern, "/**")

	if p.uriVars == 0 { //always true,doesn't it?
		p.length = len(pattern)
	}
	return p
}

func (p *patternInfo) initCounters() {
	for pos := 0; pos < len(p.pattern); {
		if p.pattern[pos] == byte('{') {
			p.uriVars++
			pos++
		} else if p.pattern[pos] == byte('*') {
			if pos+1 < len(p.pattern) && p.pattern[pos+1] == byte('*') {
				p.doubleWildcards++
				pos += 2
			} else if pos > 0 && p.pattern[pos-1:] != ".*" {
				p.singleWildcards++
				pos++
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}
func (p *patternInfo) getTotalCount() int {
	return p.uriVars + p.singleWildcards + (2 * p.doubleWildcards)
}

func (p *patternInfo) getLength() int {
	//if patternInfo is public ,p.length will not be 0 in java
	if p.length == 0 {
		p.length = len(VARIABLE_PATTERN.ReplaceAllString(p.pattern, "#"))
	}
	return p.length
}

func (p *patternInfo) isLeastSpecific() bool {
	return p.catchAllPattern
}

//endregion

//region AntPathStringMatcher

type AntPathStringMatcher struct {
	caseSensitive bool
	exactMatch    bool
	rawPattern    string
	variableNames []string
	pattern       *regexp.Regexp
}

// const DEFAULT_VARIABLE_PATTERN="(.*)"
const DEFAULT_VARIABLE_PATTERN = "((?s).*)"

func NewAntPathStringMatcher(pattern string) *AntPathStringMatcher {
	return NewAntPathStringMatcherWithCaseSensitive(pattern, true)
}

func NewAntPathStringMatcherWithCaseSensitive(pattern string, caseSensitive bool) *AntPathStringMatcher {
	GLOB_PATTERN := regexp.MustCompile("\\?|\\*|\\{((?:\\{[^/]+?\\}|[^/{}]|\\\\[{}])+?)\\}")
	a := &AntPathStringMatcher{
		caseSensitive: caseSensitive,
		rawPattern:    pattern,
	}
	patternBuilder := bytes.NewBufferString("")
	allStrs := GLOB_PATTERN.FindAllString(pattern, -1)
	allIndexs := GLOB_PATTERN.FindAllStringIndex(pattern, -1)
	end := 0
	for k := range allIndexs {
		patternBuilder.WriteString(quote(pattern, end, allIndexs[k][0]))
		if strings.EqualFold("?", allStrs[k]) {
			patternBuilder.WriteString(".")
		} else if strings.EqualFold("*", allStrs[k]) {
			patternBuilder.WriteString(".*")
		} else if strings.HasPrefix(allStrs[k], "{") && strings.HasSuffix(allStrs[k], "}") {
			colonIdx := strings.IndexRune(allStrs[k], ':')
			if colonIdx == -1 {
				patternBuilder.WriteString(DEFAULT_VARIABLE_PATTERN)
				a.variableNames = append(a.variableNames, GLOB_PATTERN.FindStringSubmatch(allStrs[k])[1])
			} else {
				variablePattern := allStrs[k][colonIdx+1 : len(allStrs[k])-1]
				patternBuilder.WriteString("(")
				patternBuilder.WriteString(variablePattern)
				patternBuilder.WriteString(")")
				variableName := allStrs[k][1:colonIdx]
				a.variableNames = append(a.variableNames, variableName)
			}
		}
		end = allIndexs[k][1]
	}
	if len(allStrs) == 0 {
		a.exactMatch = true
		a.pattern = nil
	} else {
		a.exactMatch = false
		patternBuilder.WriteString(quote(pattern, end, len(pattern)))
		var str string
		if !a.caseSensitive {
			str = "(?i)"
		}
		a.pattern = regexp.MustCompile(str + patternBuilder.String())
	}
	return a
}

func quote(s string, start, end int) string {
	if start >= end {
		return ""
	}
	return regexp.QuoteMeta(s[start:end])
}

func (a *AntPathStringMatcher) matchStrings(str string, uriTemplateVariables map[string]string) bool {
	if a.exactMatch {
		if a.caseSensitive {
			return a.rawPattern == str
		} else {
			return strings.EqualFold(a.rawPattern, str)
		}
	} else if a.pattern != nil {
		strs := a.pattern.FindStringSubmatch(str)
		if strs == nil || len(strs) == 0 {
			return false
		}
		count := len(strs) - 1
		if strs[0] == str {
			if uriTemplateVariables != nil {
				if len(a.variableNames) != count {
					panic(fmt.Sprintf("The number of capturing groups in the pattern segment " +
						a.pattern.String() + " does not match the number of URI template variables it defines, " +
						"which can occur if capturing groups are used in a URI template regex. " +
						"Use non-capturing groups instead."))
				}
				for i := 1; i <= count; i++ {
					name := a.variableNames[i-1]
					if strings.HasPrefix(name, "*") {
						panic("Capturing patterns (" + name + ") are not " +
							"supported by the AntPathMatcher. Use the PathPatternParser instead.")
					}
					uriTemplateVariables[name] = strs[i]
				}
			}
			return true
		}
	}
	return false
}

//endregion
