package cluster

import (
	"context"
	"redigo/config"
	databaseinstance "redigo/database"
	"redigo/interface/database"
	"redigo/interface/resp"
	consistenthash "redigo/lib/consistent_hash"
	"redigo/lib/logger"
	"redigo/resp/reply"
	"strings"

	pool "github.com/jolestar/go-commons-pool/v2"
)

// ClusterDatabase is a cluster instance
type ClusterDatabase struct {
	self       string                      // self node id
	nodes      []string                    // cluster nodes
	peerPicker *consistenthash.NodeMap     // consistent hash ring
	peerConn   map[string]*pool.ObjectPool // connection pool for each node
	db         database.Database           // database instance
}

// MakeClusterDatabase creates a new ClusterDatabase instance
func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:       config.Properties.Self,
		db:         databaseinstance.NewStandaloneDatabase(),
		peerPicker: consistenthash.NewNodeMap(nil),
		peerConn:   make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	nodes = append(nodes, config.Properties.Peers...)
	nodes = append(nodes, config.Properties.Self)
	// Add nodes to the consistent hash ring
	cluster.peerPicker.AddNodes(nodes...)
	ctx := context.Background()
	// Create connection pools for each peer
	for _, peer := range config.Properties.Peers {
		cluster.peerConn[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{Peer: peer})
	}
	cluster.nodes = nodes
	return cluster
}

type CmdFunc func(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply

var routerMap = makeRouter()

// Exec executes a command on the cluster database
func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("ClusterDatabase Exec panic:" + err.(error).Error())
			result = reply.MakeUnknownReply()
		}
	}()

	cmdName := strings.ToLower(string(args[0]))

	if cmdFunc, ok := routerMap[cmdName]; ok {
		return cmdFunc(c, client, args)
	} else {
		result = reply.MakeStandardErrorReply("ERR unknown command '" + cmdName + "'")
	}

	return
}

// Close closes the cluster database
func (c *ClusterDatabase) Close() {
	c.db.Close()
}

// AfterClientClose is called after a client closes
func (c *ClusterDatabase) AfterClientClose(client resp.Connection) {
	c.db.AfterClientClose(client)
}
