package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type CmdlineRef struct {
	longOpt     string
	shortOpt    string
	args        []string
	description string
}

type NamedOp struct {
	name   string
	action func(string) string
}
type Operation struct {
	cmdline CmdlineRef
	factory func([]string) NamedOp
}
type Modifier struct {
	cmdline CmdlineRef
	setup   func(string) (string, NamedOp)
}
type Options struct {
	verbose bool
	dryRun  bool
	help    bool
	yes     bool
	copy    bool
}
type Option struct {
	cmdline CmdlineRef
	enable  func(options Options) Options
}

type OperationCollection []Operation

func (container OperationCollection) GetCmdlineRefs() []CmdlineRef {
	var refs []CmdlineRef
	for i := 0; i < len(container); i++ {
		refs = append(refs, container[i].cmdline)
	}
	return refs
}

type ModifierCollection []Modifier

func (container ModifierCollection) GetCmdlineRefs() []CmdlineRef {
	var refs []CmdlineRef
	for i := 0; i < len(container); i++ {
		refs = append(refs, container[i].cmdline)
	}
	return refs
}

type OptionCollection []Option

func (container OptionCollection) GetCmdlineRefs() []CmdlineRef {
	var refs []CmdlineRef
	for i := 0; i < len(container); i++ {
		refs = append(refs, container[i].cmdline)
	}
	return refs
}

func (ref CmdlineRef) Matches(token string) bool {
	return (ref.shortOpt != "" && token == ref.shortOpt) || (ref.longOpt != "" && token == ref.longOpt)
}

func words(inputs ...string) []string {
	return inputs[:]
}
func shorted(input string) string {
	if len(input) > 8 {
		return fmt.Sprintf("%s...", input[0:5])
	}
	return input
}
func named(name, arg string) string {
	return fmt.Sprintf("%s %s", name, shorted(arg))
}

func buildOperations() OperationCollection {
	return append([]Operation{},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--remove",
				shortOpt:    "-r",
				args:        words("<string>"),
				description: "removes all occurrences of <string>",
			},
			factory: func(args []string) NamedOp {
				return NamedOp{
					name: named("remove first", args[0]),
					action: func(input string) string {
						return strings.Replace(input, args[0], "", 1)
					},
				}
			},
		},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--remove-all",
				shortOpt:    "-R",
				args:        words("<string>"),
				description: "removes all occurrences of <string>",
			},
			factory: func(args []string) NamedOp {
				return NamedOp{
					name: named("remove all", args[0]),
					action: func(input string) string {
						return strings.ReplaceAll(input, args[0], "")
					},
				}
			},
		},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--append",
				shortOpt:    "-a",
				args:        words("<string>"),
				description: "appends <string>",
			},
			factory: func(args []string) NamedOp {
				return NamedOp{
					name: named("append", args[0]),
					action: func(input string) string {
						return fmt.Sprintf("%s%s", input, args[0])
					},
				}
			},
		},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--subst",
				shortOpt:    "-s",
				args:        words("<old>", "<new>"),
				description: "replaces the first occurrence of <old> with <new>",
			},
			factory: func(args []string) NamedOp {
				return NamedOp{
					name: named("replace first", args[0]),
					action: func(input string) string {
						return strings.Replace(input, args[0], args[1], 1)
					},
				}
			},
		},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--subst-all",
				shortOpt:    "-S",
				args:        words("<old>", "<new>"),
				description: "replaces all occurrences of <old> with <new>",
			},
			factory: func(args []string) NamedOp {
				return NamedOp{
					name: named("replace all", args[0]),
					action: func(input string) string {
						return strings.ReplaceAll(input, args[0], args[1])
					},
				}
			},
		},
		//Operation{
		//	cmdline: CmdlineRef{
		//		longOpt:     "--regsub",
		//		shortOpt:    "-x",
		//		args:        words("<pattern>", "<replacement>"),
		//		description: "replaces the first match of <pattern> with <replacement>",
		//	},
		//	factory: func(args []string) NamedOp {
		//		re := regexp.MustCompile(args[0])
		//		return NamedOp{
		//			name: named("replace first by regex", args[0]),
		//			action: func(input string) string {
		//				submatch := re.FindStringIndex(input)
		//				if submatch != nil {
		//					return string(re.ExpandString([]byte{}, args[1], input, submatch))
		//				}
		//				return input
		//			},
		//		}
		//	},
		//},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--regsub-all",
				shortOpt:    "-X",
				args:        words("<pattern>", "<replacement>"),
				description: "replaces all matches of <pattern> with <replacement>",
			},
			factory: func(args []string) NamedOp {
				re := regexp.MustCompile(args[0])
				return NamedOp{
					name: named("replace all by regex", args[0]),
					action: func(input string) string {
						return re.ReplaceAllString(input, args[1])
					},
				}
			},
		},
		Operation{
			cmdline: CmdlineRef{
				longOpt:     "--trim",
				shortOpt:    "-t",
				args:        words(),
				description: "remove leading/trailing whitespaces (always implicitly done at last)",
			},
			factory: func(args []string) NamedOp {
				return NamedOp{
					name: "trim",
					action: func(input string) string {
						return strings.Trim(input, " \t")
					},
				}
			},
		},
	)
}

func buildModifiers() ModifierCollection {
	return append([]Modifier{},
		Modifier{
			cmdline: CmdlineRef{
				longOpt:     "--preserve-extension",
				shortOpt:    "-p",
				description: "preserve file extension, pipeline operations act on base filename",
			},
			setup: func(initial string) (string, NamedOp) {
				lastDot := strings.LastIndex(initial, ".")
				extension := initial[lastDot:]
				basename := initial[0:lastDot]
				if lastDot == 0 {
					extension = ""
					basename = initial
				}
				return basename, NamedOp{
					name: named("preserve extension", extension),
					action: func(input string) string {
						return fmt.Sprintf("%s%s", input, extension)
					},
				}
			},
		},
	)
}

func buildOptions() OptionCollection {
	return append([]Option{},
		Option{
			cmdline: CmdlineRef{
				longOpt:     "--help",
				shortOpt:    "-h",
				description: "print out usage help",
			},
			enable: func(opts Options) Options {
				opts.help = true
				return opts
			},
		},
		Option{
			cmdline: CmdlineRef{
				longOpt:     "--verbose",
				shortOpt:    "-v",
				description: "print detailed pipeline operations with intermediate values",
			},
			enable: func(opts Options) Options {
				opts.verbose = true
				return opts
			},
		},
		Option{
			cmdline: CmdlineRef{
				longOpt:     "--dry-run",
				shortOpt:    "-d",
				description: "do not perform any file renames",
			},
			enable: func(opts Options) Options {
				opts.dryRun = true
				return opts
			},
		},
		Option{
			cmdline: CmdlineRef{
				longOpt:     "--yes",
				shortOpt:    "-y",
				description: "do not ask for confirmation before renames",
			},
			enable: func(opts Options) Options {
				opts.yes = true
				return opts
			},
		},
		Option{
			cmdline: CmdlineRef{
				longOpt:     "--copy",
				shortOpt:    "-c",
				description: "copy files instead of moving",
			},
			enable: func(opts Options) Options {
				opts.copy = true
				return opts
			},
		},
	)
}

type Rename struct {
	src string
	dst string
}

func confirm(s string, tries int) bool {
	r := bufio.NewReader(os.Stdin)

	for ; tries > 0; tries-- {
		fmt.Printf("%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Empty input (i.e. "\n")
		if len(res) < 2 {
			continue
		}

		return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
	}

	return false
}

func main() {
	availableOps := buildOperations()
	availableMods := buildModifiers()
	availableOpts := buildOptions()

	var modifiers []func(string) (string, NamedOp)
	var pipeline []NamedOp
	var options Options

	printCmdlineRefs := func(refs []CmdlineRef) {
		padding := 0
		for _, ref := range refs {
			l := 0
			l += len(ref.longOpt)
			if ref.args != nil {
				for j := 0; j < len(ref.args); j++ {
					l += len(ref.args[j]) + 1
				}
			}
			if padding < l {
				padding = l
			}
		}
		format := fmt.Sprintf(" %%s %%-%ds  %%s\n", padding)
		for _, ref := range refs {
			short := "   "
			if ref.shortOpt != "" {
				short = fmt.Sprintf("%s,", ref.shortOpt)
			}
			long := ref.longOpt
			if ref.args != nil {
				if len(ref.args) > 0 {
					long = fmt.Sprintf("%s %s", long, strings.Join(ref.args, " "))
				}
			}
			fmt.Printf(format, short, long, ref.description)
		}
	}

	help := func() {
		fmt.Printf("Utility to batch rename files\n")
		fmt.Printf("Usage: %s [options...] [modifiers...] [operations...] [files...]\n", os.Args[0])
		fmt.Printf("\nOptions:\n")
		printCmdlineRefs(availableOpts.GetCmdlineRefs())
		fmt.Printf("\nModifiers:\n")
		printCmdlineRefs(availableMods.GetCmdlineRefs())
		fmt.Printf("\nOperations:\n")
		printCmdlineRefs(availableOps.GetCmdlineRefs())
	}

	args := os.Args[1:]
	for len(args) > 0 {
		match := false
		for _, mod := range availableMods {
			if mod.cmdline.Matches(args[0]) {
				match = true
				modifiers = append(modifiers, mod.setup)
				break
			}
		}
		for _, opt := range availableOpts {
			if opt.cmdline.Matches(args[0]) {
				match = true
				options = opt.enable(options)
				break
			}
		}
		if !match {
			break
		}
		args = args[1:]
	}
	if options.help {
		help()
		os.Exit(0)
	}
	for len(args) > 0 {
		match := false
		for _, op := range availableOps {
			if op.cmdline.Matches(args[0]) {
				match = true
				if len(args) <= len(op.cmdline.args) {
					fmt.Printf("parse error: %s expects %d arguments\n\n", args[0], len(op.cmdline.args))
					help()
					os.Exit(1)
				}
				pipeline = append(pipeline, op.factory(args[1:len(op.cmdline.args)+1]))
				args = args[len(op.cmdline.args)+1:]
				break
			}
		}
		if !match {
			break
		}
	}

	var renames []Rename
	for k := 0; k < len(args); k++ {
		file := args[k]
		if file == "--" {
			continue
		}
		if strings.HasPrefix(file, "-") {
			fmt.Printf("unknown option/modifier/operation: %s\n\n", args[0])
			help()
			os.Exit(1)
		}
		if options.verbose {
			fmt.Printf("%s\n", file)
		}
		var unwrap []NamedOp
		for _, mod := range modifiers {
			newfile, unwrapper := mod(file)
			file = newfile
			if options.verbose {
				fmt.Printf(" %s -> %s\n", unwrapper.name, file)
			}
			unwrap = append(unwrap, unwrapper)
		}
		for _, op := range pipeline {
			file = op.action(file)
			if options.verbose {
				fmt.Printf(" %s -> %s\n", op.name, file)
			}
		}
		file = strings.Trim(file, " \t")
		for i := len(unwrap) - 1; i >= 0; i-- {
			file = unwrap[i].action(file)
			if options.verbose {
				fmt.Printf(" %s -> %s\n", unwrap[i].name, file)
			}
		}
		if options.verbose {
			fmt.Printf(" => %s\n", file)
		}
		if file != args[k] && file != "" {
			renames = append(renames, Rename{args[k], file})
		}
	}
	ret := 0
	if len(renames) > 0 {
		if !options.dryRun {
			if !options.yes {
				fmt.Printf("will rename %d files:\n", len(renames))
				for _, r := range renames {
					fmt.Printf(" %s -> %s\n", r.src, r.dst)
				}
				if !confirm("Continue?", 1) {
					os.Exit(9)
				}
			}
			if options.verbose {
				fmt.Printf("renaming...\n")
			}
			for _, r := range renames {
				if !options.dryRun {
					fmt.Printf("%s -> %s\n", r.src, r.dst)
					err := os.Rename(r.src, r.dst)
					if err != nil {
						fmt.Printf("error: %s\n", err)
						ret = 5
					}
				}
			}
		} else {
			for _, r := range renames {
				fmt.Printf("%s -> %s\n", r.src, r.dst)
			}
		}
	} else {
		fmt.Printf("nothing to do...")
	}
	os.Exit(ret)
}
