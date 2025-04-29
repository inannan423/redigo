package cluster

import (
	"context"
	"errors"
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/client"
	"redigo/resp/reply"
	"strconv"
)

// getPeerClient retrieves a client for the specified peer node
func (c *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := c.peerConn[peer]
	if !ok {
		return nil, errors.New("peer not found")
	}
	conn, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	// Turn the borrowed object into a client
	client, ok := conn.(*client.Client)
	if !ok {
		return nil, errors.New("invalid connection type")
	}
	return client, nil
}

// returnPeerClient returns a client to the specified peer node
func (c *ClusterDatabase) returnPeerClient(peer string, client *client.Client) error {
	pool, ok := c.peerConn[peer]
	if !ok {
		return errors.New("peer not found")
	}
	// Return the client to the pool
	return pool.ReturnObject(context.Background(), client)
}

// relay exec executes a command on the specified peer node
func (c *ClusterDatabase) relayExec(peer string, conn resp.Connection, args [][]byte) resp.Reply {
	if peer == c.self {
		return c.db.Exec(conn, args)
	}
	client, err := c.getPeerClient(peer)
	if err != nil {
		return reply.MakeStandardErrorReply(err.Error())
	}
	defer func() {
		c.returnPeerClient(peer, client)
	}()
	client.Send(utils.ToCmdLine("SELECT", strconv.Itoa(conn.GetDBIndex())))
	return client.Send(args)
}

// broadcastExec executes a command on all peer nodes
func (c *ClusterDatabase) broadcastExec(conn resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, peer := range c.nodes {
		result := c.relayExec(peer, conn, args)
		results[peer] = result
	}
	return results
}
