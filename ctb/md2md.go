package main

import (
	"git.catbo.net/muravjov/go2023/markdown"
	"git.catbo.net/muravjov/go2023/util"
)

func md2md(md2html bool, dumpAST bool, args []string) bool {
	if len(args) != 2 {
		util.Errorf("html2markdown: strictly 2 arguments required")
		return false
	}

	srcFilename, dstFilename := args[0], args[1]

	dat, res := openSrc(srcFilename)
	if !res {
		return res
	}

	dstF, res := openDst(dstFilename)
	if !res {
		return res
	}
	defer dstF.Close()

	markdown.Convert(dat, dstF, !md2html, dumpAST, false)

	return true
}
