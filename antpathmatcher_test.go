package antpathmatcher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// @Author :George
// @File: antpathmatcher_test
// @Version: 1.0.0
// @Date 2023/10/16 14:09

var pathMatcher *AntPathMatcher

func init() {
	//pathMatcher = NewAntPathMatcher()
}

func Test_match(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	// test exact matching
	e.True(pathMatcher.Match("test", "test"))
	e.True(pathMatcher.Match("/test", "/test"))

	// SPR-14141
	e.True(pathMatcher.Match("https://example.org", "https://example.org"))
	e.False(pathMatcher.Match("/test.jpg", "test.jpg"))
	e.False(pathMatcher.Match("test", "/test"))
	e.False(pathMatcher.Match("/test", "test"))

	// test matching with ?'s
	e.True(pathMatcher.Match("t?st", "test"))
	e.True(pathMatcher.Match("??st", "test"))
	e.True(pathMatcher.Match("tes?", "test"))
	e.True(pathMatcher.Match("te??", "test"))
	e.True(pathMatcher.Match("?es?", "test"))
	e.False(pathMatcher.Match("tes?", "tes"))
	e.False(pathMatcher.Match("tes?", "testt"))
	e.False(pathMatcher.Match("tes?", "tsst"))

	// test matching with *'s
	e.True(pathMatcher.Match("*", "test"))
	e.True(pathMatcher.Match("test*", "test"))
	e.True(pathMatcher.Match("test*", "testTest"))
	e.True(pathMatcher.Match("test/*", "test/Test"))
	e.True(pathMatcher.Match("test/*", "test/t"))
	e.True(pathMatcher.Match("test/*", "test/"))
	e.True(pathMatcher.Match("*test*", "AnothertestTest"))
	e.True(pathMatcher.Match("*test", "Anothertest"))
	e.True(pathMatcher.Match("*.*", "test."))
	e.True(pathMatcher.Match("*.*", "test.test"))
	e.True(pathMatcher.Match("*.*", "test.test.test"))
	e.True(pathMatcher.Match("test*aaa", "testblaaaa"))
	e.False(pathMatcher.Match("test*", "tst"))
	e.False(pathMatcher.Match("test*", "tsttest"))
	e.False(pathMatcher.Match("test*", "test/"))
	e.False(pathMatcher.Match("test*", "test/t"))
	e.False(pathMatcher.Match("test/*", "test"))
	e.False(pathMatcher.Match("*test*", "tsttst"))
	e.False(pathMatcher.Match("*test", "tsttst"))
	e.False(pathMatcher.Match("*.*", "tsttst"))
	e.False(pathMatcher.Match("test*aaa", "test"))
	e.False(pathMatcher.Match("test*aaa", "testblaaab"))

	// test matching with ?'s and /'s
	e.True(pathMatcher.Match("/?", "/a"))
	e.True(pathMatcher.Match("/?/a", "/a/a"))
	e.True(pathMatcher.Match("/a/?", "/a/b"))
	e.True(pathMatcher.Match("/??/a", "/aa/a"))
	e.True(pathMatcher.Match("/a/??", "/a/bb"))
	e.True(pathMatcher.Match("/?", "/a"))

	// test matching with **'s
	e.True(pathMatcher.Match("/**", "/testing/testing"))
	e.True(pathMatcher.Match("/*/**", "/testing/testing"))
	e.True(pathMatcher.Match("/**/*", "/testing/testing"))
	e.True(pathMatcher.Match("/bla/**/bla", "/bla/testing/testing/bla"))
	e.True(pathMatcher.Match("/bla/**/bla", "/bla/testing/testing/bla/bla"))
	e.True(pathMatcher.Match("/**/test", "/bla/bla/test"))
	e.True(pathMatcher.Match("/bla/**/**/bla", "/bla/bla/bla/bla/bla/bla"))
	e.True(pathMatcher.Match("/bla*bla/test", "/blaXXXbla/test"))
	e.True(pathMatcher.Match("/*bla/test", "/XXXbla/test"))
	e.False(pathMatcher.Match("/bla*bla/test", "/blaXXXbl/test"))
	e.False(pathMatcher.Match("/*bla/test", "XXXblab/test"))
	e.False(pathMatcher.Match("/*bla/test", "XXXbl/test"))

	e.False(pathMatcher.Match("/????", "/bala/bla"))
	e.False(pathMatcher.Match("/**/*bla", "/bla/bla/bla/bbb"))

	e.True(pathMatcher.Match("/*bla*/**/bla/**", "/XXXblaXXXX/testing/testing/bla/testing/testing/"))
	e.True(pathMatcher.Match("/*bla*/**/bla/*", "/XXXblaXXXX/testing/testing/bla/testing"))
	e.True(pathMatcher.Match("/*bla*/**/bla/**", "/XXXblaXXXX/testing/testing/bla/testing/testing"))
	e.True(pathMatcher.Match("/*bla*/**/bla/**", "/XXXblaXXXX/testing/testing/bla/testing/testing.jpg"))

	e.True(pathMatcher.Match("*bla*/**/bla/**", "XXXblaXXXX/testing/testing/bla/testing/testing/"))
	e.True(pathMatcher.Match("*bla*/**/bla/*", "XXXblaXXXX/testing/testing/bla/testing"))
	e.True(pathMatcher.Match("*bla*/**/bla/**", "XXXblaXXXX/testing/testing/bla/testing/testing"))
	e.False(pathMatcher.Match("*bla*/**/bla/*", "XXXblaXXXX/testing/testing/bla/testing/testing"))

	e.False(pathMatcher.Match("/x/x/**/bla", "/x/x/x/"))
	e.True(pathMatcher.Match("/foo/bar/**", "/foo/bar"))
	e.True(pathMatcher.Match("", ""))
	e.True(pathMatcher.Match("/{bla}.*", "/testing.html"))
	e.True(pathMatcher.Match("/{bla}", "//x\ny"))
	//todo from https://github.com/spring-projects/spring-framework/blob/HEAD/spring-core/src/test/java/org/springframework/util/AntPathMatcherTests.java
	e.True(pathMatcher.Match("/{var:.*}", "/x\ny"))
}

// SPR-14247
func Test_matchWithTrimTokensEnabled(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	pathMatcher.trimTokens = true
	e := assert.New(t)
	e.True(pathMatcher.MatchStart("/foo/bar", "/foo /bar"))
}

func Test_matchStart(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	// test exact matching
	e.True(pathMatcher.MatchStart("test", "test"))
	e.True(pathMatcher.MatchStart("/test", "/test"))
	e.False(pathMatcher.MatchStart("/test.jpg", "test.jpg"))
	e.False(pathMatcher.MatchStart("test", "/test"))
	e.False(pathMatcher.MatchStart("/test", "test"))

	// test matching with ?'s
	e.True(pathMatcher.MatchStart("t?st", "test"))
	e.True(pathMatcher.MatchStart("??st", "test"))
	e.True(pathMatcher.MatchStart("tes?", "test"))
	e.True(pathMatcher.MatchStart("te??", "test"))
	e.True(pathMatcher.MatchStart("?es?", "test"))
	e.False(pathMatcher.MatchStart("tes?", "tes"))
	e.False(pathMatcher.MatchStart("tes?", "testt"))
	e.False(pathMatcher.MatchStart("tes?", "tsst"))

	// test matching with *'s
	e.True(pathMatcher.MatchStart("*", "test"))
	e.True(pathMatcher.MatchStart("test*", "test"))
	e.True(pathMatcher.MatchStart("test*", "testTest"))
	e.True(pathMatcher.MatchStart("test/*", "test/Test"))
	e.True(pathMatcher.MatchStart("test/*", "test/t"))
	e.True(pathMatcher.MatchStart("test/*", "test/"))
	e.True(pathMatcher.MatchStart("*test*", "AnothertestTest"))
	e.True(pathMatcher.MatchStart("*test", "Anothertest"))
	e.True(pathMatcher.MatchStart("*.*", "test."))
	e.True(pathMatcher.MatchStart("*.*", "test.test"))
	e.True(pathMatcher.MatchStart("*.*", "test.test.test"))
	e.True(pathMatcher.MatchStart("test*aaa", "testblaaaa"))
	e.False(pathMatcher.MatchStart("test*", "tst"))
	e.False(pathMatcher.MatchStart("test*", "test/"))
	e.False(pathMatcher.MatchStart("test*", "tsttest"))
	e.False(pathMatcher.MatchStart("test*", "test/"))
	e.False(pathMatcher.MatchStart("test*", "test/t"))
	e.True(pathMatcher.MatchStart("test/*", "test"))
	e.True(pathMatcher.MatchStart("test/t*.txt", "test"))
	e.False(pathMatcher.MatchStart("*test*", "tsttst"))
	e.False(pathMatcher.MatchStart("*test", "tsttst"))
	e.False(pathMatcher.MatchStart("*.*", "tsttst"))
	e.False(pathMatcher.MatchStart("test*aaa", "test"))
	e.False(pathMatcher.MatchStart("test*aaa", "testblaaab"))

	// test matching with ?'s and /'s
	e.True(pathMatcher.MatchStart("/?", "/a"))
	e.True(pathMatcher.MatchStart("/?/a", "/a/a"))
	e.True(pathMatcher.MatchStart("/a/?", "/a/b"))
	e.True(pathMatcher.MatchStart("/??/a", "/aa/a"))
	e.True(pathMatcher.MatchStart("/a/??", "/a/bb"))
	e.True(pathMatcher.MatchStart("/?", "/a"))

	// test matching with **'s
	e.True(pathMatcher.MatchStart("/**", "/testing/testing"))
	e.True(pathMatcher.MatchStart("/*/**", "/testing/testing"))
	e.True(pathMatcher.MatchStart("/**/*", "/testing/testing"))
	e.True(pathMatcher.MatchStart("test*/**", "test/"))
	e.True(pathMatcher.MatchStart("test*/**", "test/t"))
	e.True(pathMatcher.MatchStart("/bla/**/bla", "/bla/testing/testing/bla"))
	e.True(pathMatcher.MatchStart("/bla/**/bla", "/bla/testing/testing/bla/bla"))
	e.True(pathMatcher.MatchStart("/**/test", "/bla/bla/test"))
	e.True(pathMatcher.MatchStart("/bla/**/**/bla", "/bla/bla/bla/bla/bla/bla"))
	e.True(pathMatcher.MatchStart("/bla*bla/test", "/blaXXXbla/test"))
	e.True(pathMatcher.MatchStart("/*bla/test", "/XXXbla/test"))
	e.False(pathMatcher.MatchStart("/bla*bla/test", "/blaXXXbl/test"))
	e.False(pathMatcher.MatchStart("/*bla/test", "XXXblab/test"))
	e.False(pathMatcher.MatchStart("/*bla/test", "XXXbl/test"))

	e.False(pathMatcher.MatchStart("/????", "/bala/bla"))
	e.True(pathMatcher.MatchStart("/**/*bla", "/bla/bla/bla/bbb"))

	e.True(pathMatcher.MatchStart("/*bla*/**/bla/**", "/XXXblaXXXX/testing/testing/bla/testing/testing/"))
	e.True(pathMatcher.MatchStart("/*bla*/**/bla/*", "/XXXblaXXXX/testing/testing/bla/testing"))
	e.True(pathMatcher.MatchStart("/*bla*/**/bla/**", "/XXXblaXXXX/testing/testing/bla/testing/testing"))
	e.True(pathMatcher.MatchStart("/*bla*/**/bla/**", "/XXXblaXXXX/testing/testing/bla/testing/testing.jpg"))

	e.True(pathMatcher.MatchStart("*bla*/**/bla/**", "XXXblaXXXX/testing/testing/bla/testing/testing/"))
	e.True(pathMatcher.MatchStart("*bla*/**/bla/*", "XXXblaXXXX/testing/testing/bla/testing"))
	e.True(pathMatcher.MatchStart("*bla*/**/bla/**", "XXXblaXXXX/testing/testing/bla/testing/testing"))
	e.True(pathMatcher.MatchStart("*bla*/**/bla/*", "XXXblaXXXX/testing/testing/bla/testing/testing"))

	e.True(pathMatcher.MatchStart("/x/x/**/bla", "/x/x/x/"))

	e.True(pathMatcher.MatchStart("", ""))
}

func Test_uniqueDeliminator(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	pathMatcher.SetPathSeparator(".")
	e := assert.New(t)

	// test exact matching
	e.True(pathMatcher.Match("test", "test"))
	e.True(pathMatcher.Match(".test", ".test"))
	e.False(pathMatcher.Match(".test/jpg", "test/jpg"))
	e.False(pathMatcher.Match("test", ".test"))
	e.False(pathMatcher.Match(".test", "test"))

	// test matching with ?'s
	e.True(pathMatcher.Match("t?st", "test"))
	e.True(pathMatcher.Match("??st", "test"))
	e.True(pathMatcher.Match("tes?", "test"))
	e.True(pathMatcher.Match("te??", "test"))
	e.True(pathMatcher.Match("?es?", "test"))
	e.False(pathMatcher.Match("tes?", "tes"))
	e.False(pathMatcher.Match("tes?", "testt"))
	e.False(pathMatcher.Match("tes?", "tsst"))

	// test matching with *'s
	e.True(pathMatcher.Match("*", "test"))
	e.True(pathMatcher.Match("test*", "test"))
	e.True(pathMatcher.Match("test*", "testTest"))
	e.True(pathMatcher.Match("*test*", "AnothertestTest"))
	e.True(pathMatcher.Match("*test", "Anothertest"))
	e.True(pathMatcher.Match("*/*", "test/"))
	e.True(pathMatcher.Match("*/*", "test/test"))
	e.True(pathMatcher.Match("*/*", "test/test/test"))
	e.True(pathMatcher.Match("test*aaa", "testblaaaa"))
	e.False(pathMatcher.Match("test*", "tst"))
	e.False(pathMatcher.Match("test*", "tsttest"))
	e.False(pathMatcher.Match("*test*", "tsttst"))
	e.False(pathMatcher.Match("*test", "tsttst"))
	e.False(pathMatcher.Match("*/*", "tsttst"))
	e.False(pathMatcher.Match("test*aaa", "test"))
	e.False(pathMatcher.Match("test*aaa", "testblaaab"))

	// test matching with ?'s and .'s
	e.True(pathMatcher.Match(".?", ".a"))
	e.True(pathMatcher.Match(".?.a", ".a.a"))
	e.True(pathMatcher.Match(".a.?", ".a.b"))
	e.True(pathMatcher.Match(".??.a", ".aa.a"))
	e.True(pathMatcher.Match(".a.??", ".a.bb"))
	e.True(pathMatcher.Match(".?", ".a"))

	// test matching with **'s
	e.True(pathMatcher.Match(".**", ".testing.testing"))
	e.True(pathMatcher.Match(".*.**", ".testing.testing"))
	e.True(pathMatcher.Match(".**.*", ".testing.testing"))
	e.True(pathMatcher.Match(".bla.**.bla", ".bla.testing.testing.bla"))
	e.True(pathMatcher.Match(".bla.**.bla", ".bla.testing.testing.bla.bla"))
	e.True(pathMatcher.Match(".**.test", ".bla.bla.test"))
	e.True(pathMatcher.Match(".bla.**.**.bla", ".bla.bla.bla.bla.bla.bla"))
	e.True(pathMatcher.Match(".bla*bla.test", ".blaXXXbla.test"))
	e.True(pathMatcher.Match(".*bla.test", ".XXXbla.test"))
	e.False(pathMatcher.Match(".bla*bla.test", ".blaXXXbl.test"))
	e.False(pathMatcher.Match(".*bla.test", "XXXblab.test"))
	e.False(pathMatcher.Match(".*bla.test", "XXXbl.test"))
}

func Test_extractPathWithinPattern(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/commit.html", "/docs/commit.html"), "")

	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/*", "/docs/cvs/commit"), "cvs/commit")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/cvs/*.html", "/docs/cvs/commit.html"), "commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/**", "/docs/cvs/commit"), "cvs/commit")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/**/*.html", "/docs/cvs/commit.html"), "cvs/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/**/*.html", "/docs/commit.html"), "commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/*.html", "/commit.html"), "commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/*.html", "/docs/commit.html"), "docs/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("*.html", "/commit.html"), "/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("*.html", "/docs/commit.html"), "/docs/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("**/*.*", "/docs/commit.html"), "/docs/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("*", "/docs/commit.html"), "/docs/commit.html")

	// SPR-10515
	e.Equal(pathMatcher.ExtractPathWithinPattern("**/commit.html", "/docs/cvs/other/commit.html"), "/docs/cvs/other/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/**/commit.html", "/docs/cvs/other/commit.html"), "cvs/other/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/**/**/**/**", "/docs/cvs/other/commit.html"), "cvs/other/commit.html")

	e.Equal(pathMatcher.ExtractPathWithinPattern("/d?cs/*", "/docs/cvs/commit"), "docs/cvs/commit")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/docs/c?s/*.html", "/docs/cvs/commit.html"), "cvs/commit.html")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/d?cs/**", "/docs/cvs/commit"), "docs/cvs/commit")
	e.Equal(pathMatcher.ExtractPathWithinPattern("/d?cs/**/*.html", "/docs/cvs/commit.html"), "docs/cvs/commit.html")
}

func Test_extractUriTemplateVariables(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	result := pathMatcher.ExtractUriTemplateVariables("/hotels/{hotel}", "/hotels/1")
	e.Equal(result["hotel"], "1")

	result = pathMatcher.ExtractUriTemplateVariables("/h?tels/{hotel}", "/hotels/1")
	e.Equal(result["hotel"], "1")

	result = pathMatcher.ExtractUriTemplateVariables("/hotels/{hotel}/bookings/{booking}", "/hotels/1/bookings/2")
	e.Equal(result["hotel"], "1")
	e.Equal(result["booking"], "2")

	result = pathMatcher.ExtractUriTemplateVariables("/**/hotels/**/{hotel}", "/foo/hotels/bar/1")
	e.Equal(result["hotel"], "1")

	result = pathMatcher.ExtractUriTemplateVariables("/{page}.html", "/42.html")
	e.Equal(result["page"], "42")

	result = pathMatcher.ExtractUriTemplateVariables("/{page}.*", "/42.html")
	e.Equal(result["page"], "42")

	result = pathMatcher.ExtractUriTemplateVariables("/A-{B}-C", "/A-b-C")
	e.Equal(result["B"], "b")

	result = pathMatcher.ExtractUriTemplateVariables("/{name}.{extension}", "/test.html")
	e.Equal(result["name"], "test")
	e.Equal(result["extension"], "html")
}

func Test_extractUriTemplateVariablesRegex(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)

	result := pathMatcher.ExtractUriTemplateVariables("{symbolicName:[\\w\\.]+}-{version:[\\w\\.]+}.jar", "com.example-1.0.0.jar")
	e.Equal(result["symbolicName"], "com.example")
	e.Equal(result["version"], "1.0.0")

	result = pathMatcher.ExtractUriTemplateVariables("{symbolicName:[\\w\\.]+}-sources-{version:[\\w\\.]+}.jar",
		"com.example-sources-1.0.0.jar")
	e.Equal(result["symbolicName"], "com.example")
	e.Equal(result["version"], "1.0.0")

}

// SPR-7787
func Test_extractUriTemplateVarsRegexQualifiers(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	result := pathMatcher.ExtractUriTemplateVariables("{symbolicName:[\\p{L}\\.]+}-sources-{version:[\\p{N}\\.]+}.jar",
		"com.example-sources-1.0.0.jar")
	e.Equal(result["symbolicName"], "com.example")
	e.Equal(result["version"], "1.0.0")

	result = pathMatcher.ExtractUriTemplateVariables("{symbolicName:[\\w\\.]+}-sources-{version:[\\d\\.]+}-{year:\\d{4}}{month:\\d{2}}{day:\\d{2}}.jar",
		"com.example-sources-1.0.0-20100220.jar")
	e.Equal(result["symbolicName"], "com.example")
	e.Equal(result["version"], "1.0.0")
	e.Equal(result["year"], "2010")
	e.Equal(result["month"], "02")
	e.Equal(result["day"], "20")

	result = pathMatcher.ExtractUriTemplateVariables("{symbolicName:[\\p{L}\\.]+}-sources-{version:[\\p{N}\\.\\{\\}]+}.jar",
		"com.example-sources-1.0.0.{12}.jar")
	e.Equal(result["symbolicName"], "com.example")
	e.Equal(result["version"], "1.0.0.{12}")
}

// SPR-8455
func Test_extractUriTemplateVarsRegexCapturingGroups(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	defer func() {
		err := recover()
		e.Contains(err.(string), "The number of capturing groups in the pattern")
	}()
	pathMatcher.ExtractUriTemplateVariables("/web/{id:foo(bar)?}", "/web/foobar")
}

func Test_combine(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	e.Equal(pathMatcher.Combine("/hotels/*", "booking"), "/hotels/booking")
	e.Equal(pathMatcher.Combine("/hotels/*", "/booking"), "/hotels/booking")
	e.Equal(pathMatcher.Combine("/hotels/**", "booking"), "/hotels/**/booking")
	e.Equal(pathMatcher.Combine("/hotels/**", "/booking"), "/hotels/**/booking")
	e.Equal(pathMatcher.Combine("/hotels", "/booking"), "/hotels/booking")
	e.Equal(pathMatcher.Combine("/hotels", "booking"), "/hotels/booking")
	e.Equal(pathMatcher.Combine("/hotels/", "booking"), "/hotels/booking")
	e.Equal(pathMatcher.Combine("/hotels/*", "{hotel}"), "/hotels/{hotel}")
	e.Equal(pathMatcher.Combine("/hotels/**", "{hotel}"), "/hotels/**/{hotel}")
	e.Equal(pathMatcher.Combine("/hotels", "{hotel}"), "/hotels/{hotel}")
	e.Equal(pathMatcher.Combine("/hotels", "{hotel}.*"), "/hotels/{hotel}.*")
	e.Equal(pathMatcher.Combine("/hotels/*/booking", "{booking}"), "/hotels/*/booking/{booking}")
	e.Equal(pathMatcher.Combine("/*.html", "/hotel.html"), "/hotel.html")
	e.Equal(pathMatcher.Combine("/*.html", "/hotel"), "/hotel.html")
	e.Equal(pathMatcher.Combine("/*.html", "/hotel.*"), "/hotel.html")
	e.Equal(pathMatcher.Combine("/**", "/*.html"), "/*.html")
	e.Equal(pathMatcher.Combine("/*", "/*.html"), "/*.html")
	e.Equal(pathMatcher.Combine("/*.*", "/*.html"), "/*.html")
	// SPR-8858
	e.Equal(pathMatcher.Combine("/{foo}", "/bar"), "/{foo}/bar")
	// SPR-7970
	e.Equal(pathMatcher.Combine("/user", "/user"), "/user/user")
	// SPR-10062
	e.Equal(pathMatcher.Combine("/{foo:.*[^0-9].*}", "/edit/"), "/{foo:.*[^0-9].*}/edit/")
	// SPR-10554
	e.Equal(pathMatcher.Combine("/1.0", "/foo/test"), "/1.0/foo/test")
	// SPR-12975
	e.Equal(pathMatcher.Combine("/", "/hotel"), "/hotel")
	// SPR-12975
	e.Equal(pathMatcher.Combine("/hotel/", "/booking"), "/hotel/booking")

}

func Test_combineWithTwoFileExtensionPatterns(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	e := assert.New(t)
	e.Panics(func() { pathMatcher.Combine("/*.html", "/*.txt") })
}

func Test_patternComparator(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	comparator := pathMatcher.GetPatternComparator("/hotels/new")
	e := assert.New(t)

	e.Equal(comparator.Compare("/hotels/new", "/hotels/new"), 0)
	e.Equal(comparator.Compare("/hotels/new", "/hotels/*"), -1)
	e.Equal(comparator.Compare("/hotels/*", "/hotels/new"), 1)
	e.Equal(comparator.Compare("/hotels/*", "/hotels/*"), 0)

	e.Equal(comparator.Compare("/hotels/new", "/hotels/{hotel}"), -1)
	e.Equal(comparator.Compare("/hotels/{hotel}", "/hotels/new"), 1)
	e.Equal(comparator.Compare("/hotels/{hotel}", "/hotels/{hotel}"), 0)
	e.Equal(comparator.Compare("/hotels/{hotel}/booking", "/hotels/{hotel}/bookings/{booking}"), -1)
	e.Equal(comparator.Compare("/hotels/{hotel}/bookings/{booking}", "/hotels/{hotel}/booking"), 1)

	// SPR-10550
	e.Equal(comparator.Compare("/hotels/{hotel}/bookings/{booking}/cutomers/{customer}", "/**"), -1)
	e.Equal(comparator.Compare("/**", "/hotels/{hotel}/bookings/{booking}/cutomers/{customer}"), 1)
	e.Equal(comparator.Compare("/**", "/**"), 0)

	e.Equal(comparator.Compare("/hotels/{hotel}", "/hotels/*"), -1)
	e.Equal(comparator.Compare("/hotels/*", "/hotels/{hotel}"), 1)

	e.Equal(comparator.Compare("/hotels/*", "/hotels/*/**"), -1)
	e.Equal(comparator.Compare("/hotels/*/**", "/hotels/*"), 1)

	e.Equal(comparator.Compare("/hotels/new", "/hotels/new.*"), -1)
	e.Equal(comparator.Compare("/hotels/{hotel}", "/hotels/{hotel}.*"), 2)

	// SPR-6741
	e.Equal(comparator.Compare("/hotels/{hotel}/bookings/{booking}/cutomers/{customer}", "/hotels/**"), -1)
	e.Equal(comparator.Compare("/hotels/**", "/hotels/{hotel}/bookings/{booking}/cutomers/{customer}"), 1)
	e.Equal(comparator.Compare("/hotels/foo/bar/**", "/hotels/{hotel}"), 1)
	e.Equal(comparator.Compare("/hotels/{hotel}", "/hotels/foo/bar/**"), -1)

	// gh-23125
	e.Equal(comparator.Compare("/hotels/*/bookings/**", "/hotels/**"), -11)

	// SPR-8683
	e.Equal(comparator.Compare("/**", "/hotels/{hotel}"), 1)

	// longer is better
	e.Equal(comparator.Compare("/hotels", "/hotels2"), 1)

	// SPR-13139
	e.Equal(comparator.Compare("*", "*/**"), -1)
	e.Equal(comparator.Compare("*/**", "*"), 1)
}

func Test_patternComparatorSort(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	comparator := pathMatcher.GetPatternComparator("/hotels/new")
	e := assert.New(t)
	_, _ = comparator, e
	//todo whole case sort.(Compara<string>) i have no idea how to implementation
}

// SPR-8687
func Test_trimTokensOff(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	pathMatcher.trimTokens = false
	e := assert.New(t)
	e.True(pathMatcher.Match("/group/{groupName}/members", "/group/sales/members"))
	e.True(pathMatcher.Match("/group/{groupName}/members", "/group/  sales/members"))
	e.False(pathMatcher.Match("/group/{groupName}/members", "/Group/  Sales/Members"))
}

// SPR-13286
func Test_caseInsensitive(t *testing.T) {
	pathMatcher = NewAntPathMatcher()
	pathMatcher.caseSensitive = false
	e := assert.New(t)
	e.True(pathMatcher.Match("/group/{groupName}/members", "/group/sales/members"))
	e.True(pathMatcher.Match("/group/{groupName}/members", "/Group/Sales/Members"))
	e.True(pathMatcher.Match("/Group/{groupName}/Members", "/group/Sales/members"))
}
