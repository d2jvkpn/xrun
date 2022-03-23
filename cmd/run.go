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
		pipeline, task string
		parallel       int
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
				p   uint
			)
			if pl, err = internal.LoadPipeline(pipeline); err != nil {
				log.Fatalln(err)
			}

			if parallel > -1 {
				p = uint(parallel)
				err = pl.RunTask(task, p)
			} else {
				err = pl.RunTask(task)
			}

			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	fSet = command.Flags()

	fSet.StringVarP(&pipeline, "pipeline", "y", "pipeline.yaml", "pipeline yaml")
	fSet.StringVarP(&task, "task", "t", "", "task name")
	fSet.IntVarP(&parallel, "parallel", "p", -1, "parallel number, 0 for no limit")

	cobra.MarkFlagRequired(fSet, "task")

	return command
}
