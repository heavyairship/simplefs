package model

import (
	"encoding/json"
	"fmt"
)

type IType string

const (
	Directory IType = "directory"
	File      IType = "file"
)

type DataBlock struct {
	Data [1024]byte
	End  uint
}

type Entries map[string]*INode
type DataBlocks []*DataBlock

type INode struct {
	Type     IType
	Children Entries
	Data     DataBlocks
}

func toStrings(entries Entries) []string {
	out := make([]string, 0, len(entries))
	for e := range entries {
		out = append(out, e)
	}
	return out
}

func (e *Entries) PrettyPrint() string {
	s, err := json.MarshalIndent(toStrings(*e), "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return string(s)
}

func (i *INode) PrettyPrint() string {
	s, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return string(s)
}
