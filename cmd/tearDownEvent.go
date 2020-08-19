/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"time"

	"github.com/apeunit/evtvzd/pkg/evtvzd"
	"github.com/spf13/cobra"
)

// tearDownEventCmd represents the tearDownEvent command
var tearDownEventCmd = &cobra.Command{
	Use:   "teardown-event",
	Short: "Destroy the resources associated to an event",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run:   tearDownEvent,
}

func init() {
	rootCmd.AddCommand(tearDownEventCmd)

}

func tearDownEvent(cmd *cobra.Command, args []string) {
	fmt.Println("Teardown Event")
	fmt.Println("Event ID is", args[0])
	start := time.Now()
	err := evtvzd.DestroyEvent(settings, args[0])
	if err != nil {
		fmt.Println("There was an error shutting down the event: ", err)
	}
	fmt.Println("Operation completed in", time.Since(start))
}
