package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IType string

const BlockSize = 8
const Separator = "/"

const (
	Directory IType = "directory"
	File      IType = "file"
)

type DataBlock struct {
	Data [BlockSize]byte
	End  int
}

type Entries map[string]*INode
type DataBlocks []*DataBlock

type INode struct {
	Type     IType
	Parent   *INode `json:"-"`
	Children Entries
	Contents DataBlocks
}

func toStrings(entries Entries) []string {
	out := make([]string, 0, len(entries))
	for e := range entries {
		out = append(out, e)
	}
	return out
}

func (e Entries) PrettyPrint() string {
	names := []string{}
	for name, entry := range e {
		suffix := ""
		if entry.Type == Directory {
			suffix = Separator
		}
		names = append(names, name+suffix)
	}
	out := strings.Join(names, "\n")
	return out
}

func (i *INode) PrettyPrint() string {
	s, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return string(s)
}
