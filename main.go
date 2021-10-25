package main

import (
	"flag"

	"k8s.io/klog/v2"
	"storageclass-accessor/webhook"
)

func main() {
	rootCmd := webhook.CmdWebhook

	loggingFlags := &flag.FlagSet{}
	klog.InitFlags(loggingFlags)
	rootCmd.PersistentFlags().AddGoFlagSet(loggingFlags)
	rootCmd.Execute()
}
