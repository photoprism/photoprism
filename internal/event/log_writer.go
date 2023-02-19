package event

import "github.com/sirupsen/logrus"

// LogWriter is an output writer wrapper for using Logrus with the standard logger.
type LogWriter struct {
	Log   Logger
	Level logrus.Level
}

// Write implements io.Writer.
func (w *LogWriter) Write(b []byte) (int, error) {
	n := len(b)

	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}

	if w.Log != nil {
		w.Log.Log(w.Level, string(b))
	}

	return n, nil
}
