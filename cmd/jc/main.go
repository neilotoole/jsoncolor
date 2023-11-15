// Package main contains a trivial CLI that accepts JSON input either
// via stdin or via "-i path/to/input.json", and outputs JSON
// to stdout, or if "-o path/to/output.json" is set, outputs to that file.
// If -c (colorized) is true, output to stdout will be colorized if possible
// (but never colorized for file output).
//
// Examples:
//
//	$ cat example.json | jc
//	$ cat example.json | jc -c false
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mattn/go-colorable"
	json "github.com/neilotoole/jsoncolor"
)

var (
	flagPretty     = flag.Bool("p", true, "output pretty JSON")
	flagColorize   = flag.Bool("c", true, "output colorized JSON")
	flagInputFile  = flag.String("i", "", "path to input JSON file")
	flagOutputFile = flag.String("o", "", "path to output JSON file")
)

func printUsage() {
	const msg = `
jc (jsoncolor) is a trivial CLI to demonstrate the neilotoole/jsoncolor package.
It accepts JSON input, and outputs colorized, prettified JSON.

Example Usage:

  # Pipe a JSON file, using defaults (colorized and prettified); print to stdout
  $ cat testdata/sakila_actor.json | jc

  # Read input from a JSON file, print to stdout, DO colorize but DO NOT prettify
  $ jc -c -p=false -i ./testdata/sakila_actor.json 

  # Pipe a JSON input file to jc, outputting to a specified file; and DO NOT prettify
  $ cat ./testdata/sakila_actor.json | jc -p=false -o /tmp/out.json`
	fmt.Fprintln(os.Stderr, msg)
}

func main() {
	flag.Parse()
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printUsage()
		os.Exit(1)
	}
}

func doMain() error {
	var (
		input []byte
		err   error
	)

	if flagInputFile != nil && *flagInputFile != "" {
		// Read from file
		var f *os.File
		if f, err = os.Open(*flagInputFile); err != nil {
			return err
		}
		defer f.Close()

		if input, err = ioutil.ReadAll(f); err != nil {
			return err
		}
	} else {
		// Probably read from stdin...
		var fi os.FileInfo
		if fi, err = os.Stdin.Stat(); err != nil {
			return err
		}

		if (fi.Mode() & os.ModeCharDevice) == 0 {
			// Read from stdin
			if input, err = ioutil.ReadAll(os.Stdin); err != nil {
				return err
			}
		} else {
			return errors.New("invalid args")
		}
	}

	jsn := new(interface{}) // generic interface{} that will hold the parsed JSON
	if err = json.Unmarshal(input, jsn); err != nil {
		return fmt.Errorf("invalid input JSON: %w", err)
	}

	var out io.Writer
	if flagOutputFile != nil && *flagOutputFile != "" {
		// Output file is specified via -o flag
		var fpath string
		if fpath, err = filepath.Abs(*flagOutputFile); err != nil {
			return fmt.Errorf("failed to get absolute path for -o %q: %w", *flagOutputFile, err)
		}

		// Ensure the parent dir exists
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to make parent dir for -o %q: %w", *flagOutputFile, err)
		}

		var f *os.File
		if f, err = os.Create(fpath); err != nil {
			return fmt.Errorf("failed to open output file specified by -o %q: %w", *flagOutputFile, err)
		}
		defer f.Close()
		out = f
	} else {
		// Output file NOT specified via -o flag, use stdout.
		out = os.Stdout
	}

	var enc *json.Encoder

	if flagColorize != nil && *flagColorize && json.IsColorTerminal(out) {
		out = colorable.NewColorable(out.(*os.File)) // colorable is needed for Windows
		enc = json.NewEncoder(out)
		clrs := json.DefaultColors()
		enc.SetColors(clrs)
	} else {
		// We are NOT doing color output: either flag not set, or we
		// could be outputting to a file etc.
		// Therefore DO NOT call enc.SetColors.
		enc = json.NewEncoder(out)
	}

	if flagPretty != nil && *flagPretty {
		// Pretty-print, i.e. set indent
		enc.SetIndent("", "  ")
	}

	return enc.Encode(jsn)
}
