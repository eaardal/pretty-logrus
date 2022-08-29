module github.com/eaardal/pretty-logrus

go 1.19

// replace github.com/y/original-project => /path/to/x/my-version
// replace pretty-logrus => ./
// replace pretty-logrus => github.com/eaardal/pretty-logrus@latest

require (
	github.com/fatih/color v1.13.0
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
