package main

import (
	"os"

	"git.catbo.net/muravjov/go2023/util"
	"github.com/spf13/cobra"
)

func runCLI() (exitOK bool) {
	var rootCmd *cobra.Command
	rootCmd = &cobra.Command{
		Use:   "ctb",
		Short: "ctb provides a workflow to translate articles, documentations, books etc, from catbo.net",
		Run: func(cmd *cobra.Command, args []string) {
			// print all commands' help when no subcommand
			// nolint: errcheck
			rootCmd.Help()
		},
	}

	// * html2markdown
	var llmProvider string
	var logRequests bool
	html2markdownCmd := &cobra.Command{
		Use:   "html2markdown srcfile|- dstfile|-",
		Short: "translate html to markdown",
		Run: func(cmd *cobra.Command, args []string) {
			exitOK = html2markdown(llmProvider, logRequests, args)
		},
	}
	html2markdownCmd.Flags().StringVar(&llmProvider, "llm", "", "openai | gigachat")
	html2markdownCmd.MarkFlagRequired("llm")
	html2markdownCmd.Flags().BoolVar(&logRequests, "log-requests", false, "log requests to llm provider")

	rootCmd.AddCommand(html2markdownCmd)

	// * md2md
	var md2html bool
	var dumpAST bool

	md2mdCmd := &cobra.Command{
		Use:   "md2md srcfile|- dstfile|-",
		Short: "convert markdown to markdown",
		Run: func(cmd *cobra.Command, args []string) {
			exitOK = md2md(md2html, dumpAST, args)
		},
	}
	md2mdCmd.Flags().BoolVar(&md2html, "md2html", false, "markdown to html instead")
	md2mdCmd.Flags().BoolVar(&dumpAST, "dumpAST", false, "dump dumps an AST tree structure to stdout")

	rootCmd.AddCommand(md2mdCmd)

	if err := rootCmd.Execute(); err != nil {
		util.Errorf("CLI error: %s", err)
		exitOK = false
		return
	}

	return
}

func main() {
	code := 0
	if !runCLI() {
		code = 1
	}
	os.Exit(code)
}
