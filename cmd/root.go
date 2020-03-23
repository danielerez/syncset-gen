package cmd

	import (
		"encoding/json"
		"errors"
		"fmt"
		"io/ioutil"
		"log"
		"os"

		"github.com/matt-simons/ss/pkg"
		hivev1 "github.com/openshift/hive/pkg/apis/hive/v1"
		"github.com/spf13/cobra"
		"sigs.k8s.io/yaml"
	)

	func init() {
		viewCmd.Flags().StringVarP(&selector, "selector", "s", "", "The selector key/value pair used to match the SelectorSyncSet to Cluster(s)")
		viewCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The cluster name used to match the SyncSet to a Cluster")
		viewCmd.Flags().StringVarP(&resources, "resources", "r", "", "The directory of resource manifest files to use")
		viewCmd.Flags().StringVarP(&patches, "patches", "p", "", "The directory of patch manifest files to use")
		viewCmd.Flags().BoolVarP(&flatten, "flatten", "f", false, "Output a single SelectorSyncSet")
		RootCmd.AddCommand(viewCmd)
	}

	var selector, clusterName, resources, patches, name string
	var flatten bool
	var input []byte

	var RootCmd = &cobra.Command{
		Use:   "ss",
		Short: "SyncSet/SelectorSyncSet generator.",
		Long:  ``,
	}

	func isInputFromPipe() bool {
	    fileInfo, _ := os.Stdin.Stat()
	    return fileInfo.Mode() & os.ModeCharDevice == 0
	}

	var viewCmd = &cobra.Command{
		Use:   "view",
		Short: "Parses a manifest directory and prints a SyncSet/SelectorSyncSet representation of the objects it contains.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if selector == "" && clusterName == "" {
				return errors.New("one of --selector or --cluster-name must be specified")
			}
			if selector != "" && clusterName != "" {
				return errors.New("only one of --selector or --cluster-name can be specified")
			}
			if len(args) < 1 {
				return errors.New("name must be specified")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if clusterName != "" {
				secrets := pkg.TransformSecrets(args[0], "ss", resources)
				for _, s := range secrets {
					j, err := json.MarshalIndent(&s, "", "    ")
					if err != nil {
						log.Fatalf("error: %v", err)
					}
					fmt.Printf("%s\n", string(j))
				}
				var ss hivev1.SyncSet
				ss = pkg.CreateSyncSet(args[0], clusterName, resources, patches)
				j, err := json.MarshalIndent(&ss, "", "    ")
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				fmt.Printf("%s\n\n", string(j))
			} else {
				if isInputFromPipe() {
					input , _ = ioutil.ReadAll(os.Stdin)
				}
//				secrets := pkg.TransformSecrets(args[0], "sss", resources)
//				for _, s := range secrets {
//					j, err := json.MarshalIndent(&s, "", "    ")
//					if err != nil {
//						log.Fatalf("error: %v", err)
//					}
//					fmt.Printf("%s\n", string(j))
//				}
				ss1, ss2 := pkg.CreateSelectorSyncSet(args[0], selector, input, resources, patches, flatten)
			j1, err := yaml.Marshal(&ss1)
			j2, err := yaml.Marshal(&ss2)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			if flatten {
				fmt.Printf("%s\n\n", string(j1))
			} else {
				fmt.Printf("%s---\n%s\n\n", string(j1), string(j2))
			}
		}
	},
}
