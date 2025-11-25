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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ports []int32
var coordinator_port int32

func main() {
	name := os.Args[1]
	for _, v := range os.Args[2:] {
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
			a, err := client.SendBid(context.Background(), &proto.Bid{Name: name, Amount: int32(bid)})

			var state proto.State
			if err != nil {
				state = proto.State_EXCEPTION
			} else {
				state = a.State
			}

			switch state {
			case proto.State_FAIL:
				fmt.Println("Bid failed")
			case proto.State_SUCCESS:
				fmt.Printf("Your bid of %d was successful\n", bid)
			case proto.State_EXCEPTION:
				fmt.Println("An exception has occured")
				client = get_client()
			}
		} else if words[0] == "state" {
			res, err := client.GetResult(context.Background(), &proto.Empty{})
			if err != nil {
				fmt.Printf("%s\n", err)
				client = get_client()
				continue
			}
			if res.Done {
				fmt.Printf("Bidder %s won with a bid of $%d!\n", res.Winner, res.HighestBid)
			} else {
				fmt.Printf("Bidder %s has the highest bid of $%d.\n", res.Winner, res.HighestBid)
			}
		}
	}
	// client := proto.NewAuctionNodeServiceClient(conn)

	// select {}
}

func get_client() proto.AuctionNodeServiceClient {
	var client proto.AuctionNodeServiceClient
	var client_port int32

	i := 0
	for i < len(ports) {
		port := ports[i]
		conn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))

		client_temp := proto.NewAuctionNodeServiceClient(conn)
		_, err := client_temp.Ping(context.Background(), &proto.Empty{})
		if err == nil {
			client = client_temp
			client_port = port
		} else {
			log.Printf("Port %d was not available", port)
			i++
			continue
		}

		// Is it a leader?
		coordinator, c_err := client.GetCoordinator(context.Background(), &proto.Empty{})
		if c_err != nil {
			fmt.Printf("%s\n", c_err)
			time.Sleep(time.Second * 2)
			continue
		}

		if coordinator.Port != port {
			log.Printf("Port %d is not a coordinator, trying coordinator port %d", port, coordinator.Port)
			conn, _ = grpc.NewClient(fmt.Sprintf("localhost:%d", coordinator.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))

			client_temp := proto.NewAuctionNodeServiceClient(conn)
			_, err := client_temp.Ping(context.Background(), &proto.Empty{})
			if err == nil {
				client = client_temp
				client_port = coordinator.Port
			} else {
				continue
			}
		}

		break
	}

	if client == nil {
		log.Fatalf("No connection!")
	}

	log.Printf("Connected to port %d", client_port)
	return client
}
