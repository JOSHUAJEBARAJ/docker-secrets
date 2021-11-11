/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/JOSHUAJEBARAJ/docker-secrets/pkg/client"
	"github.com/JOSHUAJEBARAJ/docker-secrets/pkg/utils"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "pass the image name which you want to scan  along with tag\n eg: docker scan alpine:latest",

	Run: func(cmd *cobra.Command, args []string) {
		image_name := strings.Join(args, " ")
		//fmt.Println("scan called", image_name)

		// get all images
		cli, err := client.Init()

		if err != nil {
			fmt.Println("Error Intialize the client", err)
			os.Exit(1)
		}
		images, err := client.GetImages(cli)
		if err != nil {
			fmt.Println("Error getting images", err)
			os.Exit(1)
		}
		var id string
		// search for image present in local system
		for _, v := range images {

			if v.Name == image_name {
				id = v.Id

			}
		}

		// checking whether the image is present in the local system
		if id == "" {
			fmt.Printf("Image: %s  not present in local machine 😟\n", image_name)
			return
		}
		// saving images
		save_error := client.Save(id)
		if save_error != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// untaring images
		output_error := utils.Untar("output.tar", "output")
		if output_error != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// untaring the subfolder
		untar_err := utils.Outputar()
		if untar_err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// scanning
		scan_err := utils.Scan()

		if scan_err != nil {
			fmt.Println(err)
			os.Exit(1)

		}

	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
