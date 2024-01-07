package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sa6mwa/audiobook-chapter-splitter/fflame"
)

func main() {
	var title string

	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "input.m4b output_directory")
		fmt.Fprintln(os.Stderr, "")
		flag.PrintDefaults()
	}

	flag.StringVar(&title, "t", title, "Title. If empty, title tag from metadata or basename of input file will be used")

	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	meta, err := fflame.GetMeta(flag.Args()[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := fflame.Encode(flag.Args()[0], flag.Args()[1], title, meta); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("OK")
}
