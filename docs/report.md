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

To avoid a crashed leader not being discovered the followers ping the leader every five seconds with a five second timeout. The timeout is five seconds to avoid unnecessary failovers.

To ensure zero data loss we use synchronous replication. This is because we are handling money and care more for data loss than latancy. 
rashing doesn't really affect the system. Although synchronous repclient tries to get some data.
le case on refused error immediately ,when the follower is crashed.
In tries to cas some  of a failover (leader crashes) the pings from the followers or when a client gets 

## Correctness 1
Our implementation is 100% linearizable!

## Correctness 2
Our protocol works correctly in the absence of failures (during normal operation). The client can connect to the leader

