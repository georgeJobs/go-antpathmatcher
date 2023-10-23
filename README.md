# go-antpathmatcher

[![Go Report Card](https://goreportcard.com/badge/github.com/georgeJobs/go-antpathmatcher)](https://goreportcard.com/report/github.com/georgeJobs/go-antpathmatcher)
[![Run Tests](https://github.com/georgeJobs/go-antpathmatcher/actions/workflows/go.yml/badge.svg)](https://github.com/georgeJobs/go-antpathmatcher/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/georgeJobs/go-antpathmatcher?status.svg)](https://godoc.org/github.com/georgeJobs/go-antpathmatcher)
## 🌌Usage 

### Start using it

Download and install it:

```sh
go get github.com/georgeJobs/go-antpathmatcher
```

Import it in your code:

```go
import "github.com/georgeJobs/go-antpathmatcher"
```

## 🌰Example

See the [example](_example/example01/main.go).

```go
package main

import (
	"github.com/georgeJobs/go-antpathmatcher"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	e := gin.New()
	e.Use(Auth())
	e.GET("/test/12.html",func(c *gin.Context){
		c.File("./12.html")
	})
	e.GET("/hello/x", func(c *gin.Context) {
		c.JSON(200,gin.H{"message":"x success"})
	})
	e.GET("/hello/y",func(c *gin.Context){
		c.JSON(200,gin.H{"message":"y success"})
	})
	e.Run(":8080")
}

var pathMatch *antpathmatcher.AntPathMatcher

func init() {
	pathMatch = antpathmatcher.NewAntPathMatcher()
}

func Auth() func(c *gin.Context) {
	return func(c *gin.Context) {
		pathSlice := []string{"/test/*.html", "/hello/*"}
		for k := range pathSlice {
			if pathMatch.Match(pathSlice[k], c.Request.URL.Path) {
				goto JUMP
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "status": 403, "message": "Forbidden", "timestamp": time.Now().Format("2006-01-02 15:04:05"), "path": c.Request.URL.Path})
		return
	JUMP:
		c.Next()
	}
}
```

## 📝 License

**go-antpathmatcher** is released under the MIT License. Check out the LICENSE for more information.

If you have any questions, feel free to contact us at：xx@jobs.com 📧