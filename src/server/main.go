package main

import (
	"certsync/src/core/cert"
	"certsync/src/core/config"
	"certsync/src/core/meta"
	"certsync/src/core/server"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/savsgio/atreugo/v11"
)

var (
	configFileServer     string
	generateConfigServer bool
	printUsageServer     bool
)

func init() {
	flag.StringVar(&configFileServer, "c", "server.conf", "config file path, default server.conf")
	flag.BoolVar(&generateConfigServer, "g", false, "generate config, default server.conf")
	flag.BoolVar(&printUsageServer, "h", false, "print usage")

	flag.Usage = serverUsage
}

func main() {
	flag.Parse()

	if printUsageServer {
		serverUsage()
		os.Exit(0)
	}

	if generateConfigServer {
		metaData := meta.ServerConfig{
			Address:   "0.0.0.0",
			Port:      8080,
			Token:     "your-auth-token",
			Storage:   "your-cert-storage",
			CheckTime: "03:00:00",
			DNS:       "dns-server",
		}
		metaData.Certs = []*meta.CertConfig{
			{
				Alias:           "example.com",
				Auth:            "your-cert-auth",
				Email:           "your-cert-email",
				Domain:          "example.com",
				DomainCheck:     "https://example.com",
				Provider:        "TENCENT",
				AccessKey:       "your-access-key",
				AccessKeySecret: "your-access-key-secret",
				AutoRenew:       true,
				UploadToken:     "your-upload-token",
			},
		}
		err := config.WriteConfig(configFileServer, &metaData)
		if err != nil {
			slog.Error("failed to write config", "error", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	var serverConfig meta.ServerConfig
	if err := config.ReadConfig(configFileServer, &serverConfig); err != nil {
		slog.Error("failed to read config", "error", err)
		os.Exit(1)
	}

	if serverConfig.Empty(nil) {
		slog.Error("config file error: empty")
		os.Exit(1)
	}

	serverConfig.Generate()
	server.MetaData = &serverConfig

	certStorage := cert.NewCertStorage(serverConfig.Storage)
	server.CertStorage = certStorage

	scheduler := server.NewScheduler(&serverConfig, certStorage)
	server.GlobalScheduler = scheduler
	scheduler.Start()

	address := fmt.Sprintf("%s:%d", serverConfig.Address, serverConfig.Port)
	slog.Info("starting server", "address", address)

	atreugoConfig := atreugo.Config{
		Addr: address,
	}

	atreugoServer := atreugo.New(atreugoConfig)

	atreugoServer.GET("/", server.IndexHandler)
	atreugoServer.GET("/h5/upload", server.UploadPageHandler)
	atreugoServer.GET("/api/{alias}/check", server.CheckHandler)
	atreugoServer.GET("/api/{alias}/download", server.DownloadHandler)
	atreugoServer.GET("/api/{alias}/expire", server.ExpireHandler)
	atreugoServer.GET("/api/{alias}/upload_verify", server.VerifyUploadTokenHandler)
	atreugoServer.POST("/api/{alias}/upload", server.UploadHandler)

	if err := atreugoServer.ListenAndServe(); err != nil {
		scheduler.Stop()
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func serverUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  server startup: ")
	fmt.Fprintln(os.Stderr, "    certsync-server [-c config file]")
	fmt.Fprintln(os.Stderr, "  server startup in background: ")
	fmt.Fprintln(os.Stderr, "    nohup certsync-server [-c config file] &")
	fmt.Fprintln(os.Stderr, "  generate demo config file: ")
	fmt.Fprintln(os.Stderr, "    certsync-server -g [-c config file]")
	fmt.Fprintln(os.Stderr, "  print usage: ")
	fmt.Fprintln(os.Stderr, "    certsync-server -h")
	fmt.Fprintln(os.Stderr, "Options:")
	flag.PrintDefaults()
}
