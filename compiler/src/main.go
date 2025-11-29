package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
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
	// parse command line options
	var optProtoFilePath string
	var optLanguage string
	var optOutputDir string
	var optSearchPath []string
	var optNewLineType string

	flagSet := flag.NewFlagSet("main", flag.ContinueOnError)
	flagSet.StringVarP(&optProtoFilePath, "-proto_file_path", "f", "", "")
	flagSet.StringVarP(&optLanguage, "-language", "l", "", "")
	flagSet.StringVarP(&optOutputDir, "-output_dir", "o", "", "")
	flagSet.StringSliceVarP(&optSearchPath, "-search_path", "I", []string{}, "")
	flagSet.StringVarP(&optNewLineType, "-new_line_type", "n", "", "")
	flagSet.Parse(os.Args[1:])

	// check command line options
	// -- required options
	if optProtoFilePath == "" ||
		optLanguage == "" {
		printUsage()
		os.Exit(1)
	}
	// -- option default value
	if optOutputDir == "" {
		optOutputDir = "."
	}
	if optNewLineType == "" {
		optNewLineType = "unix"
	}

	// -- check option proto_file_path
	if utilCheckFileExists(optProtoFilePath) == false {
		fmt.Fprintf(os.Stderr,
			"error: can not find protocol file `%s`\n",
			optProtoFilePath)
		os.Exit(1)
	}

	// -- check option language
	if optLanguage != "cpp" &&
		optLanguage != "php" &&
		optLanguage != "csharp" {
		fmt.Fprintf(os.Stderr,
			"error: language `%s` is not supported\n",
			optLanguage)
		os.Exit(1)
	}

	// -- check option output_dir
	if utilCheckDirExists(optOutputDir) == false {
		fmt.Fprintf(os.Stderr,
			"error: can not find output directory `%s`\n",
			optOutputDir)
		os.Exit(1)
	}

	// -- check option new_line_type
	if optNewLineType != "dos" &&
		optNewLineType != "unix" {
		fmt.Fprintf(os.Stderr,
			"error: new_line_type `%s` is invalid\n",
			optNewLineType)
		os.Exit(1)
	}

	os.Exit(0)
}
