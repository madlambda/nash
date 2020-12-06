package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/parser"
)

func docUsage(out io.Writer) {
	fmt.Fprintf(out, "Usage: %s <package>.<fn name or wildcard>\n", filepath.Base(os.Args[0]))
}

func printDoc(stdout, _ io.Writer, docs []*ast.CommentNode, fn *ast.FnDeclNode) {
	fmt.Fprintf(stdout, "fn %s(%s)\n", fn.Name(), strings.Join(fn.Args(), ", "))

	for _, doc := range docs {
		fmt.Fprintf(stdout, "\t%s\n", doc.String()[2:])
	}

	fmt.Println()
}

func lookFn(stdout, stderr io.Writer, fname string, pack string, pattern *regexp.Regexp) {
	content, err := ioutil.ReadFile(fname)

	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", err.Error())
		return
	}

	parser := parser.NewParser(fname, string(content))

	tree, err := parser.Parse()

	if err != nil {
		return
	}

	nodelen := len(tree.Root.Nodes)

	for i, j := 0, 1; j < nodelen; i, j = i+1, j+1 {
		var comments []*ast.CommentNode

		node := tree.Root.Nodes[i]
		next := tree.Root.Nodes[j]

		if node.Type() == ast.NodeComment {
			comments = append(comments, node.(*ast.CommentNode))
			last := node

			// process comments
			for i = i + 1; i < nodelen-1; i++ {
				node = tree.Root.Nodes[i]

				if node.Type() == ast.NodeComment &&
					node.Line() == last.Line()+1 {
					comments = append(comments, node.(*ast.CommentNode))
					last = node
				} else {
					break
				}
			}

			i--
			j = i + 1
			next = tree.Root.Nodes[j]

			if next.Line() != last.Line()+1 {
				comments = []*ast.CommentNode{}
			}

			if next.Type() == ast.NodeFnDecl {
				fn := next.(*ast.FnDeclNode)

				if pattern.MatchString(fn.Name()) {
					printDoc(stdout, stderr, comments, next.(*ast.FnDeclNode))

				}
			}
		} else if node.Type() == ast.NodeFnDecl {
			fn := node.(*ast.FnDeclNode)

			if pattern.MatchString(fn.Name()) {
				// found func, but no docs :-(
				printDoc(stdout, stderr, []*ast.CommentNode{}, fn)
			}
		}

	}
}

func walk(stdout, stderr io.Writer, nashpath, pkg string, pattern *regexp.Regexp) error {
	return filepath.Walk(nashpath+"/lib", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		dirpath := filepath.Dir(path)
		dirname := filepath.Base(dirpath)
		ext := filepath.Ext(path)

		if ext != "" && ext != ".sh" {
			return nil
		}

		if dirname != pkg {
			return nil
		}

		lookFn(stdout, stderr, path, pkg, pattern)

		return nil
	})
}

func doc(stdout, stderr io.Writer, args []string) error {
	if len(args) < 1 {
		docUsage(stderr)
		return nil
	}

	packfn := args[0]
	parts := strings.Split(packfn, ".")

	if len(parts) != 2 {
		docUsage(stderr)
		return nil
	}

	pkg := parts[0]

	if strings.ContainsAny(parts[1], ".[]()") {
		return fmt.Errorf("Only wildcards * and ? supported")
	}

	patternStr := strings.Replace(parts[1], "*", ".*", -1)
	patternStr = strings.Replace(patternStr, "?", ".?", -1)
	patternStr = "^" + patternStr + "$"

	pattern, err := regexp.Compile(patternStr)

	if err != nil {
		return fmt.Errorf("invalid pattern: %s", err.Error())
	}

	nashpath := os.Getenv("NASHPATH")

	if nashpath == "" {
		homepath := os.Getenv("HOME")

		if homepath == "" {
			return fmt.Errorf("NASHPATH not set...\n")
		}

		fmt.Fprintf(stderr, "NASHPATH not set. Using ~/.nash\n")

		nashpath = homepath + "/.nash"
	}

	return walk(stdout, stderr, nashpath, pkg, pattern)
}
