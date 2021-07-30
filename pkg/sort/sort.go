package sort

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
)


type FileOpener interface {
	// GetTempFile opens existing file, GetTempFile(0, 0) == original file
	GetTempFile(age, num int) (file *bufio.Scanner, closer func(), err error)
	// NewTempFile opens file for write, existing file shall be over written
	NewTempFile(age, num int) (w io.Writer, closer func(), err error)
}

type FileManager struct {
	// File is file path to be sorted
	File    string
	// TempDir is a dir to be used to store temporary files, can be destroyed with FileManager.Clear
	TempDir string
}

func (r FileManager) fileName(age, num int) string {
	if age == 0 && num == 0 {
		return r.File
	}
	return path.Join(r.TempDir, fmt.Sprintf("%d_%d", age, num))
}

func (r FileManager) GetTempFile(age, num int) (file *bufio.Scanner, closer func(), err error) {
	f, err := os.OpenFile(r.fileName(age, num), os.O_RDONLY, 0666)
	if err != nil {
		return nil, nil, err
	}
	closer = func() {
		_ = f.Close()
	}
	return bufio.NewScanner(f), closer,nil
}

func (r FileManager) NewTempFile(age, num int) (w io.Writer, closer func(), err error) {
	f, err := os.OpenFile(r.fileName(age, num), os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		return nil, nil, err
	}
	closer = func() {
		_ = f.Close()
	}
	return f, closer, nil
}

func (r FileManager) Clear() error {
	return os.RemoveAll(r.TempDir)
}

// priorityQueue used to merge k files together, it is based on golang's heap, so we can sort k lines from k files
type priorityQueue []pqElement

type pqElement struct {
	// fileIndex used to get next line from that file (if we can)
	fileIndex int
	// line is actual line (\n removed) read from file
	line string
}

func (p priorityQueue) Len() int  {
	return len(p)
}

func (p priorityQueue) Less(i, j int) bool {
	return p[i].line < p[j].line
}

func (p priorityQueue) Swap(i, j int)  {
	p[i], p[j] = p[j], p[i]
}

func (p *priorityQueue) Push(x interface{}) {
	*p = append(*p, x.(pqElement))
}

func (p *priorityQueue) Pop() interface{} {
	old := *p
	item := old[len(old)-1]
	*p = old[0:len(old)-1]
	return item
}

type fileWithCloser struct {
	f *bufio.Scanner
	c func()
}

// mergeFilesInto merges sorted files into dst while maintaining sorted state via priorityQueue
func mergeFilesInto(files []fileWithCloser, dst io.Writer) error {
	defer func() {
		for _, f := range files {
			f.c()
		}
	}()

	var h priorityQueue
	for i, file := range files {
		if file.f.Scan() {
			h = append(h, pqElement{fileIndex: i, line: file.f.Text()})
		}
	}

	heap.Init(&h)

	for len(h) > 0 {
		el := heap.Pop(&h).(pqElement)
		if _, err := dst.Write([]byte(el.line + "\n")); err != nil {
			return err
		}
		if files[el.fileIndex].f.Scan() {
			heap.Push(&h, pqElement{fileIndex: el.fileIndex,line: files[el.fileIndex].f.Text()})
		}
	}
	return nil
}

// LinesInPlace sorts '\n' terminated strings, while having at maximum maxLinesInMem strings in memory
// It is divides file into maxLinesInMem sized sorted files (sorted in memory)
// then merges maxLinesInMem files at time via mergeFilesInto
// when only one file remained it is copied to starting file which is FileOpener.GetTempFile(0, 0)
// created files removal should be done outside
// time complexity is O(n*log_k(n)*log(k)) while we need to go through file only log_k(n) times
// memory complexity is O(n*log_k(n))
func LinesInPlace(maxLinesInMem int, manager FileOpener) error {
	if maxLinesInMem < 2 {
		return fmt.Errorf("maxLinesInMem should be at least 2 not %d", maxLinesInMem)
	}
	if manager == nil {
		return errors.New("FileOpener and comparator shouldn't be nil")
	}
	original, closer, err := manager.GetTempFile(0, 0)
	if err != nil {
		return err
	}

	buf := make([]string, maxLinesInMem)
	pos := 0
	var currentLevelFiles int

	dumpBuf := func() error {
		file, c, err := manager.NewTempFile(1, currentLevelFiles)
		if err != nil {
			return err
		}
		sort.Strings(buf)
		for _, l := range buf {
			_, err := file.Write([]byte(l + "\n"))
			if err != nil {
				return err
			}
		}
		pos = 0
		c()
		currentLevelFiles++
		return nil
	}

	// Create n / k temporary k-sized sorted files
	for original.Scan() {
		if pos + 1 >= maxLinesInMem {
			if err := dumpBuf(); err != nil {
				return err
			}
		}
		buf[pos] = original.Text()
		pos++
	}
	if len(buf) > 0 {
		if err := dumpBuf(); err != nil {
			return err
		}
	}
	closer()

	if currentLevelFiles == 0 {
		// already sorted
		return nil
	}

	age := 1
	newCurrentFiles := 0
	files := make([]fileWithCloser, maxLinesInMem)
	pos = 0

	mergeFiles := func() (err error) {
		var f io.Writer
		var c func()

		if currentLevelFiles <= maxLinesInMem {
			f, c, err = manager.NewTempFile(0, 0)
		} else {
			f, c, err = manager.NewTempFile(age+1, newCurrentFiles)
			newCurrentFiles++
		}
		if err != nil {
			return err
		}
		if err := mergeFilesInto(files, f); err != nil {
			return err
		}
		c()
		pos = 0
		return nil
	}

	// merge k files at the time until one left
	for currentLevelFiles > 1 {
		for i := 0; i < currentLevelFiles; i++ {
			if pos + 1 >= maxLinesInMem {
				if err := mergeFiles(); err != nil {
					return err
				}
			}
			f, c, err := manager.GetTempFile(age, i)
			if err != nil {
				return err
			}
			files[pos] = fileWithCloser{f:f, c:c}
			pos++
		}
		if len(files) > 0 {
			if err := mergeFiles(); err != nil {
				return err
			}
		}
		if currentLevelFiles <= maxLinesInMem {
			return nil
		}
		age++
		currentLevelFiles = newCurrentFiles
		newCurrentFiles = 0
	}



	// move single last file into old one
	dst, dstCloser, err := manager.NewTempFile(0, 0)
	if err != nil {
		return err
	}
	defer dstCloser()
	src, srcCloser, err := manager.GetTempFile(age, 0)
	if err != nil {
		return err
	}
	defer srcCloser()
	for src.Scan() {
		if _, err := dst.Write([]byte(src.Text() + "\n")); err != nil {
			return err
		}
	}
	return nil
}
