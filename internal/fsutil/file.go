package fsutil

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Returns true if file exists
func Exists(filename string) bool {
	info, err := os.Stat(filename)

	return err == nil && !info.IsDir()
}

// Returns full path; ~ replaced with actual home directory
func ExpandedFilename(filename string) string {
	if filename == "" {
		panic("filename was empty")
	}

	if len(filename) > 2 && filename[:2] == "~/" {
		if usr, err := user.Current(); err == nil {
			filename = filepath.Join(usr.HomeDir, filename[2:])
		}
	}

	result, err := filepath.Abs(filename)

	if err != nil {
		panic(err)
	}

	return result
}

// Extract Zip file in destination directory
func Unzip(src, dest string) ([]string, error) {

	var fileNames []string

	r, err := zip.OpenReader(src)

	if err != nil {
		return fileNames, err
	}

	defer r.Close()

	for _, f := range r.File {
		// Skip directories like __OSX
		if strings.HasPrefix(f.Name, "__") {
			continue
		}

		rc, err := f.Open()

		if err != nil {
			return fileNames, err
		}

		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		fileNames = append(fileNames, fpath)

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
				return fileNames, err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fileNames, err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return fileNames, err
			}

		}
	}

	return fileNames, nil
}

// Download a file from a URL
func Download(filepath string, url string) (err error) {
	os.MkdirAll("/tmp/photoprism", os.ModePerm)

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
