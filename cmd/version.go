package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	//go:embed .project
	_Project string
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

func (v *Version) Json() string {
	bts, _ := json.Marshal(v)
	return string(bts)
}

func NewVersionCmd(name, version, buildTime string) (command *cobra.Command) {
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
				Project:   strings.Fields(_Project)[0],
				Version:   version,
				BuildTime: buildTime,
				GoVersion: strings.Replace(runtime.Version(), "go", "", 1),
			}

			if jsonFmt {
				fmt.Println(v.Json())
			} else {
				fmt.Printf("%s\n", v)
			}
		},
	}

	fSet = command.Flags()
	fSet.BoolVarP(&jsonFmt, "json", "j", false, "output command in json object")

	return command
}
