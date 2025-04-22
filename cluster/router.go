package cluster

import (
	"fmt"
	"redigo/interface/resp"
	"redigo/resp/reply"
)

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["exists"] = defaultFunc // exists key
	routerMap["type"] = defaultFunc   // type key
	routerMap["set"] = defaultFunc    // set key
	routerMap["get"] = defaultFunc    // get key
	routerMap["setnx"] = defaultFunc  // setnx key
	routerMap["getset"] = defaultFunc // getset key

	routerMap["ping"] = pingFunc     // ping command
	routerMap["rename"] = renameFunc // rename key
	routerMap["renamex"] = renameFunc
	routerMap["flushdb"] = flushDBFunc // flushdb command
	routerMap["del"] = delFunc         // del key
	routerMap["select"] = selectFunc   // select database

	routerMap["lpush"] = defaultFunc
	routerMap["rpush"] = defaultFunc
	routerMap["lpop"] = defaultFunc
	routerMap["rpop"] = defaultFunc
	routerMap["lrange"] = defaultFunc
	routerMap["llen"] = defaultFunc
	routerMap["lindex"] = defaultFunc
	routerMap["lset"] = defaultFunc

	return routerMap
}

// defaultFunc is a default function that executes a command on the cluster database
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	fmt.Println("args:", args)
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key)
	fmt.Println("peer:", peer)
	return cluster.relayExec(peer, conn, args)
}

// pingFunc is a function that executes a command on the cluster database
func pingFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(conn, args)
}

// renameFunc is a function that executes a command on the cluster database
func renameFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeStandardErrorReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[1])
	dest := string(args[2])

	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeStandardErrorReply("ERR source and destination keys are on different nodes")
	}

	return cluster.relayExec(srcPeer, conn, args)
}

// flushDBFunc is a function that executes a command on the cluster database
func flushDBFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcastExec(conn, args)
	var errReply reply.ErrorReply
	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return reply.MakeOKReply()
	}
	return reply.MakeStandardErrorReply("error: " + errReply.Error())
}

// delFunc is a function that executes a command on the cluster database
func delFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	// Check the number of arguments
	if len(args) < 2 {
		return reply.MakeArgNumErrReply("del")
	}

	// If there is only one key, route directly to the corresponding node
	if len(args) == 2 {
		key := string(args[1])
		peer := cluster.peerPicker.PickNode(key)
		// Note: The full command, including "DEL", needs to be passed
		fullArgs := make([][]byte, 2)
		fullArgs[0] = []byte("DEL")
		fullArgs[1] = args[1]
		return cluster.relayExec(peer, c, fullArgs)
	}

	// --- Modification starts ---
	// Handle multiple keys: group keys by node
	groupedKeys := make(map[string][][]byte) // key: peer address, value: list of keys handled by the peer
	for i := 1; i < len(args); i++ {         // Iterate over all keys to delete, starting from index 1
		key := string(args[i])
		peer := cluster.peerPicker.PickNode(key)
		if _, ok := groupedKeys[peer]; !ok {
			groupedKeys[peer] = make([][]byte, 0)
		}
		groupedKeys[peer] = append(groupedKeys[peer], args[i]) // Add the original []byte key to the list
	}

	// Execute delete operation for each node
	var deleted int64 = 0
	var firstErrReply reply.ErrorReply // Save the first encountered error

	for peer, keys := range groupedKeys {
		// Construct the DEL command for the current node: ["DEL", key1, key2, ...]
		nodeArgs := make([][]byte, len(keys)+1)
		nodeArgs[0] = []byte("DEL") // The command itself
		copy(nodeArgs[1:], keys)    // Copy the list of keys handled by this node

		// Send the command to the specific node
		nodeReply := cluster.relayExec(peer, c, nodeArgs)

		// Handle the response
		if reply.IsErrReply(nodeReply) {
			// If it is an error response, record the first error and stop processing other nodes (optional, can also choose to continue processing other nodes)
			if firstErrReply == nil {
				if errReply, ok := nodeReply.(reply.ErrorReply); ok {
					firstErrReply = errReply
				} else {
					firstErrReply = reply.MakeStandardErrorReply("unknown error from peer")
				}
			}
			// You can choose to break or continue here, depending on whether you want the entire operation to fail if one node fails
			// break // Stop and return an error if one node fails
			continue // Continue attempting to delete keys on other nodes, then summarize results or return the first error
		}

		// If it is an integer response, accumulate the number of deleted keys
		if intReply, ok := nodeReply.(*reply.IntReply); ok {
			deleted += intReply.Code
		} else {
			// If the response is neither the expected integer nor an error, treat it as an error
			if firstErrReply == nil {
				firstErrReply = reply.MakeStandardErrorReply("unexpected reply type from peer")
			}
			// break // Same as above
			continue // Same as above
		}
	}

	// If an error was encountered during processing, return the first error
	if firstErrReply != nil {
		// You can choose to return more detailed error information or just the first error
		return reply.MakeStandardErrorReply("error occurs during multi-key delete: " + firstErrReply.Error())
	}

	// If all nodes succeeded (or partial errors were ignored), return the total number of deleted keys
	return reply.MakeIntReply(deleted)
}

// selectFunc is a function that executes a command on the cluster database
func selectFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(conn, args)
}
