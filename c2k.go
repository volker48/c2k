package main
import (
	"os"
	"bufio"
	"flag"
	"io"
	"log"
)
const (
	defaultDelimiter = "\n"
	delimiterUsage = "Delimiter to split on (defaults to newline)"
	incompleteRead = "c2k: %s: incomplete read"
	noSuchFile = "c2k: %s: no such file"
)

func main() {
	var delimiter string
	flag.StringVar(&delimiter, "delimiter", defaultDelimiter, delimiterUsage)
	flag.Parse()

	bufwrtr := bufio.NewWriter(os.Stdout)
	for _, fileName := range (flag.Args()) {
		handle, err := os.Open(fileName)
		if err != nil {
			log.Printf(noSuchFile, fileName)
			continue
		}
		rdr := bufio.NewReader(handle)
		for {
			line, err := rdr.ReadBytes([]byte(delimiter)[0])
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf(incompleteRead, fileName)
				break
			}
			bufwrtr.Write(line)
		}
		handle.Close()
	}
	bufwrtr.Flush()
}
