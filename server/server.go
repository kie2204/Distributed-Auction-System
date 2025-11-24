package main

import (
	proto "Auction/grpc"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	RELEASED = iota
	WANTED
	HELD
)

type Server struct {
	proto.UnimplementedAuctionNodeServiceServer
}

const defaultPort int32 = 33345
const minPort int32 = 10000
const maxPort int32 = 40000

var port int32 = defaultPort
var nodes map[int32]proto.AuctionNodeServiceClient

func main() {
	nodes = make(map[int32]proto.AuctionNodeServiceClient)

	server := &Server{}
	listener := get_listener()
	go server.start_server(listener)

	if port != defaultPort {
		add_client(defaultPort, true)
	}

	// f, err := os.Create(fmt.Sprintf("events-%d.log", port))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// writer := bufio.NewWriter(f)
	// log.SetOutput(writer)

	is_ready := len(os.Args) == 2 && os.Args[1] == "ready"
	if is_ready {
		server.Ready(context.Background(), &proto.Empty{})
		// for _, v := range nodes {
		// 	v.Ready(context.Background(), &proto.Empty{})
		// }
	}

	select {}
}

func get_listener() net.Listener {
	var listener net.Listener
	for {
		var err error
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			prevPort := port
			port = rand.Int31n(maxPort-minPort) + minPort
			log.Printf("Port %d is unavailable, retrying with port %d", prevPort, port)
			continue
		}
		break
	}

	return listener
}

func (s *Server) start_server(listener net.Listener) {
	grpcServer := grpc.NewServer()

	proto.RegisterAuctionNodeServiceServer(grpcServer, s)
	err := grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}
}

// ======================= Elections =========================

var state = RELEASED
var requests chan chan bool = make(chan chan bool, 1000)
var leaderPort int32

func (s *Server) Election(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	go func() {
		var wg sync.WaitGroup
		var isLeader = true

		for p, n := range nodes {
			if p > port {
				wg.Add(1)
				go func() {
					ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
					_, err := n.Election(ctx, &proto.Empty{})
					if err == nil {
						isLeader = false
					}
					wg.Done()
				}()
			}
		}
		wg.Wait()

		if isLeader {
			log.Printf("I, port %d, am now coordinator", port)
			leaderPort = port
			for _, n := range nodes {
				go n.Coordinator(context.Background(), &proto.Request{Port: port})
			}
		}
	}()

	return &proto.Empty{}, nil
}

func (s *Server) Coordinator(ctx context.Context, in *proto.Request) (*proto.Empty, error) {
	leaderPort = in.Port
	log.Printf("Port %d is now coordinator", leaderPort)
	return &proto.Empty{}, nil
}

func (s *Server) GetCoordinator(ctx context.Context, in *proto.Empty) (*proto.Request, error) {
	return &proto.Request{
		Port: leaderPort,
	}, nil
}

// =================== Service discovery =====================

func (s *Server) Announce(ctx context.Context, in *proto.Node) (*proto.Empty, error) {
	log.Printf("Announcement from port %d", in.Port)
	add_client(in.Port, false)

	return &proto.Empty{}, nil
}

func (s *Server) GetNodes(ctx context.Context, in *proto.Empty) (*proto.Nodes, error) {
	keys := make([]int32, 0, len(nodes))
	for k := range nodes {
		keys = append(keys, k)
	}

	return &proto.Nodes{Port: keys}, nil
}

func (s *Server) Ready(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	log.Printf("Finding leader.")
	go s.Election(context.Background(), &proto.Empty{})
	return &proto.Empty{}, nil
}

func add_client(_port int32, announce bool) {
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", _port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewAuctionNodeServiceClient(conn)
	nodes[_port] = client
	log.Printf("Client at port %d added", _port)

	_nodes, err := client.GetNodes(context.Background(), &proto.Empty{})
	if err != nil {
		log.Printf("Error: %s", err)
	} else {
		for _, v := range _nodes.Port {
			if nodes[v] == nil && v != port {
				add_client(v, true)
			}
		}
	}

	if announce {
		log.Printf("Announcing to %d", _port)
		client.Announce(context.Background(), &proto.Node{Port: port})
	}
}

// ======================== Bidding ==========================
var timestamp int32 = 0
var highest_bid int32 = 0
var bidder int32

func (s *Server) SendBid(ctx context.Context, in *proto.Bid) (*proto.Ack, error) {
	var state proto.State

	if port != leaderPort {
		state = proto.State_EXCEPTION
		return &proto.Ack{State: state}, nil
	}

	if highest_bid >= in.Amount {
		state = proto.State_FAIL
		return &proto.Ack{State: state}, nil
	}

	highest_bid = in.Amount
	bidder = in.Port
	state = proto.State_SUCCESS

	//Distribute to followers
	var wg sync.WaitGroup
	for _, v := range nodes {
		wg.Add(1)
		go func() {
			out := &proto.Update{
				Timestamp:  timestamp,
				BidderPort: in.Port,
				Amount:     in.Amount,
			}
			v.SendUpdate(context.Background(), out)
			wg.Done()
		}()
	}
	wg.Wait()

	return &proto.Ack{State: state}, nil
}
func (s *Server) GetResult(ctx context.Context, in *proto.Empty) (*proto.Result, error) {
	var done bool //are we done?? delete this variable

	return &proto.Result{Done: done, WinnerPort: bidder, HighestBid: highest_bid}, nil
}

// ====================== Replication ========================

func (s *Server) SendUpdate(ctx context.Context, in *proto.Update) (*proto.Empty, error) {
	highest_bid = in.Amount
	bidder = in.BidderPort
	timestamp = in.Timestamp
	return &proto.Empty{}, nil
}
