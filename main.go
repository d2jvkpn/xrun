package main

import (
	_ "embed"
	"strings"

	"github.com/d2jvkpn/xrun/cmd"

	"github.com/spf13/cobra"
)

//go:generate bash go_build.sh

var (
	//go:embed .version
	Version   string // 0.1.0
	BuildTime string
)

func init() {
	Version = strings.Fields(Version)[0]
}

func main() {
	link := "https://github.com/d2jvkpn/xrun"
	root := &cobra.Command{Use: "xrun"}

	root.AddCommand(cmd.NewVersionCmd("version", Version, BuildTime, link))
	root.AddCommand(cmd.NewRunCmd("run"))
	root.AddCommand(cmd.NewDemoCmd("demo"))

	root.Execute()
}
