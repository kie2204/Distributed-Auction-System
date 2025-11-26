# Distributed Auction System

## How to run the system
### Getting started
1. First of all, make sure that you have golang installed. You can download golang [here](https://go.dev/dl/).
2. Navigate to this package though your terminal.

### Running the servers
1. To run a server, run this command in your terminal: `go run server/server.go`.
2. Repeat step 1. `n - 1` times, where `n` is the amount of server nodes you want.
3. Lastly, run the same command again but with the argument `ready` at the end: `go run server/server.go ready`. This creates the last server and tells the other nodes.

### Running the clients
1. To run a client, run this command in your terminal: `go run client/client.go <name> <node_port> (<node_port>)...`, where `<name>` is an optional name for the client and `<node_port>` is the port of one of the server nodes. You can add as many node port arguments as you would like, as long as they exist in the system (port 33345 is a server node per default) (we recommend at least 2). The client can ask these nodes for the current coordinator's port.
2. Repeat step 1. however many times you want.

## Client options
### How to bid
In the terminal of a client, run this command: `bid <amount>`, where `<amount>` is the amount of $$$ you want to bid. Bidding an amount that is lower or equal to the current top bid is prohibited and therefore not possible.

### How to see the current state of the auction
In the terminal of a client, run this command: `state`. This will show the name of the top bidder and the amount of $$$ this person bidded. If the auction is over, it should say something like `Bidder <name> won with a bid of $<amount>`, where `<name>` is the name of the person who won, and `<amount>` is the amount the person bidded.a

## Ruining the auction
You can try to ruin the auction by crashing the leader server node. This can be done by pressing `ctrl+c` inside the terminal of the leader node. This will stop the server from working. The system wont know immediatly that the leader node is down, only when a client tries to bid/get the current state, or when one of the follower nodes tries to ping the dead leader but without getting a pong. If it is because of a client querying a command, it will get prompted with an error saying that the server is down and that the client should try again later. While this is happening, the server nodes are holding an election, electing the new leader. When a new leader is found (usually takes a few seconds) and if a client tries to bid/get state, the client will instead try to contact one of the servers it knows to get the new leader node. The auction can then continue.

