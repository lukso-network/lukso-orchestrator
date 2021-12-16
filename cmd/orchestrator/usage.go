// This code was adapted from https://github.com/ethereum/go-ethereum/blob/master/cmd/geth/usage.go
package main

import (
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"io"
	"sort"

	"github.com/urfave/cli/v2"
)

var appHelpTemplate = `NAME:
   {{.App.Name}} - {{.App.Usage}}
USAGE:
   {{.App.HelpName}} [options]{{if .App.Commands}} command [command options]{{end}} {{if .App.ArgsUsage}}{{.App.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if .App.Version}}
AUTHOR:
   {{range .App.Authors}}{{ . }}{{end}}
   {{end}}{{if .App.Commands}}
GLOBAL OPTIONS:
   {{range .App.Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}{{if .FlagGroups}}
{{range .FlagGroups}}{{.Name}} OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
{{end}}{{end}}{{if .App.Copyright }}
COPYRIGHT:
   {{.App.Copyright}}
VERSION:
   {{.App.Version}}
   {{end}}{{if len .App.Authors}}
   {{end}}
`

type flagGroup struct {
	Name  string
	Flags []cli.Flag
}

var appHelpFlagGroups = []flagGroup{
	{
		Name: "cmd",
		Flags: []cli.Flag{
			cmd.DataDirFlag,
			cmd.VerbosityFlag,
			cmd.ForceClearDB,
			cmd.ClearDB,
			cmd.BoltMMapInitialSizeFlag,
		},
	},
	{
		Name: "orchestrator",
		Flags: []cli.Flag{
			cmd.IPCPathFlag,
			cmd.HTTPEnabledFlag,
			cmd.HTTPListenAddrFlag,
			cmd.HTTPPortFlag,
			cmd.HTTPVirtualHosts,
			cmd.WSEnabledFlag,
			cmd.WSListenAddrFlag,
			cmd.WSPortFlag,
			cmd.VanguardGRPCEndpoint,
			cmd.PandoraRPCEndpoint,
			cmd.VanguardGenesisTime,
			cmd.SecondsPerSlot,
		},
	},
	{
		Name: "log",
		Flags: []cli.Flag{
			cmd.LogFormat,
			cmd.LogFileName,
		},
	},
}

func init() {
	cli.AppHelpTemplate = appHelpTemplate

	type helpData struct {
		App        interface{}
		FlagGroups []flagGroup
	}

	originalHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, tmpl string, data interface{}) {
		if tmpl == appHelpTemplate {
			for _, group := range appHelpFlagGroups {
				sort.Sort(cli.FlagsByName(group.Flags))
			}
			originalHelpPrinter(w, tmpl, helpData{data, appHelpFlagGroups})
		} else {
			originalHelpPrinter(w, tmpl, data)
		}
	}
}
