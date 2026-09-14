package main

import (
	"flag"
	"fmt"

	"github.com/IsaqueB/notification-server/cmd/api"
	"github.com/IsaqueB/notification-server/cmd/worker"
)

func main() {
	runFlag := flag.String("r", "worker", "select service to run [api]|[worker] default:[worker]")
	flag.Parse()

	switch *runFlag {
	case "api":
		api.Run()
	case "worker":
		worker.Run()
	default:
		fmt.Println("Invalid service")
	}
}
