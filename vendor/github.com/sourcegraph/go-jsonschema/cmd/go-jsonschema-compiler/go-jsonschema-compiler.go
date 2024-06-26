// The go-jsonschema-compiler generates Go types from JSON Schemas. The Go types hold instances of
// the JSON Schemas and can marshal/unmarshal to/from JSON using encoding/json.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"os"

	"github.com/nxpkg/go-jsonschema/compiler"
	"github.com/nxpkg/go-jsonschema/jsonschema"
)

var (
	packageName = flag.String("pkg", "schema", "Go package name to use in emitted source code")
	outputFile  = flag.String("o", "", "write result to file instead of stdout")
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\tgo-jsonschema-compiler [flags] files...")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "go-jsonschema-compiler: no JSON Schema files listed.")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(2)
	}

	schemas := make([]*jsonschema.Schema, flag.NArg())
	for i, filename := range flag.Args() {
		var err error
		schemas[i], err = readSchema(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-jsonschema-compiler: error reading JSON Schema from %s: %s.\n", filename, err)
			os.Exit(2)
		}
	}

	decls, imports, err := compiler.Compile(schemas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-jsonschema-compiler: compilation error: %s.\n", err)
		os.Exit(2)
	}
	var buf bytes.Buffer
	file := &ast.File{
		Name:    ast.NewIdent(*packageName),
		Imports: imports,
		Decls:   decls,
	}
	if err := format.Node(&buf, token.NewFileSet(), file); err != nil {
		fmt.Fprintf(os.Stderr, "go-jsonschema-compiler: code formatting error: %s.\n", err)
		os.Exit(2)
	}
	out := buf.Bytes()
	if !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, '\n')
	}

	var outFile io.WriteCloser
	if *outputFile == "" {
		outFile = os.Stdout
	} else {
		var err error
		outFile, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-jsonschema-compiler: output error: %s.\n", err)
			os.Exit(2)
		}
	}
	defer outFile.Close()

	fmt.Fprintln(outFile, "// Code generated by go-jsonschema-compiler. DO NOT EDIT.")
	fmt.Fprintln(outFile)
	outFile.Write(out)
}

func readSchema(filename string) (*jsonschema.Schema, error) {
	var f io.ReadCloser
	if filename == "-" {
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	var schema *jsonschema.Schema
	if err := json.NewDecoder(f).Decode(&schema); err != nil {
		return nil, err
	}
	return schema, nil
}
