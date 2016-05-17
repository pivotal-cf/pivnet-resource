package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/version"
)

var (
	// buildVersion is deliberately left uninitialized so it can be set at compile-time
	buildVersion string
)

func main() {
	if buildVersion == "" {
		version.Version = "dev"
	} else {
		version.Version = buildVersion
	}

	parser := flags.NewParser(&commands.Pivnet, flags.HelpFlag)

	_, err := parser.Parse()
	if err != nil {
		if err == commands.ErrShowHelpMessage {
			helpParser := flags.NewParser(&commands.Pivnet, flags.HelpFlag)
			helpParser.NamespaceDelimiter = "-"
			helpParser.ParseArgs([]string{"-h"})
			helpParser.WriteHelp(os.Stderr)
			os.Exit(0)
		}
		log.Fatal(err)
	}
}
