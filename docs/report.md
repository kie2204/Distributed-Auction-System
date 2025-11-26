# Mandatory Activity 5 - Distributed Auction System

## Introduction
We have developed a distributed auction system with leader-based replication. We have implemented service discovery for the servers, so that they can find each other without manual user configuration. 

The system can handle n-1 crashes with n servers, meaning you can crash all but one server and continue operation (assuming the client also knows ports of the remaining servers). 

As described in the mandatory activity, our system has implemented the two required API methods bid (to send bids) and result (to get current auction state).

## Architecture
### Service discovery
We use a simple service discovery algorithm. The first node will always connect at the same port (33345). If the port is not available, we will know that it is an active node, so we can connect to it and exchange ports. Then, every node will be kept up-to-date on which nodes exist, since any new nodes will be shared with all existing nodes.

### Replication
The servers use leader-based replication to ensure n-1 fault tolerance. The leader is elected through the bully algorithm where their ID is their port number. Once a leader (coordinator) is elected its followers recieve changes from the leader. 

To avoid a crashed leader not being discovered the followers ping the leader every five seconds with a five second timeout. The timeout is five seconds, and not something less like 500 ms, to avoid unnecessary failovers due to slow connection.

To ensure zero data loss we use synchronous replication. This is because we are handling "money" and care more for data loss more than latancy. 
The leader update it's followers through an rpc call and the followers updates thier variables and send back an acknowledgement.
A follower crashing doesn't really affect the system. Although synchronous replication is used the leader immediately gets a connection error when trying to update a crashed follower. This avoids unnecessary latency of waiting for a timeout because the follower won't send back an acknowledgement.

A failover (leader chrashes and new leader gets elected) gets started when either a followers ping gets timedout or a client tries to get data but can't due to crash of leader. The system calls an election and use bully algorithm to choose new leader.

## Correctness 1
Our implementation is 100% linearizable!

## Correctness 2
Our protocol works as expected during normal operation (in the absence of failures). 
- The client can connect to any node and discover the leader if necessary. 
- Bids can be sent to the leader without issue, and the server replicates the bids to any working followers before sending a success reply. 
- The state can be queried on any of the nodes, whether they are leaders or followers.

Our protocol can also work in the presence of failures. 

- If a follower fails, the system will continue as usual; the leader will simply ignore the failed follower. 
- If a leader fails, the followers will eventually discover the failure and start an election. Then a follower will be promoted as the new leader. 
- If the client can no longer reach the server, and it knows multiple ports, it will reconnect to a new port and discover the new leader. From there, bidding can continue as usual. 
