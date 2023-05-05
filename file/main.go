package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

func main() {
	filename := "messages.txt"

	// Open the file for reading
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Start reading from the beginning of the file
	offset := int64(0)

	// Read messages until the end of the file is reached
	for {
		// Lock the file before seeking to the current offset
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
		if err != nil {
			panic(err)
		}

		// Set the file pointer to the current offset
		_, err = file.Seek(offset, 0)
		if err != nil {
			panic(err)
		}

		var buf [5]byte
		_, err = file.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				// break
				err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				if err != nil {
					panic(err)
				}
				time.Sleep(time.Second*10)
				continue
			}
			panic(err)
		}

		fmt.Println(string(buf[:]))

		// Unlock the file after reading the message
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		if err != nil {
			panic(err)
		}

		// Update th offset to the end of the message
		offset, err = file.Seek(0, io.SeekCurrent)
		if err != nil {
			panic(err)
		}
	}
}
