package releaser

import (
	"bufio"
	"io"
)

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

func prefixedWriter(logger Logger, p string) io.Writer {
	reader, writer := io.Pipe()
	lineScanner := bufio.NewScanner(reader)
	go func() {
		for lineScanner.Scan() {
			line := lineScanner.Text()
			logger.Printf("[%s] %s", p, line)
		}
	}()
	return writer
}
