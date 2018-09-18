package photoprism

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func GetExpandedFilename(filename string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if filename == "" {
		panic("filename was empty")
	}

	if len(filename) > 2 && filename[:2] == "~/" {
		filename = filepath.Join(dir, filename[2:])
	}

	result, _ := filepath.Abs(filename)

	return result
}

func getRandomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)

	return err == nil && !info.IsDir()
}

func pathExists(pathname string) bool {
	info, err := os.Stat(pathname)

	return err == nil && info.IsDir()
}

func fileHash(filename string) string {
	var result []byte

	file, err := os.Open(filename)

	if err != nil {
		return ""
	}

	defer file.Close()

	hash := sha512.New512_224()

	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(result))
}

func directoryIsEmpty(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		return false
	}

	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF {
		return true
	}

	return false
}

func unzip(src, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)

	if err != nil {
		return filenames, err
	}

	defer r.Close()

	for _, f := range r.File {
		// Skip directories like __OSX
		if strings.HasPrefix(f.Name, "__") {
			continue
		}

		rc, err := f.Open()

		if err != nil {
			return filenames, err
		}

		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
				return filenames, err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return filenames, err
			}

		}
	}

	return filenames, nil
}

func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func uniqueStrings(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if !m[val] && val != "" {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
