package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_visitNode(t *testing.T) {
	tests := []struct {
		name   string
		node   *FileNode
		prefix string
		expect func(t *testing.T, node *FileNode, err error)
	}{
		{
			name: "./testdata",
			node: &FileNode{
				name:     "./testdata",
				path:     "./testdata",
				isDir:    true,
				isRoot:   true,
				expanded: true,
			},
			prefix: "",
			expect: func(t *testing.T, node *FileNode, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := visitNode(tt.node, tt.prefix, true)
			tt.expect(t, tt.node, err)

			nodes := tt.node.flatten()

			for _, n := range nodes {
				fmt.Println(n.String()) //nolint:forbidigo
			}
		})
	}
}
