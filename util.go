package photoprism

import (
	"math/rand"
	"time"
	"os/user"
	"path/filepath"
	"os"
	"crypto/md5"
	"io"
)

func GetRandomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func GetExpandedFilename(filename string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if filename[:2] == "~/" {
		filename = filepath.Join(dir, filename[2:])
	}

	result, _ := filepath.Abs(filename)

	return result
}

func FileExists (filename string) bool {
	_, err := os.Stat(filename)

	return err == nil
}

func Md5Sum (filename string) []byte {
	var result []byte

	file, err := os.Open(filename)

	if err != nil {
		return result
	}

	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return result
	}

	return hash.Sum(result)
}