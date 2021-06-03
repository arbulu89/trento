package internal

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"regexp"
	"strings"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func Md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Get data from key = value kind of texts
func FindMatches(pattern string, text []byte) map[string]interface{} {
	configMap := make(map[string]interface{})

	r := regexp.MustCompile(pattern)
	values := r.FindAllStringSubmatch(string(text), -1)
	for _, match := range values {
		// I was not able to create a regular expression which excludes lines starting with #...
		if strings.HasPrefix(match[1], "#") {
			continue
		}
		configMap[match[1]] = match[2]
	}
	return configMap
}
