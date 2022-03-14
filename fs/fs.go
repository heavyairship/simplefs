package fs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/simplefs/model"
)

type FileSystem interface {
	List(path string) (model.Entries, error)
	Link(src string, dst string) error
	Touch(path string) error
	MakeDir(path string) error
	Remove(path string) error
	ChangeDir(path string) error
	PrintCurrentWorkingDir() (string, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Truncate(path string) error
	PrettyPrint() string
}

type fileSystem struct {
	root *model.INode
	cwd  *model.INode
}

func NewFileSystem() FileSystem {
	root := newDir(nil)
	root.Parent = root
	return &fileSystem{root: root, cwd: root}
}

func (f *fileSystem) List(path string) (model.Entries, error) {
	target, name, err := f.locate(path)
	if err != nil {
		return nil, fmt.Errorf("ls: %s", err)
	}
	if target.Type == model.Directory {
		return target.Children, nil
	}
	return model.Entries{name: target}, nil
}

func (f *fileSystem) Link(srcPath string, dstPath string) error {
	src, _, err := f.locate(srcPath)
	if err != nil {
		return fmt.Errorf("ln: %s", err)
	}
	if src.Type == model.Directory {
		return fmt.Errorf("ln: %s: Is a directory", srcPath)
	}
	dstParent, _, err := f.locate(parentPath(dstPath))
	if err != nil {
		return fmt.Errorf("ln: %s", err)
	}
	if dstParent.Type != model.Directory {
		return fmt.Errorf("ln: %s: Not a directory", dstPath)
	}
	name := basename(dstPath)
	if _, ok := dstParent.Children[name]; ok {
		return fmt.Errorf("ln: %s: File exists", dstPath)
	}
	dstParent.Children[name] = src
	return nil
}

func (f *fileSystem) Touch(path string) error {
	parent, _, err := f.locate(parentPath(path))
	if err != nil {
		return fmt.Errorf("touch: %s", err)
	}
	if parent.Type != model.Directory {
		return fmt.Errorf("touch: %s: Not a directory", path)
	}
	name := basename(path)
	if _, ok := parent.Children[name]; !ok {
		parent.Children[name] = newFile(parent)
	}
	return nil
}

func (f *fileSystem) MakeDir(path string) error {
	pathComponents := strings.Split(path, model.Separator)
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
			if root.Type != model.Directory {
				return fmt.Errorf("mkdir: %s: Not a directory", path)
			}
			root.Children[component] = newDir(root)
			root = root.Children[component]
		} else {
			root = result
		}

	}
	return nil
}

func (f *fileSystem) Remove(path string) error {
	target, name, err := f.locate(path)
	if err != nil {
		return fmt.Errorf("rm: %s", err)
	}
	delete(target.Parent.Children, name)
	return nil
}

func (f *fileSystem) ChangeDir(path string) error {
	target, _, err := f.locate(path)
	if err != nil {
		return fmt.Errorf("cd: %s", err)
	}
	if target.Type != model.Directory {
		return fmt.Errorf("cd: %s: Not a directory", path)
	}
	f.cwd = target
	return nil
}

func (f *fileSystem) PrintCurrentWorkingDir() (string, error) {
	components := []string{}
	curr := f.cwd
	for curr.Parent != curr {
		currName := ""
		for name, child := range curr.Parent.Children {
			if child == curr {
				currName = name
				break
			}
		}
		if currName == "" {
			return "", fmt.Errorf("pwd: Internal error: could not find name for INode in parent")
		}
		components = append(components, currName)
		curr = curr.Parent
	}
	reverse(components)
	out := strings.Join(components, model.Separator)
	if !strings.HasPrefix(out, model.Separator) {
		out = model.Separator + out
	}
	return out, nil
}

func (f *fileSystem) Read(path string) ([]byte, error) {
	target, _, err := f.locate(path)
	if err != nil {
		return nil, fmt.Errorf("read: %s", err)
	}
	if target.Type != model.File {
		return nil, fmt.Errorf("read: %s: Cannot read a directory", path)
	}
	out := make([]byte, model.BlockSize*len(target.Contents))
	totalBytesRead := 0
	for _, block := range target.Contents {
		bytesRead := copy(out[totalBytesRead:], block.Data[:block.End])
		if bytesRead != block.End {
			return nil, fmt.Errorf("read: %s: Internal error: could not read from file", path)
		}
		totalBytesRead += bytesRead
	}
	return out[:totalBytesRead], nil
}

func (f *fileSystem) Write(path string, data []byte) error {
	target, _, err := f.locate(path)
	if err != nil {
		return fmt.Errorf("write: %s", err)
	}
	if target.Type != model.File {
		return fmt.Errorf("write: %s: Cannot write a directory", path)
	}
	bytesWritten := 0
	totalBytes := len(data)
	for bytesWritten < totalBytes {
		var lastBlock *model.DataBlock = nil
		if len(target.Contents) > 0 {
			lastBlock = target.Contents[len(target.Contents)-1]
		}
		var block *model.DataBlock = nil
		if lastBlock != nil && lastBlock.End < model.BlockSize {
			block = lastBlock
		} else {
			block = &model.DataBlock{End: 0}
		}
		bytes := copy(block.Data[block.End:], data[bytesWritten:])
		block.End += bytes
		bytesWritten += bytes
		if block != lastBlock {
			target.Contents = append(target.Contents, block)
		}
	}
	return nil
}

func (f *fileSystem) Truncate(path string) error {
	target, _, err := f.locate(path)
	if err != nil {
		return fmt.Errorf("trunc: %s", err)
	}
	if target.Type != model.File {
		return fmt.Errorf("trunc: %s: Cannot truncate a directory", path)
	}
	target.Contents = nil
	return nil
}

func (f *fileSystem) PrettyPrint() string {
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
	if strings.HasPrefix(path, model.Separator) {
		return f.root
	} else {
		return f.cwd
	}
}

func newDir(parent *model.INode) *model.INode {
	return &model.INode{Type: model.Directory, Parent: parent, Children: make(model.Entries), Contents: nil}
}

func newFile(parent *model.INode) *model.INode {
	return &model.INode{Type: model.File, Parent: parent, Children: nil, Contents: make(model.DataBlocks, 0)}
}

func (f *fileSystem) locate(path string) (*model.INode, string, error) {
	pathComponents := strings.Split(path, model.Separator)
	root := f.selectRoot(path)
	name := model.Separator
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
			return nil, "", fmt.Errorf("%s: No such file or directory", path)
		}
		root = result
		name = component
	}
	return root, name, nil
}

func parentPath(path string) string {
	components := strings.Split(path, model.Separator)
	return strings.Join(components[:len(components)-1], model.Separator)
}

func basename(path string) string {
	components := strings.Split(path, model.Separator)
	return strings.Join(components[len(components)-1:], model.Separator)
}

func reverse(l []string) {
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}
