package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/buraksaglam089/go-healthcheck/monitor"
)

func main() {
	fmt.Println("Gopher watch is starting...")

	configPath := flag.String("config", "targets.json", "Path to the configuration file")
	flag.Parse()

	file, err := os.Open(*configPath)
	if err != nil {
		fmt.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	targets := []monitor.Target{}
	err = json.NewDecoder(file).Decode(&targets)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	m := monitor.NewMonitor(targets)
	m.Run(ctx)

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	fmt.Println("\nGracefully shutting down...")
}
