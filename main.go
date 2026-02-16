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
	"github.com/buraksaglam089/go-healthcheck/storage"
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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := storage.NewFileLogger("storage/target_json.json")
	m := monitor.NewMonitor(targets, logger)
	go m.Run(ctx)

	<-stop
	cancel()

	fmt.Println("\nGracefully shutting down...")
}
