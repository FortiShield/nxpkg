// +build tools

package main

import (
	_ "github.com/google/zoekt/cmd/zoekt-archive-index"
	_ "github.com/google/zoekt/cmd/zoekt-nxpkg-indexserver"
	_ "github.com/google/zoekt/cmd/zoekt-webserver"
	_ "github.com/kevinburke/differ"
	_ "github.com/kevinburke/go-bindata/go-bindata"
	_ "github.com/mattn/goreman"
	_ "github.com/nxpkg/go-jsonschema/cmd/go-jsonschema-compiler"
	_ "github.com/nxpkg/godockerize"
	_ "golang.org/x/tools/cmd/stringer"
	_ "honnef.co/go/tools/cmd/megacheck"
	_ "honnef.co/go/tools/cmd/staticcheck"
	_ "honnef.co/go/tools/cmd/unused"
)
