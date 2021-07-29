package sort

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/arovesto/external-sort/pkg/random"
)


type mockManager struct {
	files map[string]*bytes.Buffer
}

func newMockFileManager(c string) *mockManager {
	return &mockManager{
		files: map[string]*bytes.Buffer{"0_0" : bytes.NewBuffer([]byte(c))},
	}
}

func (m mockManager) getResult() string {
	return m.files[m.getName(0, 0)].String()
}

func (m mockManager) getName(age, num int) string {
	return fmt.Sprintf("%d_%d", age, num)
}

func (m mockManager) GetTempFile(age, num int) (*bufio.Scanner, func(), error) {
	return bufio.NewScanner(m.files[m.getName(age, num)]), func() {}, nil
}

func (m mockManager) NewTempFile(age, num int) (io.Writer, func(), error) {
	m.files[m.getName(age, num)] = &bytes.Buffer{}
	return m.files[m.getName(age, num)], func() {}, nil
}

func naiveSort(content string) string {
	// golang's strings.Split() adds bonus unwonted empty line if last symbol is \n
	// we ourself consider lines \n-terminated, so we need to fix that
	last := ""
	if len(content) > 1 && content[len(content) - 1] == '\n' && strings.Trim(content, "\n") != "" {
		content = content[:len(content) - 1]
		last = "\n"
	}
	lines := strings.Split(content, "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n") + last
}

const bigTest = `a
b
c
d
1
2
v
crewr
ewrwer
cv
some
body
once
told me
the world is gonna
roll me
i ant the sharpest tool
in the shed
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARandomstringAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAdontread
`

func TestSort(t *testing.T) {
	cases := []string {
		"",
		"\n",
		"\n\n",
		"1\n",
		"1\n2\n",
		"2\n111\n",
		"9\n8\n81\n7\n6\n",
		bigTest,
	}
	for linesInMem := 2; linesInMem < 20; linesInMem++ {
		for _, c := range cases {
			naive := naiveSort(c)
			m := newMockFileManager(c)
			if err := LinesInPlace(linesInMem, m); err != nil {
				t.Error("sort failed", err)
				return
			}
			if naive != m.getResult() {
				t.Errorf("failed to run sort test on %d in mem and '%s' string. Got '%s', want '%s'", linesInMem, c, m.getResult(), naive)
				return
			}
		}
	}
}

const lineMaxLength = 50
const lineCount = 10000
const lineMemory = lineCount / 10

func BenchmarkSort(b *testing.B) {
	var dirs []string
	defer func() {
		b.StopTimer()
		for _, d := range dirs {
			if err := os.RemoveAll(d); err != nil {
				log.Println(err)
			}
		}
	}()

	runTest := func() {
		b.StopTimer()
		dir, err := os.MkdirTemp(".", "test*")
		if err != nil {
			b.Fatal("failed to create dir")
		}
		dirs = append(dirs, dir)
		file, err := os.CreateTemp(dir, "test")
		if err != nil {
			b.Fatal("failed to create file")
		}
		m := FileManager{
			File:    file.Name(),
			TempDir: dir,
		}
		_ = file.Close()
		writer, _, _ := m.NewTempFile(0, 0)
		if err := random.Populate(writer, lineMaxLength, lineCount); err != nil {
			b.Error("populate failed")
			return
		}
		b.StartTimer()
		if err := LinesInPlace(lineMemory, m); err != nil {
			b.Error("sort failed")
			return
		}
	}

	for i := 0; i < b.N; i++ {
		runTest()
	}
}
