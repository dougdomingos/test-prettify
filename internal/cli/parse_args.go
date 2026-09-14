package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ResolveInput determines input source from CLI args or stdin. It returns
// the stream reader for the provided resource and a cleanup function.
// 
// If provided with one argument, returns a file reader. If stdin is piped,
// uses stdin. Otherwise returns error.
func ResolveInput(args []string, stdin io.Reader) (io.Reader, func(), error) {
	if len(args) == 1 {
		file, err := os.Open(args[0])
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to open file: %w", err)
		}
		return file, func() { _ = file.Close() }, nil
	}

	if isPiped(stdin) {
		return stdin, func() {}, nil
	}

	return nil, func() {}, errors.New("no input provided: pass file path or pipe data to stdin")
}

// isPiped checks whether reader is connected to piped stdin, not terminal.
func isPiped(r io.Reader) bool {
	file, ok := r.(*os.File)
	if !ok {
		return r != nil
	}
	
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	
	return (stat.Mode() & os.ModeCharDevice) == 0
}
