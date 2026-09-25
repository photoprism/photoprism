package meta

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

// JSONMaxFileBytes is the default encoded-size limit for JSON metadata sidecars.
const JSONMaxFileBytes int64 = 1 << 20

// ErrJSONFileTooLarge identifies a JSON sidecar exceeding the configured size limit.
var ErrJSONFileTooLarge = errors.New("json: file size exceeds limit")

// JSONFileLimit returns the positive byte limit configured for JSON metadata.
func JSONFileLimit() int64 {
	limit, err := strconv.ParseUint(strings.TrimSpace(os.Getenv("PHOTOPRISM_JSON_LIMIT")), 10, 64)
	if err != nil || limit == 0 || limit >= uint64(^uint(0)>>1) {
		return JSONMaxFileBytes
	}

	return int64(limit) //nolint:gosec // Limit is below the platform's maximum signed integer.
}

// readJSONSidecar bounds the advertised size and actual bytes read before parsing.
func readJSONSidecar(r io.Reader, size int64) ([]byte, error) {
	limit := JSONFileLimit()

	if size > limit {
		return nil, ErrJSONFileTooLarge
	}

	data, err := io.ReadAll(io.LimitReader(r, limit+1))

	if int64(len(data)) > limit {
		return nil, ErrJSONFileTooLarge
	} else if err != nil {
		return nil, err
	}

	return data, nil
}
