package main

import (
	"os"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"

	"github.com/crazygit/cert-manager-alidns-webhook/pkg/alidns"
)

var GroupName = os.Getenv("GROUP_NAME")

const defaultGroupName = "alidns.crazygit.github.io"

func main() {
	if GroupName == "" {
		// 默认使用开发环境的 groupName
		GroupName = defaultGroupName
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName, &alidns.Solver{})
}
