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
	Remove(path string) error
	ChangeDir(path string) error
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	PrettyPrint() string
}

type fileSystem struct {
	root *model.INode
	cwd  *model.INode
}

func newDir(parent *model.INode) *model.INode {
	return &model.INode{Type: model.Directory, Parent: parent, Children: make(model.Entries), Contents: nil}
}

func newFile(parent *model.INode) *model.INode {
	return &model.INode{Type: model.File, Parent: parent, Children: nil, Contents: make(model.DataBlocks, 0)}
}

func NewFileSystem() FileSystem {
	root := newDir(nil)
	root.Parent = root
	return &fileSystem{root: root, cwd: root}
}

func (f *fileSystem) List(path string) (model.Entries, error) {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	name := Separator
	for _, component := range pathComponents {
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		result, ok := root.Children[component]
		if !ok {
			return nil, fmt.Errorf("ls: %s: No such file or directory", path)
		}
		root = result
		name = component
	}
	if root.Type == model.Directory {
		return root.Children, nil
	}
	return model.Entries{name: root}, nil
}

func (f *fileSystem) Touch(path string) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for idx, component := range pathComponents {
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		if idx == len(pathComponents)-1 {
			if _, ok := root.Children[component]; !ok {
				root.Children[component] = newFile(root)
			}
			break
		}
		result, ok := root.Children[component]
		if !ok {
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
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		if idx == len(pathComponents)-1 {
			if _, ok := root.Children[component]; ok {
				return fmt.Errorf("mkdir: %s: File exists", path)
			}
		}
		result, ok := root.Children[component]
		if !ok {
			root.Children[component] = newDir(root)
			root = root.Children[component]
		} else {
			root = result
		}

	}
	return nil
}

func (f *fileSystem) Remove(path string) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	name := Separator
	for _, component := range pathComponents {
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		result, ok := root.Children[component]
		if !ok {
			return fmt.Errorf("rm: %s: No such file or directory", path)
		}
		root = result
		name = component
	}
	delete(root.Parent.Children, name)
	return nil
}

func (f *fileSystem) ChangeDir(path string) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for _, component := range pathComponents {
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		result, ok := root.Children[component]
		if !ok {
			return fmt.Errorf("cd: %s: No such file or directory", path)
		}
		if result.Type == model.File {
			return fmt.Errorf("cd: %s: Not a directory", path)
		}
		root = result
	}
	f.cwd = root
	return nil
}

func (f *fileSystem) Write(path string, data []byte) error {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for _, component := range pathComponents {
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		result, ok := root.Children[component]
		if !ok {
			return fmt.Errorf("write: %s: No such file or directory", path)
		}
		root = result
	}
	if root.Type != model.File {
		return fmt.Errorf("write: %s: Cannot write a directory", path)
	}
	bytesWritten := 0
	totalBytes := len(data)
	for bytesWritten < totalBytes {
		block := &model.DataBlock{End: 0}
		bytesToWrite := min(totalBytes-bytesWritten, model.BlockSize)
		block.End = copy(block.Data[:], data[bytesWritten:bytesWritten+bytesToWrite])
		if block.End != bytesToWrite {
			return fmt.Errorf("write: %s: Internal error: could not write to file", path)
		}
		bytesWritten += bytesToWrite
		root.Contents = append(root.Contents, block)
	}
	return nil
}

func (f *fileSystem) Read(path string) ([]byte, error) {
	pathComponents := strings.Split(path, Separator)
	root := f.selectRoot(path)
	for _, component := range pathComponents {
		if component == "" {
			continue
		}
		if component == ".." {
			root = root.Parent
			continue
		}
		result, ok := root.Children[component]
		if !ok {
			return nil, fmt.Errorf("read: %s: No such file or directory", path)
		}
		root = result
	}
	if root.Type != model.File {
		return nil, fmt.Errorf("read: %s: Cannot read a directory", path)
	}
	out := make([]byte, model.BlockSize*len(root.Contents))
	totalBytesRead := 0
	for _, block := range root.Contents {
		bytesRead := copy(out[totalBytesRead:], block.Data[:block.End])
		if bytesRead != block.End {
			return nil, fmt.Errorf("read: %s: Internal error: could not read from file", path)
		}
		totalBytesRead += bytesRead
	}
	return out[:totalBytesRead], nil
}

func (f *fileSystem) PrettyPrint() string {
	f.root.Parent = nil // prevent cycle
	s, err := json.MarshalIndent(f.cwd, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return string(s)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
func (f *fileSystem) selectRoot(path string) *model.INode {
	if strings.HasPrefix(path, Separator) {
		return f.root
	} else {
		return f.cwd
	}
}
