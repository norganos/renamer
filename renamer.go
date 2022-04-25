package main

import (
	"fmt"
	"os"
	"strings"
)

type Op struct {
	cmd  string
	args []string
}

func parseCmd(inArgs []string, inOps []Op) (outArgs []string, outOps []Op) {
	args := 0
	cmd := ""
	opt := inArgs[0]
	if opt == "--remove" || opt == "-r" {
		cmd = "remove"
		args = 1
	} else if opt == "--append" || opt == "-a" {
		cmd = "append"
		args = 1
	} else if opt == "--subst" || opt == "-s" {
		cmd = "subst"
		args = 2
	} else {
		fmt.Printf("parse error: don't know what to do with argument %s\n", opt)
		os.Exit(1)
	}
	if len(inArgs) < args+1 {
		fmt.Printf("parse error: %s expects %d arguments\n", opt, args)
		os.Exit(1)
	}
	cmdArgs := inArgs[1 : args+1]
	outOps = append(inOps, Op{cmd, cmdArgs})
	outArgs = inArgs[args+1:]
	return
}

func main() {
	var ops []Op
	args := os.Args[1:]
	for len(args) > 0 {
		if strings.HasPrefix(args[0], "-") {
			args, ops = parseCmd(args, ops)
		} else {
			break
		}
	}
	for i := 0; i < len(args); i++ {
		file := args[i]
		for j := 0; j < len(ops); j++ {
			op := ops[j]
			if op.cmd == "remove" {
				file = strings.ReplaceAll(file, op.args[0], "")
			} else if op.cmd == "append" {
				file = fmt.Sprintf("%s%s", file, op.args[0])
			} else if op.cmd == "subst" {
				file = strings.ReplaceAll(file, op.args[0], op.args[1])
			}
		}
		if file != args[i] {
			fmt.Printf("%s\n", file)
		}
	}
}
