// Package filewrapper wrap working with file storage
package filewrapper

import (
	"bufio"
	"encoding/gob"
	"io"
	"os"

	"github.com/grishagavrin/link-shortener/internal/errs"
)

// Write data to path
func Write(path string, data interface{}) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	// Handle for file close
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(errs.ErrFileStorageNotClose)
		}
	}(f)

	// Convert to gob
	buffer := bufio.NewWriter(f)
	ge := gob.NewEncoder(buffer)
	// encode
	if err := ge.Encode(data); err != nil {
		return err
	}
	_ = buffer.Flush()
	return nil
}

// Read data from path to data variable
func Read(path string, data interface{}) error {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	// handle for file close
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(errs.ErrFileStorageNotClose)
		}
	}(f)
	gd := gob.NewDecoder(f)
	if err := gd.Decode(data); err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}
