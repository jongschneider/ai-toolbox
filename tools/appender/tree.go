package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type caseInsensitive struct {
	values []string
}

func (ci caseInsensitive) Len() int {
	return len(ci.values)
}

func (ci caseInsensitive) Less(i, j int) bool {
	return strings.ToLower(ci.values[i]) < strings.ToLower(ci.values[j])
}

func (ci caseInsensitive) Swap(i, j int) {
	ci.values[i], ci.values[j] = ci.values[j], ci.values[i]
}

func visit(path, indent string, w io.Writer) (dirs, files int, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, 0, fmt.Errorf("stat %s: %w", path, err)
	}

	if _, err := w.Write([]byte(fi.Name() + "\n")); err != nil {
		return 0, 0, fmt.Errorf("write %s: %w", fi.Name(), err)
	}
	if !fi.IsDir() {
		return 0, 1, nil
	}

	dir, err := os.Open(path)
	if err != nil {
		return 1, 0, fmt.Errorf("open %s: %w", path, err)
	}
	names, err := dir.Readdirnames(-1)
	_ = dir.Close() // safe to ignore this error.
	if err != nil {
		return 1, 0, fmt.Errorf("read dir names %s: %w", path, err)
	}
	names = removeHidden(names)
	sort.Sort(caseInsensitive{names})
	add := "│   "
	for i, name := range names {
		if i == len(names)-1 {
			if _, err := w.Write([]byte(indent + "└── ")); err != nil {
				return 0, 0, fmt.Errorf("write %s: %w", name, err)
			}
			add = "    "
		} else {
			if _, err := w.Write([]byte(indent + "├── ")); err != nil {
				return 0, 0, fmt.Errorf("write %s: %w", name, err)
			}
		}
		d, f, err := visit(filepath.Join(path, name), indent+add, w)
		if err != nil {
			log.Println(err)
		}
		dirs, files = dirs+d, files+f
	}
	return dirs + 1, files, nil
}

func removeHidden(files []string) []string {
	var clean []string
	for _, f := range files {
		if f[0] != '.' {
			clean = append(clean, f)
		}
	}
	return clean
}
