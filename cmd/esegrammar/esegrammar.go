package main

import (
	"fmt"
	"os"

	"github.com/freshwebio/esengine/grammar"
	"github.com/namsral/flag"
)

func main() {
	if len(os.Args) == 1 {
		usage()
	} else {
		switch os.Args[1] {
		case "build":
			build()
		default:
			usage()
		}
	}
}

func usage() {
	fmt.Print(`
`)
}

func build() {
	outputFile := flag.String("output", "", "The target file containing the generated symbols and parse table")
	grammarFile := flag.String("grammar", "", "The file containing the grammar")
	grammarFmt := flag.String("format", "yaml", "The storage format of the input grammar")
	pkg := flag.String("package", "", "The go package the file's contents will belong to")
	// Exclude build from the arguments that are parsed, otherwise no arguments
	// will be parsed.
	flag.CommandLine.Parse(os.Args[2:])
	grammar.Build(*grammarFile, *grammarFmt, *outputFile, *pkg)
}
