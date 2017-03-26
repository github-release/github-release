package github

import (
	"os"
	"testing"
)

func TestGetFileSize(t *testing.T) {
	const fname = "file.go"

	methods := []func(f *os.File) (int64, error){
		GetFileSize,
		fsizeStat,
		fsizeSeek,
	}
	results := make([]int64, len(methods))
	for i, fn := range methods {
		f, err := os.Open(fname)
		if err != nil {
			t.Fatal(i, err)
		}
		defer f.Close()

		size, err := fn(f)
		if err != nil {
			t.Fatal(i, err)
		}
		if size == 0 {
			t.Fatal(i, "file size of ", fname, " was zero")
		}

		results[i] = size
	}

	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Error(i, "method does not correspond", results[i], "!=", results[0])
		}
	}
}
