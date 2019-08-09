package osutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func CreateTmpFile(prefix string, contents string) (string, error) {
	f, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", fmt.Errorf("error creating temp file (%s)", err)
	}
	defer f.Close()

	_, err = io.WriteString(f, contents)
	if err != nil {
		return f.Name(), fmt.Errorf("error writing to temp file %s (%s)", f.Name(), err)
	}

	return f.Name(), nil
}

func ReadFile(filename string) (string, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file %s (%s)", filename, err)
	}

	return string(bs), nil
}

func ReadAndDeleteFile(filename string) (string, error) {
	s, err := ReadFile(filename)
	if err != nil {
		return "", err
	}

	err = os.Remove(filename)
	if err != nil {
		return s, fmt.Errorf("error deleting file %s (%s)", filename, err)
	}

	return s, nil
}
