package KtFile

import (
	"axj/Kt/Kt"
	"errors"
	"os"
	"path/filepath"
)

func Open(file string) *os.File {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return nil
	}

	Kt.Err(err, true)
	return f
}

func Create(file string, append bool) *os.File {
	err := CreateDir(filepath.Dir(file))
	if err != nil {
		Kt.Err(err, true)
		return nil
	}

	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if append {
		flag |= os.O_APPEND
	}

	f, err := os.OpenFile(file, flag, 0666)
	Kt.Err(err, true)
	return f
}

func CreateDir(dir string) error {
	if dir == "" {
		return nil
	}

	f, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	if err != nil {
		if f == nil {
			return errors.New("Nil Dir " + dir)

		} else if !f.IsDir() {
			return errors.New("Not Dir " + dir)
		}
	}

	return err
}
