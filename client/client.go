package main

import (
	proto "Auction/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ports []int32
var coordinator_port int32

func main() {
	for _, v := range os.Args[1:] {
		port, _ := strconv.Atoi(v)
		ports = append(ports, int32(port))
	}

	for k, v := range ports {
		log.Printf("%d: %d", k, v)
	}

	client := get_client()

	for {
		reader := bufio.NewReader(os.Stdin)
		lineB, _, _ := reader.ReadLine()
		line := string(lineB)
		words := strings.Split(line, " ")
		if words[0] == "bid" {
			bid, _ := strconv.ParseInt(words[1], 10, 32)
			a, _ := client.SendBid(context.Background(), &proto.Bid{Amount: int32(bid)})
			switch a.State {
			case proto.State_FAIL:
				fmt.Println("Bid failed")
			case proto.State_SUCCESS:
				fmt.Printf("Your bid of %d was successful", bid)
			_:
				fmt.Println("An exception has occured")
				client = get_client()
			}
		}
	}
	// client := proto.NewAuctionNodeServiceClient(conn)

	// select {}
}

func get_client() proto.AuctionNodeServiceClient {
	var conn *grpc.ClientConn
	var err error
	var client proto.AuctionNodeServiceClient

	for _, port := range ports {
		conn, err = grpc.NewClient(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			continue
		}

		// Is it a leader?
		client = proto.NewAuctionNodeServiceClient(conn)
		coordinator, _ := client.GetCoordinator(context.Background(), &proto.Empty{})
		if coordinator.Port != port {
			log.Printf("Port %d is not a coordinator, trying coordinator port %d", port, coordinator.Port)
			conn, err = grpc.NewClient(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			client = proto.NewAuctionNodeServiceClient(conn)
		} else {
			break
		}
	}

	if err != nil {
		log.Fatalf("No connection!")
	}

	return client
}
