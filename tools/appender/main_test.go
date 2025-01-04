//nolint:forbidigo
package main

import (
	"fmt"
	"testing"
)

func Test_buildFileTree2(t *testing.T) {
	tests := []struct {
		path    string
		indent  string
		isRoot  bool
		name    string
		wantErr bool
	}{
		{
			path:    "./testdata",
			indent:  "",
			isRoot:  true,
			name:    "basecase",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildFileTree2(tt.path, tt.indent, tt.isRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildFileTree2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			initialModel := &model{
				workDir:  tt.path,
				rootNode: got,
				cursor:   0,
			}
			initialModel.flattenTree()
			for _, n := range initialModel.flatNodes {
				fmt.Println(n.prefix, n.name)
			}
		})
	}
}
