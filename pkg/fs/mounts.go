package fs

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// MountInfoFile is the mount table of the current process on Linux.
const MountInfoFile = "/proc/self/mountinfo"

// mountInfoBuffer is the number of bytes read per mountinfo line; escaped paths take up to 4 bytes per byte.
const mountInfoBuffer = 64 * 1024

// MountPoints returns the paths of the file systems mounted below dir, as paths below dir, also if dir is a
// link. It reads MountInfoFile and returns an error if it is missing, e.g. on other operating systems.
func MountPoints(dir string) ([]string, error) {
	if dir == "" {
		return nil, nil
	}

	f, err := os.Open(MountInfoFile)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	dir = filepath.Clean(dir)
	realDir := dir

	if resolved, evalErr := filepath.EvalSymlinks(dir); evalErr == nil {
		realDir = resolved
	}

	return mountPointsFrom(f, dir, realDir)
}

// mountPointsFrom returns the mount points in a mountinfo table that are below realDir, mapped to dir, in the
// order listed and without duplicates. Only the leading part of a line longer than the read buffer is parsed.
func mountPointsFrom(r io.Reader, dir, realDir string) (result []string, err error) {
	br := bufio.NewReaderSize(r, mountInfoBuffer)

	for {
		line, more, readErr := br.ReadLine()

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return result, nil
			}

			return result, readErr
		}

		// Fields: mount ID, parent ID, major:minor, root, mount point, options, ... They are separated by
		// single spaces, as other whitespace may be part of a path.
		fields := strings.Split(string(line), " ")

		// Discard the rest of a long line.
		for more && readErr == nil {
			_, more, readErr = br.ReadLine()
		}

		// The mount point is complete only if another field follows it.
		if len(fields) > 5 {
			mountPoint := filepath.Clean(unescapeMountPath(fields[4]))

			if mountPoint != realDir && InDir(mountPoint, realDir) {
				if p := filepath.Join(dir, RelName(mountPoint, realDir)); !slices.Contains(result, p) {
					result = append(result, p)
				}
			}
		}

		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return result, readErr
		}
	}
}

// unescapeMountPath decodes the octal escapes of a path in a mountinfo table, e.g. \040 for a space.
func unescapeMountPath(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}

	var b strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+4 <= len(s) {
			if n, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(n))
				i += 3
				continue
			}
		}

		b.WriteByte(s[i])
	}

	return b.String()
}
