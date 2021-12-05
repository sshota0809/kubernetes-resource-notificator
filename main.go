package main

import (
	"github.com/sshota0809/kubernetes-resource-notificator/cmd"
)

func main() {
	cmd := cmd.NewCommand()
	cmd.Execute()
}
