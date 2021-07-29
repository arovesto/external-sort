package random

import (
	"fmt"
	"io"
	"math/rand"
)

var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")

func StringRandomLength(cap int) string {
	randomString := make([]rune, rand.Intn(cap + 1)) // maxStringLength included
	for i := range randomString {
		randomString[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(randomString) + "\n"
}

func Populate(f io.Writer, cap, count int) error {
	for i := 0; i < count; i++ {
		if _, err := f.Write([]byte(StringRandomLength(cap))); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}
	return nil
}