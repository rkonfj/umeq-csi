// Copyright 2022 rkonfj@fnla.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tasselsd/umeq-csi/pkg/state"
	"github.com/tasselsd/umeq-csi/pkg/umeq"
)

func main() {
	agentServer, ok := os.LookupEnv("AGENT_SERVER")
	if !ok {
		agentServer = "http://127.0.0.1:8080"
	}
	agent := umeq.NewAgentService(agentServer)

	var createCmd = &cobra.Command{
		Use:   "create <volumeId> <requiredBytes>",
		Short: "create a valume",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			bytes, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				panic(err)
			}
			err = agent.CreateVolume("default", args[0], bytes)
			if err != nil {
				panic(err)
			}
		}}

	var deleteCmd = &cobra.Command{
		Use:   "delete <volumeId>",
		Short: "delete a valume",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := agent.DeleteVolume(args[0])
			if err != nil {
				panic(err)
			}
		}}

	var publishCmd = &cobra.Command{
		Use:   "publish <volumeId> <nodeId>",
		Short: "publish a volume to node",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := agent.PublishVolume(args[0], args[1])
			if err != nil {
				panic(err)
			}
			devpath, _ := agent.GetDevPath(args[0])
			fmt.Println(args[1], devpath)
		}}
	var unpublishCmd = &cobra.Command{
		Use:   "unpublish <volumeId> <nodeId>",
		Short: "unpublish a volume from node",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := agent.UnpublishVolume(args[0], args[1])
			if err != nil {
				panic(err)
			}
		}}

	var kvCmd = &cobra.Command{
		Use:   "fskv <fsPath>",
		Short: "list filesystem kvStore content",
		Long:  "avaiable store is agent and attacher",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			kv := state.NewFsKvStore(args[0])
			kvs, err := kv.List()
			if err != nil {
				panic(err)
			}
			for _, val := range kvs {
				fmt.Printf("%s\n%s:%s\n", val.CodedKey, val.Key, string(val.Value))
				fmt.Println()
			}
			fmt.Println("total count:", len(kvs))
		}}

	var rootCmd = &cobra.Command{Use: "agentctl"}
	rootCmd.AddCommand(createCmd, deleteCmd)
	rootCmd.AddCommand(publishCmd, unpublishCmd)
	rootCmd.AddCommand(kvCmd)
	rootCmd.Execute()
}
