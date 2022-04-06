package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	//go:embed demo.yaml
	demo string
)

func NewDemoCmd(name string) (command *cobra.Command) {
	command = &cobra.Command{
		Use:   name,
		Short: "show demo",
		Long:  `show demo of pipeline and commandline`,

		Run: func(cmd *cobra.Command, args []string) {
			tmp := "./xrun run -y demo.yaml -t sleep -p 0 -o a1 -o a3"
			fmt.Printf(
				"cat > demo.yaml <<EOF\n%s\nEOF\n\n%s\n",
				demo, tmp,
			)
		},
	}

	return command
}
