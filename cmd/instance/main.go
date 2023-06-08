package main

import (
	"github.com/spf13/pflag"
	"github.com/tachunwu/gravity-router/pkg/server"
)

var routerURL string

func main() {

	// Parse flags
	pflag.StringVarP(&routerURL, "router-url", "", "localhost:4222", "The URL of the router")
	pflag.Parse()

	s := server.NewServer(routerURL)
	s.Start()

	select {}
}
