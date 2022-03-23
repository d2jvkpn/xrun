package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Version struct {
	Project   string `json:"project"`
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
}

func (v Version) String() string {
	return fmt.Sprintf(
		"project: %s\nversion: %s\nbuildTime: %s\ngoVersion: %s",
		v.Project, v.Version, v.BuildTime, v.GoVersion,
	)
}

func NewVersionCmd(name, version, buildTime, link string) (command *cobra.Command) {
	var (
		jsonFmt bool
		fSet    *pflag.FlagSet
	)

	command = &cobra.Command{
		Use:   name,
		Short: "version",
		Long:  `program version`,

		Run: func(cmd *cobra.Command, args []string) {
			v := Version{
				Project:   link,
				Version:   version,
				BuildTime: buildTime,
				GoVersion: strings.Replace(runtime.Version(), "go", "", 1),
			}

			if jsonFmt {
				bts, _ := json.MarshalIndent(v, "", "  ")
				fmt.Printf("%s\n", bts)
			} else {
				fmt.Printf("%s\n", v)
			}
		},
	}

	fSet = command.Flags()
	fSet.BoolVarP(&jsonFmt, "json", "j", false, "output command in json object")

	return command
}
