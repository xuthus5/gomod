package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitter.top/common/lormatter"

	"gitter.top/apps/gomod"
)

var (
	upgradeIndirect bool
	mod             *cobra.Command
)

func init() {
	formatter := &lormatter.Formatter{ShowField: true}
	logrus.SetFormatter(formatter)
	logrus.SetReportCaller(true)

	mod = &cobra.Command{
		Use:     "gomod",
		Short:   "go mod manager",
		Example: "gomod",
		PreRun: func(cmd *cobra.Command, args []string) {
			_, err := os.Stat("go.mod")
			if os.IsNotExist(err) {
				logrus.Fatalf("go.mod file not found on this directory")
			} else if err != nil {
				logrus.Fatalf("check go.mod file failed: %v", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	mod.AddCommand(upgrade())
	mod.AddCommand(analyzed())
}

func upgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "update project dependencies to latest",
		Aliases: []string{"u"},
		Run: func(cmd *cobra.Command, args []string) {
			gomod.ModUpgrade(upgradeIndirect)
		},
	}
	cmd.Flags().BoolVarP(&upgradeIndirect, "indirect", "i", false, "upgrade indirect dependency")
	return cmd
}

func analyzed() *cobra.Command {
	return &cobra.Command{
		Use:     "analyzed",
		Short:   "analyzed project dependencies",
		Aliases: []string{"a"},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}

func main() {
	if err := mod.Execute(); err != nil {
		logrus.Errorf("execute command failed: %v", err)
	}
}
