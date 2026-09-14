package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/txt/clip"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// recoveryHeaders lists the request headers a panic summary reports, in the order it reports them.
// Adding a header here makes it part of the log output, so the list holds media negotiation and
// client identification only.
var recoveryHeaders = []string{
	header.ContentType,
	header.Accept,
	header.AcceptEncoding,
	header.Range,
	header.UserAgent,
}

const (
	// unknownRoute is what a request summary reports in place of a route template when the request
	// matched none.
	unknownRoute = "?"
	// summaryValueBytes is the budget a request summary gives each value it renders, so the length
	// of the message follows the allowlist and every listed field is reported.
	summaryValueBytes = 128
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Debugf("server: %s (%s)\n%s", clean.Log(fmt.Sprint(err)), requestSummary(c), stack(3))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// requestSummary renders the caller, the method, the route template and the allowlisted headers of
// a request. The template is the registered pattern rather than the requested path.
func requestSummary(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return unknownRoute
	}

	route := c.FullPath()

	if route == "" {
		route = unknownRoute
	}

	// The parsed length is reported rather than the header, which the transport removes from a
	// chunked request.
	out := make([]string, 0, len(recoveryHeaders)+3)
	out = append(out, api.ClientIP(c))
	out = append(out, summaryValue(c.Request.Method)+" "+summaryValue(route))
	out = append(out, fmt.Sprintf("%s: %d", header.ContentLength, c.Request.ContentLength))

	for _, name := range recoveryHeaders {
		if v := c.Request.Header.Get(name); v != "" {
			out = append(out, name+": "+summaryValue(v))
		}
	}

	return strings.Join(out, ", ")
}

// summaryValue renders one value of a request summary within its budget and in quotes, so its
// extent does not depend on the characters it holds.
func summaryValue(s string) string {
	return clean.LogQuote(clip.Bytes(s, summaryValueBytes))
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we cannot find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			// #nosec G304 -- file path comes from runtime.Caller and is limited to source files.
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.ReplaceAll(name, centerDot, dot)
	return name
}
