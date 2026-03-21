package main

import (
	"certsync/src/core/client"
	"certsync/src/core/config"
	"certsync/src/core/meta"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configFileClient     string
	generateConfigClient bool
	printUsageClient     bool
)

func init() {
	flag.StringVar(&configFileClient, "c", "client.conf", "config file path, default client.conf")
	flag.BoolVar(&generateConfigClient, "g", false, "generate config, default client.conf")
	flag.BoolVar(&printUsageClient, "h", false, "print usage")

	flag.Usage = clientUsage
}

func main() {
	flag.Parse()

	if printUsageClient {
		clientUsage()
		os.Exit(0)
	}

	if generateConfigClient {
		metaData := meta.ClientConfig{
			Server:        "https://your-server.com",
			Token:         "your-server-token",
			CertAlias:     "example.com",
			CertAuth:      "your-cert-auth",
			DomainCheck:   "example.com:443",
			LocalCheck:    false,
			CertUpdateDir: "/etc/nginx/ssl/example.com",
			CertUpdateCmd: "nginx -s reload",
			Interval:      86400,
		}
		err := config.WriteConfig(configFileClient, &metaData)
		if err != nil {
			slog.Error("failed to write config", "error", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	var clientConfig meta.ClientConfig
	if err := config.ReadConfig(configFileClient, &clientConfig); err != nil {
		slog.Error("failed to read config", "error", err)
		os.Exit(1)
	}

	if clientConfig.Empty(nil) {
		slog.Error("config file error: empty")
		os.Exit(1)
	}

	clientConfig.Generate()
	client.MetaData = &clientConfig

	interval := time.Duration(clientConfig.Interval) * time.Second
	slog.Info("starting client", "interval", interval)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("running initial sync")
	if err := client.Sync(); err != nil {
		slog.Error("initial sync failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			slog.Info("starting scheduled sync")
			if err := client.Sync(); err != nil {
				slog.Error("sync failed", "error", err)
			}
		case <-stopChan:
			slog.Info("received shutdown signal, exiting")
			return
		}
	}
}

func clientUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  client startup: ")
	fmt.Fprintln(os.Stderr, "    certsync-client [-c config file]")
	fmt.Fprintln(os.Stderr, "  client startup in background: ")
	fmt.Fprintln(os.Stderr, "    nohup certsync-client [-c config file] &")
	fmt.Fprintln(os.Stderr, "  generate demo config file: ")
	fmt.Fprintln(os.Stderr, "    certsync-client -g [-c config file]")
	fmt.Fprintln(os.Stderr, "  print usage: ")
	fmt.Fprintln(os.Stderr, "    certsync-client -h")
	fmt.Fprintln(os.Stderr, "Options:")
	flag.PrintDefaults()
}
