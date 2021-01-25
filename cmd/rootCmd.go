package cmd

import (
	"fmt"

	"github.tools.sap/D001323/landep/pkg/installer"

	"github.com/Masterminds/semver/v3"

	"github.com/spf13/cobra"
	"github.tools.sap/D001323/landep/pkg/landep"
)

var (
	// Used for flags.
	pkg       string
	version   string
	namespace string

	rootCmd = &cobra.Command{
		Use:   "installer",
		Short: "Installer",
		Long:  `Installer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			installer.Init()
			landep.InitFakeTargetFactory(func(message string) {
				fmt.Println(message)
			})

			constraints, err := semver.NewConstraint(version)
			if err != nil {
				return err
			}

			pkgManager := landep.NewPackageManager(landep.Repository)
			k8sConfig := &landep.K8sConfig{URL: "https://gardener.canary.hana-ondemand.com"}
			target := landep.NewK8sTarget(namespace, k8sConfig)

			_, err = pkgManager.Apply(target, pkg, constraints, nil)
			return err
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {

	rootCmd.PersistentFlags().StringVar(&pkg, "pkg", "", "package")
	rootCmd.PersistentFlags().StringVar(&version, "version", ">=0.0", "version")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "namespace")
}
