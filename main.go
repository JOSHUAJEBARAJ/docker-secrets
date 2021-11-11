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
package main

import (
	"fmt"
	"os/exec"

	"github.com/JOSHUAJEBARAJ/docker-secrets/cmd"
)

func main() {

	// check whether detect secrets is installed

	_, err := exec.Command("detect-secrets").Output()
	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			fmt.Println("Detect secrets not found please install it using the below command \n pip3 install detect-secrets==1.0.3 ")
		}

		return
	}

	cmd.Execute()
}
