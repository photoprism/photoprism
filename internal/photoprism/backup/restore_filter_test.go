package backup

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// dumpHeader is the start of a MariaDB dump, with uniqueChecks as the unique checks header line.
func dumpHeader(uniqueChecks, eol string) string {
	return strings.Join([]string{
		"/*M!999999\\- enable the sandbox mode */ ",
		"-- MariaDB dump 10.19-11.8.6-MariaDB, for debian-linux-gnu (x86_64)",
		"/*!40101 SET NAMES utf8mb4 */;",
		uniqueChecks,
		"/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;",
		"/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;",
		"",
	}, eol)
}

// dumpLines returns comment lines of size bytes in total.
func dumpLines(size int) string {
	s := strings.Repeat(strings.Repeat("-", 99)+"\n", size/100)

	if size%100 > 0 {
		s += strings.Repeat("-", size%100-1) + "\n"
	}

	return s
}

// filterDump returns what restoreReader passes on for a MariaDB dump.
func filterDump(t *testing.T, r io.Reader) string {
	t.Helper()

	out, err := io.ReadAll(restoreReader(dsn.DriverMariaDB, r))
	require.NoError(t, err)

	return string(out)
}

func TestRestoreReader(t *testing.T) {
	off, on := string(uniqueChecksOff), string(uniqueChecksOn)

	t.Run("HeaderReplaced", func(t *testing.T) {
		assert.Equal(t, dumpHeader(on, "\n"), filterDump(t, strings.NewReader(dumpHeader(off, "\n"))))
	})
	t.Run("HeaderReplacedCrLf", func(t *testing.T) {
		assert.Equal(t, dumpHeader(on, "\r\n"), filterDump(t, strings.NewReader(dumpHeader(off, "\r\n"))))
	})
	t.Run("OneByteReads", func(t *testing.T) {
		out, err := io.ReadAll(iotest.OneByteReader(restoreReader(dsn.DriverMariaDB, iotest.OneByteReader(strings.NewReader(dumpHeader(off, "\n"))))))
		require.NoError(t, err)
		assert.Equal(t, dumpHeader(on, "\n"), string(out))
	})
	t.Run("NoHeader", func(t *testing.T) {
		dump := dumpHeader("/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS */;", "\n") + "INSERT INTO t VALUES (1);\n"
		assert.Equal(t, dump, filterDump(t, strings.NewReader(dump)))
	})
	t.Run("PartialLines", func(t *testing.T) {
		// Only a whole line is replaced, with a line break and nothing else around it.
		for _, dump := range []string{
			" " + off + "\n",
			off + " \n",
			off + "\r\r\n",
			off,
			"-- " + off + "\n",
			strings.Repeat("x", uniqueChecksLineBytes) + off + "\n",
			strings.Repeat("x", 3*uniqueChecksLineBytes+7) + off + "\n",
		} {
			assert.Equal(t, dump, filterDump(t, strings.NewReader(dump)), "%.40q", dump)
		}
	})
	t.Run("AfterLongLine", func(t *testing.T) {
		// A line longer than the buffer is passed on in parts, and the next line is checked again.
		long := strings.Repeat("x", 2*uniqueChecksLineBytes+5) + "\n"
		assert.Equal(t, long+dumpHeader(on, "\n"), filterDump(t, strings.NewReader(long+dumpHeader(off, "\n"))))
	})
	t.Run("OnceOnly", func(t *testing.T) {
		assert.Equal(t, dumpHeader(on, "\n")+off+"\n", filterDump(t, strings.NewReader(dumpHeader(off, "\n")+off+"\n")))
	})
	t.Run("HeaderInValue", func(t *testing.T) {
		dump := dumpHeader(off, "\n") + "INSERT INTO t VALUES (1,'" + off + "'),\n(2,'" + off + "\\n');\n"
		assert.Equal(t, dumpHeader(on, "\n")+"INSERT INTO t VALUES (1,'"+off+"'),\n(2,'"+off+"\\n');\n",
			filterDump(t, strings.NewReader(dump)))
	})
	t.Run("OnlyInPrefix", func(t *testing.T) {
		// A line that starts after the scan prefix is passed on unchanged.
		for _, size := range []int{uniqueChecksScanBytes - 100, uniqueChecksScanBytes - 10, uniqueChecksScanBytes, uniqueChecksScanBytes + 1000} {
			prefix := dumpLines(size)
			want := prefix + off + "\n"
			if size < uniqueChecksScanBytes {
				want = prefix + on + "\n"
			}
			assert.Equal(t, want, filterDump(t, strings.NewReader(prefix+off+"\n")), "size %d", size)
		}
	})
	t.Run("BoundedLine", func(t *testing.T) {
		// A line without a break is passed on in parts as it arrives, rather than read whole.
		pr, pw := io.Pipe()
		t.Cleanup(func() { _ = pw.Close() })
		go func() { _, _ = pw.Write([]byte(strings.Repeat("x", 3*uniqueChecksLineBytes))) }()
		read := make(chan int, 1)
		go func() {
			n, _ := restoreReader(dsn.DriverMariaDB, pr).Read(make([]byte, 8*uniqueChecksLineBytes))
			read <- n
		}()
		select {
		case n := <-read:
			assert.Positive(t, n)
			assert.LessOrEqual(t, n, uniqueChecksLineBytes)
		case <-time.After(5 * time.Second):
			t.Fatal("read did not return before the line ended")
		}
	})
	t.Run("NoHeaderLogged", func(t *testing.T) {
		// A dump without the header line is reported once, at debug level.
		for _, dump := range []string{"", "INSERT INTO t VALUES (1);\n", dumpLines(3 * uniqueChecksScanBytes)} {
			hook := captureLog(t)
			assert.Equal(t, dump, filterDump(t, strings.NewReader(dump)))
			require.Len(t, hook.AllEntries(), 1, "size %d", len(dump))
			assert.Equal(t, logrus.DebugLevel, hook.LastEntry().Level)
			assert.Equal(t, "restore: found no unique checks header line, reading the dump as it is", hook.LastEntry().Message)
		}
		for _, r := range []io.Reader{
			strings.NewReader(dumpHeader(off, "\n")),
			io.MultiReader(strings.NewReader("INSERT"), iotest.ErrReader(errors.New("read failed"))),
		} {
			hook := captureLog(t)
			_, _ = io.ReadAll(restoreReader(dsn.DriverMariaDB, r))
			assert.Empty(t, hook.AllEntries())
		}
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, filterDump(t, strings.NewReader("")))
	})
	t.Run("LargeDump", func(t *testing.T) {
		rows := strings.Repeat("INSERT INTO t VALUES (1,'"+strings.Repeat("v", 5000)+"');\n", 100)
		assert.Equal(t, dumpHeader(on, "\n")+rows, filterDump(t, strings.NewReader(dumpHeader(off, "\n")+rows)))
	})
	t.Run("ReadError", func(t *testing.T) {
		// The data read before an error is passed on, followed by the error.
		r := restoreReader(dsn.DriverMariaDB, io.MultiReader(strings.NewReader(dumpHeader(off, "\n")+"INSERT"),
			iotest.ErrReader(errors.New("read failed"))))
		out, err := io.ReadAll(r)
		assert.EqualError(t, err, "read failed")
		assert.Equal(t, dumpHeader(on, "\n")+"INSERT", string(out))
	})
	t.Run("OtherDrivers", func(t *testing.T) {
		for _, driver := range []string{dsn.DriverSQLite3, ""} {
			r := strings.NewReader(dumpHeader(off, "\n"))
			assert.Same(t, r, restoreReader(driver, r), "driver %q", driver)
		}
		out, err := io.ReadAll(restoreReader(dsn.DriverMySQL, bytes.NewReader([]byte(dumpHeader(off, "\n")))))
		require.NoError(t, err)
		assert.Equal(t, dumpHeader(on, "\n"), string(out))
	})
}

func TestIsUniqueChecksOff(t *testing.T) {
	assert.True(t, isUniqueChecksOff(append(append([]byte{}, uniqueChecksOff...), '\n')))
	assert.True(t, isUniqueChecksOff(append(append([]byte{}, uniqueChecksOff...), '\r', '\n')))
	assert.False(t, isUniqueChecksOff(uniqueChecksOff))
	assert.False(t, isUniqueChecksOff(append(append([]byte{}, uniqueChecksOn...), '\n')))
	assert.False(t, isUniqueChecksOff([]byte("\n")))
}
