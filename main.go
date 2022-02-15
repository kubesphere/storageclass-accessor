package main

import (
	"flag"

	"github.com/kubesphere/storageclass-accessor/webhook"
	"k8s.io/klog/v2"
)

func main() {
	rootCmd := webhook.CmdWebhook

	loggingFlags := &flag.FlagSet{}
	klog.InitFlags(loggingFlags)
	rootCmd.PersistentFlags().AddGoFlagSet(loggingFlags)
	rootCmd.Execute()
}
