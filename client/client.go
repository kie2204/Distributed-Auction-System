package main

import (
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ports []int
var coordinator_port int

func main() {
	for _, v := range os.Args[1:] {
		port, _ := strconv.Atoi(v)
		ports = append(ports, port)
	}

	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))

	for k, v := range ports {
		log.Printf("%d: %d", k, v)
	}
	// client := proto.NewAuctionNodeServiceClient(conn)

	// select {}
}
