package cluster

import (
	"fmt"
	"redigo/datastruct/set"
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

	// Hash operations
	routerMap["hset"] = defaultFunc      // hset key field value
	routerMap["hsetnx"] = defaultFunc    // hsetnx key field value
	routerMap["hget"] = defaultFunc      // hget key field
	routerMap["hexists"] = defaultFunc   // hexists key field
	routerMap["hdel"] = defaultFunc      // hdel key field [field ...]
	routerMap["hlen"] = defaultFunc      // hlen key
	routerMap["hgetall"] = defaultFunc   // hgetall key
	routerMap["hkeys"] = defaultFunc     // hkeys key
	routerMap["hvals"] = defaultFunc     // hvals key
	routerMap["hmget"] = defaultFunc     // hmget key field [field ...]
	routerMap["hmset"] = defaultFunc     // hmset key field value [field value ...]
	routerMap["hencoding"] = defaultFunc // hencoding key (custom command)

	// Set operations
	routerMap["sadd"] = defaultFunc        // sadd key member [member ...]
	routerMap["scard"] = defaultFunc       // scard key
	routerMap["sismember"] = defaultFunc   // sismember key member
	routerMap["smembers"] = defaultFunc    // smembers key
	routerMap["srem"] = defaultFunc        // srem key member [member ...]
	routerMap["spop"] = defaultFunc        // spop key [count]
	routerMap["srandmember"] = defaultFunc // srandmember key [count]

	// Set operations - multi-key commands (need special handling)
	routerMap["sunion"] = setUnionFunc               // sunion key [key ...]
	routerMap["sunionstore"] = setUnionStoreFunc     // sunionstore destination key [key ...]
	routerMap["sinter"] = setIntersectFunc           // sinter key [key ...]
	routerMap["sinterstore"] = setIntersectStoreFunc // sinterstore destination key [key ...]
	routerMap["sdiff"] = setDiffFunc                 // sdiff key [key ...]
	routerMap["sdiffstore"] = setDiffStoreFunc       // sdiffstore destination key [key ...]

	return routerMap
}

// defaultFunc is a default function that executes a command on the cluster database
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	fmt.Println("args:", args)
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key)
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

// setUnionFunc handles SUNION command in cluster mode
func setUnionFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.MakeArgNumErrReply("sunion")
	}

	// Create a set to hold the union result
	result := set.NewHashSet()

	// Process each key individually
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		peer := cluster.peerPicker.PickNode(key)

		// Create SMEMBERS command for this key
		smembersArgs := make([][]byte, 2)
		smembersArgs[0] = []byte("SMEMBERS")
		smembersArgs[1] = args[i]

		// Execute SMEMBERS on the appropriate node
		nodeReply := cluster.relayExec(peer, conn, smembersArgs)

		// Process the reply
		if mbReply, ok := nodeReply.(*reply.MultiBulkReply); ok {
			// Add each member to our result set
			for _, member := range mbReply.Args {
				result.Add(string(member))
			}
		} else if reply.IsErrReply(nodeReply) {
			return nodeReply // Forward any errors
		}
	}

	// Convert the result set to [][]byte format for the response
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

/**
 * setUnionStoreFunc handles SUNIONSTORE command in cluster mode
 * First gets the union using setUnionFunc, then stores it in the destination key
 */
func setUnionStoreFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return reply.MakeArgNumErrReply("sunionstore")
	}

	// Get the destination key and its node
	destKey := string(args[1])
	destPeer := cluster.peerPicker.PickNode(destKey)

	// Get the union of source sets
	sourceArgs := make([][]byte, len(args)-1)
	sourceArgs[0] = []byte("SUNION")
	copy(sourceArgs[1:], args[2:])

	// Use the above SUNION function to get the union
	unionReply := setUnionFunc(cluster, conn, sourceArgs)

	if mbReply, ok := unionReply.(*reply.MultiBulkReply); ok {
		// First delete the destination key (if exists)
		delArgs := make([][]byte, 2)
		delArgs[0] = []byte("DEL")
		delArgs[1] = args[1]
		cluster.relayExec(destPeer, conn, delArgs)

		if len(mbReply.Args) > 0 {
			// Create a new set on the destination node
			storeArgs := make([][]byte, len(mbReply.Args)+2)
			storeArgs[0] = []byte("SADD")
			storeArgs[1] = args[1]
			copy(storeArgs[2:], mbReply.Args)

			reply := cluster.relayExec(destPeer, conn, storeArgs)
			return reply
		}

		// If the union is empty, return 0
		return reply.MakeIntReply(0)
	}

	// Return error
	return unionReply
}

/**
 * setIntersectFunc handles SINTER command in cluster mode
 * Processes each key individually and computes the intersection
 */
func setIntersectFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.MakeArgNumErrReply("sinter")
	}

	// If there's only one key, just return its members
	if len(args) == 2 {
		key := string(args[1])
		peer := cluster.peerPicker.PickNode(key)

		// Create SMEMBERS command for this key
		smembersArgs := make([][]byte, 2)
		smembersArgs[0] = []byte("SMEMBERS")
		smembersArgs[1] = args[1]

		return cluster.relayExec(peer, conn, smembersArgs)
	}

	// Store the set members from each key
	var allSets []map[string]bool

	// Process each key separately
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		peer := cluster.peerPicker.PickNode(key)

		// Create SMEMBERS command for this key
		smembersArgs := make([][]byte, 2)
		smembersArgs[0] = []byte("SMEMBERS")
		smembersArgs[1] = args[i]

		// Execute SMEMBERS command on the appropriate node
		nodeReply := cluster.relayExec(peer, conn, smembersArgs)

		if mbReply, ok := nodeReply.(*reply.MultiBulkReply); ok {
			// Convert response to a set for intersection
			memberSet := make(map[string]bool)
			for _, member := range mbReply.Args {
				memberSet[string(member)] = true
			}

			// If any set is empty, the intersection is empty
			if len(memberSet) == 0 {
				return reply.MakeMultiBulkReply([][]byte{})
			}

			allSets = append(allSets, memberSet)
		} else if reply.IsErrReply(nodeReply) {
			return nodeReply
		}
	}

	// If no sets were obtained, return empty result
	if len(allSets) == 0 {
		return reply.MakeMultiBulkReply([][]byte{})
	}

	// Calculate intersection
	result := make(map[string]bool)
	// Initialize result with all elements from the first set
	for member := range allSets[0] {
		result[member] = true
	}

	// Intersect with subsequent sets
	for i := 1; i < len(allSets); i++ {
		nextSet := allSets[i]
		for member := range result {
			if !nextSet[member] {
				delete(result, member)
			}
		}

		// If intersection is empty, return early
		if len(result) == 0 {
			break
		}
	}

	// Convert result to response format
	members := make([][]byte, 0, len(result))
	for member := range result {
		members = append(members, []byte(member))
	}

	return reply.MakeMultiBulkReply(members)
}

/**
 * setDiffFunc handles SDIFF command in cluster mode
 * Gets all members from the first set, then removes any members found in other sets
 */
func setDiffFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.MakeArgNumErrReply("sdiff")
	}

	// Get the first set (base set)
	firstKey := string(args[1])
	firstPeer := cluster.peerPicker.PickNode(firstKey)

	// Create SMEMBERS command for the first key
	smembersArgs := make([][]byte, 2)
	smembersArgs[0] = []byte("SMEMBERS")
	smembersArgs[1] = args[1]

	firstSetReply := cluster.relayExec(firstPeer, conn, smembersArgs)

	if !reply.IsMultiBulkReply(firstSetReply) {
		if reply.IsErrReply(firstSetReply) {
			return firstSetReply
		}
		return reply.MakeMultiBulkReply([][]byte{})
	}

	// Add the members of the first set to the result set
	firstSetMembers := firstSetReply.(*reply.MultiBulkReply)
	result := make(map[string]bool)
	for _, member := range firstSetMembers.Args {
		result[string(member)] = true
	}

	// If there is only one set, just return all its members
	if len(args) == 2 {
		return firstSetReply
	}

	// Remove members of other sets from the result set
	for i := 2; i < len(args); i++ {
		key := string(args[i])
		peer := cluster.peerPicker.PickNode(key)

		// Create SMEMBERS command for this key
		smembersArgs := make([][]byte, 2)
		smembersArgs[0] = []byte("SMEMBERS")
		smembersArgs[1] = args[i]

		nodeReply := cluster.relayExec(peer, conn, smembersArgs)

		if mbReply, ok := nodeReply.(*reply.MultiBulkReply); ok {
			// Remove members of this set from the result set
			for _, member := range mbReply.Args {
				delete(result, string(member))
			}
		} else if reply.IsErrReply(nodeReply) {
			return nodeReply
		}

		// If the difference is already empty, return early
		if len(result) == 0 {
			break
		}
	}

	// Convert result to response format
	members := make([][]byte, 0, len(result))
	for member := range result {
		members = append(members, []byte(member))
	}

	return reply.MakeMultiBulkReply(members)
}

/**
 * setDiffStoreFunc handles SDIFFSTORE command in cluster mode
 * First gets the difference using setDiffFunc, then stores it in the destination key
 */
func setDiffStoreFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return reply.MakeArgNumErrReply("sdiffstore")
	}

	// Get the destination key and its node
	destKey := string(args[1])
	destPeer := cluster.peerPicker.PickNode(destKey)

	// Get the difference of source sets
	sourceArgs := make([][]byte, len(args)-1)
	sourceArgs[0] = []byte("SDIFF")
	copy(sourceArgs[1:], args[2:])

	// Use the setDiffFunc to get the difference
	diffReply := setDiffFunc(cluster, conn, sourceArgs)

	if mbReply, ok := diffReply.(*reply.MultiBulkReply); ok {
		// First delete the destination key (if exists)
		delArgs := make([][]byte, 2)
		delArgs[0] = []byte("DEL")
		delArgs[1] = args[1]
		cluster.relayExec(destPeer, conn, delArgs)

		if len(mbReply.Args) > 0 {
			// Create a new set on the destination node
			storeArgs := make([][]byte, len(mbReply.Args)+2)
			storeArgs[0] = []byte("SADD")
			storeArgs[1] = args[1]
			copy(storeArgs[2:], mbReply.Args)

			rep := cluster.relayExec(destPeer, conn, storeArgs)

			// For SDIFFSTORE, we need to return the cardinality of the result
			if intReply, ok := rep.(*reply.IntReply); ok {
				return reply.MakeIntReply(intReply.Code)
			}
			return rep
		}

		// If the difference is empty, return 0
		return reply.MakeIntReply(0)
	}

	// Return error if we couldn't get the difference
	return diffReply
}

/**
 * setIntersectStoreFunc handles SINTERSTORE command in cluster mode
 * First gets the intersection using setIntersectFunc, then stores it in the destination key
 */
func setIntersectStoreFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return reply.MakeArgNumErrReply("sinterstore")
	}

	// Get the destination key and its node
	destKey := string(args[1])
	destPeer := cluster.peerPicker.PickNode(destKey)

	// Get the intersection of source sets
	sourceArgs := make([][]byte, len(args)-1)
	sourceArgs[0] = []byte("SINTER")
	copy(sourceArgs[1:], args[2:])

	// Use the setIntersectFunc to get the intersection
	intersectReply := setIntersectFunc(cluster, conn, sourceArgs)

	if mbReply, ok := intersectReply.(*reply.MultiBulkReply); ok {
		// First delete the destination key (if exists)
		delArgs := make([][]byte, 2)
		delArgs[0] = []byte("DEL")
		delArgs[1] = args[1]
		cluster.relayExec(destPeer, conn, delArgs)

		if len(mbReply.Args) > 0 {
			// Create a new set on the destination node
			storeArgs := make([][]byte, len(mbReply.Args)+2)
			storeArgs[0] = []byte("SADD")
			storeArgs[1] = args[1]
			copy(storeArgs[2:], mbReply.Args)

			rep := cluster.relayExec(destPeer, conn, storeArgs)

			// For SINTERSTORE, we need to return the cardinality of the result
			if intReply, ok := rep.(*reply.IntReply); ok {
				return reply.MakeIntReply(intReply.Code)
			}
			return rep
		}

		// If the intersection is empty, return 0
		return reply.MakeIntReply(0)
	}

	// Return error if we couldn't get the intersection
	return intersectReply
}
