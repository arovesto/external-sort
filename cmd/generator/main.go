package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/arovesto/external-sort/pkg/random"
)

var fileName string
var maxStringLength int
var howMuchGenerate int
var seed int64

func main()  {
	if seed != 0 {
		rand.Seed(seed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	f := os.Stdout
	if fileName != "" {
		var err error
		if f, err = os.OpenFile(fileName, os.O_CREATE | os.O_WRONLY, 0666); err != nil {
			panic(fmt.Errorf("failed to open target file: %w", err))
		}
	}
	if err := random.Populate(f, maxStringLength, howMuchGenerate); err != nil {
		panic(fmt.Errorf("failed to generate file: %w", err))
	}
}

func init()  {
	flag.StringVar(&fileName, "file-name", "", "file to generate")
	flag.IntVar(&howMuchGenerate, "lines-count", 100, "how many lines to generate")
	flag.IntVar(&maxStringLength, "max-length", 50, "greatest line length")
	flag.Int64Var(&seed, "random-seed", 0, "seed to use create file")
	flag.Parse()
}