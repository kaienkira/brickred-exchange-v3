package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, ""+
		"brickred exchange compiler\n"+
		"usage: %s "+
		"-f <protocol_file> "+
		"-l <language>"+
		"\n"+
		"    [-o <output_dir>]\n"+
		"    [-I <search_path>]\n"+
		"    [-n <new_line_type>] (unix|dos) default is unix\n"+
		"language supported: cpp php csharp\n",
		filepath.Base(os.Args[0]))
}

func main() {
	optProtoFilePath := flag.String("f", "", "")
	optLanguage := flag.String("l", "", "")
	flag.Parse()

	if *optProtoFilePath == "" || *optLanguage == "" {
		printUsage()
		os.Exit(1)
	}
}
