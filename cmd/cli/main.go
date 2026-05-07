package main

import "redis/internal/server"

func main() {
	server.RunAsyncTCPServer()
}
