Here's a detailed prompt that would help an AI generate the same tree command implementation:

````
Create a Go implementation of the `tree` command using the following FileNode struct:

```go
type FileNode struct {
    name     string // name represents the name of the file or directory
    path     string // path represents the full path of the file or directory
    isDir    bool   // isDir is used to identify directories
    isRoot   bool   // isRoot is only used to identify the root node.
    expanded bool   // expanded is used to show/hide the children of a directory
    selected bool
    prefix   string      // prefix is used in the View method to draw the tree structure
    children []*FileNode // includes directories and files
}

func (node *FileNode) String() string {
    dirIndicator := ""
    if node.isDir && !node.isRoot {
        if node.expanded {
            dirIndicator = " "
        } else {
            dirIndicator = " "
        }
    }

    selected := ""
    if node.selected {
        selected = "  "
    }

    return fmt.Sprintf("%s%s%s%s", node.prefix, dirIndicator, node.name, selected)
}
````

Please implement two functions:

1. A recursive `visit` function with this signature that builds the tree:

```go
func visit(node *FileNode, prefix string) (err error)
```

Requirements for `visit`:

- Should read the filesystem using os.ReadDir
- Should populate the children field of each FileNode
- Use standard tree command characters for visualization:
  - ├── for intermediate items
  - └── for last items
  - │ for vertical lines
- Handle errors by logging them but continue processing
- The prefix parameter is used to build the tree structure display
- The initial call will be with the root node and an empty prefix

2. A method to flatten the tree for display:

```go
func (node *FileNode) flattenNode() []*FileNode
```

Requirements for `flattenNode`:

- Should return a slice of nodes in display order
- Should only include children of directories that are marked as expanded
- Should include the current node in the result
- Should recursively process all children of expanded directories

The goal is to be able to:

1. Call `visit()` to build the tree
2. Call `flattenNode()` to get a displayable list
3. Iterate through the flattened list calling String() on each node to display the tree

The implementation should prioritize clarity and maintainability.

```

This prompt provides:
1. The complete starting struct and String() method
2. Clear function signatures
3. Specific requirements for each function
4. Visual characters to use for the tree structure
5. The expected workflow for using the implementation
6. Clear error handling requirements

Would you like me to modify the prompt in any way?
```
