package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"runtime/debug"
)

var (
	flagW           = flag.Bool("w", false, "write result to (source) file instead of stdout")
	flagSpacing     = flag.Bool("spacing", false, "add spaces inside {{ }} braces")
	flagIndentStyle = flag.String("indent-style", "none", "indentation style: none, tabs, spaces")
	flagIndentSize  = flag.Int("indent-size", 2, "indentation size")
	flagVersion     = flag.Bool("version", false, "print version and exit")

	version string
)

func main() {
	flag.Parse()

	if *flagVersion {
		_, _ = io.WriteString(os.Stdout, "gotmplfmt "+resolveVersion(debug.ReadBuildInfo)+"\n")
		return
	}

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
	if opts.IndentSize < 0 {
		log.Fatalf("invalid indent-size: %d (must be non-negative)", opts.IndentSize)
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

func resolveVersion(readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if version != "" {
		return version
	}

	if info, ok := readBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "(devel)"
}
