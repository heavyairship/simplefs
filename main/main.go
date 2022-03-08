package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/simplefs/fs"
)

type Fruit struct {
	Name   string
	Weight uint
}

func main() {
	fs := fs.NewFileSystem()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		text := scanner.Text()
		tokens := strings.Fields(text)
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "ls":
			var input string
			if len(tokens) > 1 {
				input = tokens[1]
			} else {
				input = ""
			}
			res, err := fs.List(input)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(res.PrettyPrint())
			}
		case "mkdir":
			if len(tokens) > 1 {
				err := fs.MkDir(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			}
		case "touch":
			if len(tokens) > 1 {
				err := fs.Touch(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			}
		case "cd":
			if len(tokens) > 1 {
				err := fs.ChangeDir(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			}
		case "exit":
			os.Exit(0)
		case "debug":
			fmt.Println(fs.PrettyPrint())
		default:
			fmt.Printf("unknown command: %s\n", tokens[0])
		}
	}
}
