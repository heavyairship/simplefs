package fs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/simplefs/model"
)

const Separator = "/"

type FileSystem interface {
	List(path string) (model.Entries, error)
	Touch(path string) error
	MkDir(path string) error
	ChangeDir(path string) error
	PrettyPrint() string
}

type fileSystem struct {
	root *model.INode
	cwd  *model.INode
}

func newDir() *model.INode {
	return &model.INode{Type: model.Directory, Children: make(model.Entries), Data: nil}
}

func newFile() *model.INode {
	return &model.INode{Type: model.File, Children: nil, Data: make(model.DataBlocks, 0)}
}

func NewFileSystem() FileSystem {
	root := newDir()
	return &fileSystem{root: root, cwd: root}
}

func (f *fileSystem) List(path string) (model.Entries, error) {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for idx, component := range pathComponents {

		// Skip empty components
		if component == "" {
			continue
		}

		// List the file/directory
		if idx == len(pathComponents)-1 {
			if child, ok := root.Children[component]; ok {
				if child.Type == model.Directory {
					return child.Children, nil
				} else {
					return model.Entries{component: child}, nil
				}
			}

		}

		result, ok := root.Children[component]
		if !ok {
			// Error out b/c the directory doesn't contain the component
			return nil, fmt.Errorf("ls: %s: No such file or directory", path)
		}
		root = result
	}
	return root.Children, nil
}

func (f *fileSystem) Touch(path string) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for idx, component := range pathComponents {

		// Skip empty components
		if component == "" {
			continue
		}

		// Create a new file
		if idx == len(pathComponents)-1 {
			if _, ok := root.Children[component]; !ok {
				root.Children[component] = newFile()
			}
			break
		}

		result, ok := root.Children[component]
		if !ok {
			// Error out b/c the directory doesn't contain the component
			return fmt.Errorf("touch: %s: No such file or directory", path)
		}
		root = result
	}

	return nil
}

func (f *fileSystem) MkDir(path string) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for idx, component := range pathComponents {

		// Skip empty components
		if component == "" {
			continue
		}

		if idx == len(pathComponents)-1 {
			if _, ok := root.Children[component]; ok {
				// Error out b/c the file already exists
				return fmt.Errorf("mkdir: %s: File exists", path)
			}
		}

		// Create a new dir
		result, ok := root.Children[component]
		if !ok {
			root.Children[component] = newDir()
			root = root.Children[component]
		} else {
			root = result
		}

	}
	return nil
}

func (f *fileSystem) ChangeDir(path string) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for _, component := range pathComponents {
		// Skip empty components
		if component == "" {
			continue
		}

		result, ok := root.Children[component]
		if !ok {
			// Error out b/c the directory doesn't contain the component
			return fmt.Errorf("cd: %s: No such file or directory", path)
		}
		if result.Type == model.File {
			// Error out b/c the INode is not a directory
			return fmt.Errorf("cd: %s: Not a directory", path)
		}
		root = result
	}
	f.cwd = root
	return nil
}

func (f *fileSystem) PrettyPrint() string {
	s1, err := json.MarshalIndent(f.root, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	s2, err := json.MarshalIndent(f.cwd, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return string(strings.Join([]string{string(s1), string(s2)}, "\n"))
}

func (f *fileSystem) selectRoot(path string) *model.INode {
	if strings.HasPrefix(path, Separator) {
		return f.root
	} else {
		return f.cwd
	}
}
