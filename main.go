package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
)

var (
	flagW           = flag.Bool("w", false, "write result to (source) file instead of stdout")
	flagSpacing     = flag.Bool("spacing", false, "add spaces inside {{ }} braces")
	flagIndentStyle = flag.String("indent-style", "none", "indentation style: none, tabs, spaces")
	flagIndentSize  = flag.Int("indent-size", 2, "indentation size")
)

func main() {
	flag.Parse()

	opts := Options{
		Spacing:     *flagSpacing,
		IndentStyle: *flagIndentStyle,
		IndentSize:  *flagIndentSize,
	}

	switch opts.IndentStyle {
	case "none", "tabs", "spaces":
	default:
		log.Fatalf("invalid indent-style: %q (must be none, tabs, or spaces)", opts.IndentStyle)
	}

	args := flag.Args()
	if len(args) == 0 {
		if *flagW {
			log.Fatal("-w requires file arguments")
		}
		if err := Format(os.Stdout, os.Stdin, opts); err != nil {
			log.Fatal(err)
		}
		return
	}

	for _, path := range args {
		if err := processFile(path, opts); err != nil {
			log.Fatal(err)
		}
	}
}

func processFile(path string, opts Options) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = Format(&buf, f, opts)
	f.Close()
	if err != nil {
		return err
	}

	if *flagW {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path, buf.Bytes(), info.Mode().Perm())
	}

	_, err = io.Copy(os.Stdout, &buf)
	return err
}
