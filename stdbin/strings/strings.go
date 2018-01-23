package strings

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

func Do(input io.Reader, minTextSize uint) *bufio.Scanner {
	outputReader, outputWriter := io.Pipe()
	go searchstrings(input, minTextSize, outputWriter)
	return bufio.NewScanner(outputReader)
}

func searchstrings(input io.Reader, minTextSize uint, output *io.PipeWriter) {

	newline := byte('\n')
	buffer := []byte{}

	writeOnBuffer := func(word string) {
		buffer = append(buffer, []byte(word)...)
	}

	bufferLenInRunes := func() uint {
		return uint(len([]rune(string(buffer))))
	}

	flushBuffer := func() {
		if len(buffer) == 0 {
			return
		}
		if bufferLenInRunes() < minTextSize {
			buffer = nil
			return
		}
		buffer = append(buffer, newline)
		n, err := output.Write(buffer)
		if n != len(buffer) {
			output.CloseWithError(fmt.Errorf("strings:fatal wrote[%d] bytes wanted[%d]\n", n, len(buffer)))
			return
		}
		if err != nil {
			output.CloseWithError(fmt.Errorf("strings:fatal error[%s] writing data\n", err))
			return
		}
		buffer = nil
	}

	data := make([]byte, 1)
	for {
		// WHY: Don't see the point of checking N when reading a single byte
		_, err := input.Read(data)
		if err == io.EOF {
			flushBuffer()
			output.Close()
			return
		}
		if err != nil {
			flushBuffer()
			output.CloseWithError(err)
			return
		}

		if word := string(data); utf8.ValidString(word) {
			writeOnBuffer(word)
		} else if utf8.RuneStart(data[0]) {
			if word, ok := parseNonASCII(input, data[0]); ok {
				writeOnBuffer(word)
			} else {
				flushBuffer()
			}
		} else {
			flushBuffer()
		}
	}
}

func parseNonASCII(input io.Reader, first byte) (string, bool) {
	// TODO: how to test when seems like a rune but it is not
	data := make([]byte, 1)
	// TODO: handle io errors during rune parsing
	input.Read(data)
	// TODO: not handling invalid and other sizes
	possibleWord := string([]byte{first, data[0]})
	return possibleWord, utf8.ValidString(possibleWord)
}
