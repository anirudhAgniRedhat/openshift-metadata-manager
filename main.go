package main

import (
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Important for cloud auth
)

func main() {
	cmd.Execute()
}
