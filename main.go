package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/d2jvkpn/xrun/cmd"

	"github.com/spf13/cobra"
)

//go:generate bash go_build.sh

var (
	//go:embed .version
	_Version   string // 0.1.0
	_BuildTime string
)

func init() {
	_Version = strings.Fields(_Version)[0]
}

func main() {
	link := "https://github.com/d2jvkpn/xrun"
	root := &cobra.Command{Use: fmt.Sprintf("xrun (%s)", link)}

	root.AddCommand(cmd.NewVersionCmd("version", _Version, _BuildTime, link))
	root.AddCommand(cmd.NewRunCmd("run"))

	root.Execute()
}
