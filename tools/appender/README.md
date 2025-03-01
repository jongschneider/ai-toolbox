# Appender

Appender is a CLI tool designed to make interacting with Chat LLMs like Claude.ai easier when writing code.
The goal is to capture the filename and contents of files in to feed into a prompt in order to give the current state of the codebase we are working in.

## TODO

[X] add validation and don't select binaries
[] add tests - [teatest](https://github.com/charmbracelet/x/tree/main/exp/teatest) - https://github.com/charmbracelet/bubbletea/blob/main/examples/simple/main_test.go
[] add cohesive styling

## Expected Behavior

1. cli

```sh
$ appender
```

    - when called from the command line without any arguments, `appender` will assume that the `workDir` is the current working directory
    - if called with an argument, appender will use the path as the `workDir`

```sh
$ appender ./path/to/dir
```

2. the output of the cli will be the directory tree

```sh
$ appender
.
├── README.md
├── flake.lock
├── flake.nix
├── go.mod
├── go.work
├── go.work.sum
├── justfile
├──+ tools
│   └─+ appender
│       ├── README.md
│       ├── go.mod
│       ├── go.sum
│       └── main.go
└── vendor
```

3. The user should be able to navigate up and down each line using `j` and `k` and select the line using `space`. If a file is selected, it should have a  added to the end of the line like the following (assume README.md was selected using the `space` key):

```sh
$ appender
.
├── README.md 
├── flake.lock
├── flake.nix
├── go.mod
├── go.work
├── go.work.sum
├── justfile
├──+ tools
│   └─+ appender
│       ├── README.md
│       ├── go.mod
│       ├── go.sum
│       └── main.go
└── vendor
```

4. If a directory is selected, then the directory and all files and folders within the directory should have a  added to the end of the line like the following (assume tools was selected using the `space` key):

```sh
$ appender
.
├── README.md
├── flake.lock
├── flake.nix
├── go.mod
├── go.work
├── go.work.sum
├── justfile
├──+ tools 
│   └─+ appender 
│       ├── README.md 
│       ├── go.mod 
│       ├── go.sum 
│       └── main.go 
└── vendor
```

5. Lines with 󱞣 indicate a directory that is expanded. That will be the default. If the user presses `l` or `h` key on a line with a directory, then the 󱞣 will turn into  and all of the directories and files within that directory will disappear. Pressing `l` or `h` key will toggle them back into view and return the  to 󱞣 (see the vendor line as a directory that is not in expanded mode):

```sh
$ appender
.
├── README.md
├── flake.lock
├── flake.nix
├── go.mod
├── go.work
├── go.work.sum
├── justfile
├──+ tools 
│   └─+ appender 
│       ├── README.md 
│       ├── go.mod 
│       ├── go.sum 
│       └── main.go 
└── vendor
```

6. When the user presses `enter` key, the program will register the selected files and create a new file called `output.txt` that includes each selected file along with it's relative path in a comment.

```sh
$ appender
.
├── README.md
├── flake.lock
├── flake.nix
├── go.mod
├── go.work
├── go.work.sum
├── justfile
├──+ tools 
│   └─+ appender 
│       ├── README.md 
│       └── main.go 
└── vendor
```

Will produce a file called `output.txt` with the following contents:

```txt
# tools/appender/README.md
This is some text
    * Bullet 1
    * Bullet #2

# tools/appender/main.go
package main

import "fmt"

func main(){
    fmt.Println("Hello, World!")
}
```
