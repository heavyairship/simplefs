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
		case "ln":
			if len(tokens) > 2 {
				err := fs.Link(tokens[1], tokens[2])
				if err != nil {
					fmt.Println(err)
				}
			} else if len(tokens) > 1 {
				fmt.Println("ln: Missing argument <dst>")
			} else {
				fmt.Println("ln: Missing arguments <src> <dst>")
			}
		case "mkdir":
			if len(tokens) > 1 {
				err := fs.MakeDir(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("mkdir: Missing argument <path>")
			}
		case "touch":
			if len(tokens) > 1 {
				err := fs.Touch(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("touch: Missing argument <path>")
			}
		case "rm":
			if len(tokens) > 1 {
				err := fs.Remove(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("rm: Missing argument <path>")
			}
		case "cd":
			if len(tokens) > 1 {
				err := fs.ChangeDir(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
			}
		case "pwd":
			res, err := fs.PrintCurrentWorkingDir()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(res)
			}
		case "write":
			if len(tokens) > 2 {
				err := fs.Write(tokens[1], []byte(strings.Join(tokens[2:], "")))
				if err != nil {
					fmt.Println(err)
				}
			} else if len(tokens) > 1 {
				fmt.Println("write: Missing argument <data>")
			} else {
				fmt.Println("write: Missing arguments <path> <data>")
			}
		case "read":
			if len(tokens) > 1 {
				res, err := fs.Read(tokens[1])
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(string(res))
				}
			} else {
				fmt.Println("read: Missing argument <path>")
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
