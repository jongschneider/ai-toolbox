package main

import (
	"os"
	"testing"
)

func Test_visit(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		indent    string
		wantDirs  int
		wantFiles int
		wantErr   bool
	}{
		{
			name:      "testdata",
			path:      "./testdata",
			indent:    "",
			wantDirs:  0,
			wantFiles: 0,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := os.Stdout
			gotDirs, gotFiles, err := visit(tt.path, tt.indent, w)
			if (err != nil) != tt.wantErr {
				t.Errorf("visit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDirs != tt.wantDirs {
				t.Errorf("visit() gotDirs = %v, want %v", gotDirs, tt.wantDirs)
			}
			if gotFiles != tt.wantFiles {
				t.Errorf("visit() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}
