package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

const bufferSize = 1024 * 1024

func copyInternal(dest io.Writer, source io.Reader, bytesToCopy int64) error {
	// copy source to dest while printing progress
	// return error if any
	buff := make([]byte, bufferSize)
	offset := 0

	printProgress := func() {
		percentage := float64(offset) / float64(bytesToCopy) * 100
		fmt.Printf("\rCopy progress: %.2f%%", percentage)
	}

	for {
		n, err := source.Read(buff)
		offset += n
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		_, err = dest.Write(buff[:n])
		if err != nil {
			return err
		}
		printProgress()
	}
	printProgress()
	println()
	return nil
}

func Copy(fromPath, toPath string, offset, limit int64) error {
	var source *os.File
	var dest *os.File
	var err error
	var reader io.Reader

	// check if file has size
	fs, err := os.Stat(fromPath)
	if err != nil {
		return err
	}
	if fs.Size() == 0 {
		return ErrUnsupportedFile
	}

	// Open source file
	source, err = os.Open(fromPath)
	if err != nil {
		return err
	}
	defer source.Close()

	if offset > 0 {
		if fs.Size() < offset {
			return ErrOffsetExceedsFileSize
		}
		_, err = source.Seek(offset, io.SeekStart)
		if err != nil {
			return err
		}
	}

	// Open destination file
	dest, err = os.Create(toPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	if limit > 0 {
		reader = io.LimitReader(source, limit)
	} else {
		reader = source
	}

	bytesToCopy := fs.Size() - offset
	if limit > 0 && (limit < bytesToCopy) {
		bytesToCopy = limit
	}
	return copyInternal(dest, reader, bytesToCopy)
}
