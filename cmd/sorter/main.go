package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arovesto/external-sort/pkg/sort"
)

var fileName string
var allowedLinesInMemory int
var tempDir string

func main() {
	if fileName == "" {
		flag.PrintDefaults()
		fmt.Println("File name should be specified")
		return
	}
	if tempDir == "" {
		var err error
		if tempDir, err = os.MkdirTemp(os.TempDir(), "externalSort*"); err != nil {
			panic(fmt.Errorf("failed to created temp dir: %w", err))
		}
	} else if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.Mkdir(tempDir, 0666); err != nil {
			panic(fmt.Errorf("failed to created providen tmp-dir: %w", err))
		}
	}

	m := sort.FileManager{File: fileName, TempDir: tempDir}
	if err := sort.LinesInPlace(allowedLinesInMemory, m); err != nil {
		panic(fmt.Errorf("failed to run sort: %w", err))
	}
	if err := m.Clear(); err != nil {
		panic(fmt.Errorf("failed to remove temporary files: %w", err))
	}
}

func init()  {
	flag.StringVar(&fileName, "file-name", "", "path to file to sort")
	flag.IntVar(&allowedLinesInMemory, "max-lines-in-memory", 100, "how many lines can be in memory together")
	flag.StringVar(&tempDir, "temp-dir", "", "dir to store temporary files, they will be deleted afterwards")
	flag.Parse()
}


