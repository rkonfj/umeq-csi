package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tasselsd/umeq-csi/internel/state"
	"github.com/tasselsd/umeq-csi/internel/umeq"
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
				fmt.Printf("%s:%s\n", val.Key, string(val.Value))
			}
		}}

	var rootCmd = &cobra.Command{Use: "agentctl"}
	rootCmd.AddCommand(createCmd, deleteCmd)
	rootCmd.AddCommand(publishCmd, unpublishCmd)
	rootCmd.AddCommand(kvCmd)
	rootCmd.Execute()
}
