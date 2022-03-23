package cmd

import (
	// "fmt"
	"log"

	"github.com/d2jvkpn/xrun/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRunCmd(name string) (command *cobra.Command) {
	var (
		parallel       int
		objects        []string
		pipeline, task string
		fSet           *pflag.FlagSet
	)

	command = &cobra.Command{
		Use:   name,
		Short: `run a task`,
		Long:  `run a task in pipeline`,

		Run: func(cmd *cobra.Command, args []string) {
			var (
				err error
				pl  *internal.Pipeline
			)

			if pl, err = internal.LoadPipeline(pipeline); err != nil {
				log.Fatalln(err)
			}

			if err = pl.RunTask(task, parallel, objects...); err != nil {
				log.Fatalln(err)
			}
		},
	}

	objects = make([]string, 0)
	fSet = command.Flags()

	fSet.StringVarP(&pipeline, "pipeline", "y", "pipeline.yaml", "pipeline yaml")
	fSet.StringVarP(&task, "task", "t", "", "task name")
	fSet.IntVarP(&parallel, "parallel", "p", -1, "parallel number, 0 for no limit")
	fSet.StringArrayVarP(&objects, "object", "o", []string{}, "select objects")

	cobra.MarkFlagRequired(fSet, "task")

	return command
}
