package fileutil

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
)

func WriteFile(filename string, content interface{}) error {
	var src io.Reader
	switch content.(type) {
	case string:
		src = strings.NewReader(content.(string))
	case []byte:
		src = bytes.NewReader(content.([]byte))
	case io.ReadCloser:
		src = content.(io.ReadCloser)
	default:
		return errors.New("invalid content type")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	fileWrite := bufio.NewWriter(f)
	_, err = io.Copy(fileWrite, src)
	if err != nil {
		return err
	}
	return fileWrite.Flush()
}
