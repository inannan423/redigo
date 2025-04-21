package cluster

import (
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

	return routerMap
}

// defaultFunc is a default function that executes a command on the cluster database
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
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
	replies := cluster.broadcastExec(c, args)
	var errReply reply.ErrorReply
	var deleted int64 = 0
	for _, v := range replies {
		if reply.IsErrReply(v) {
			errReply = v.(reply.ErrorReply)
			break
		}
		intReply, ok := v.(*reply.IntReply)
		if !ok {
			errReply = reply.MakeStandardErrorReply("error")
		}
		deleted += intReply.Code
	}

	if errReply == nil {
		return reply.MakeIntReply(deleted)
	}
	return reply.MakeStandardErrorReply("error occurs: " + errReply.Error())
}

// selectFunc is a function that executes a command on the cluster database
func selectFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(conn, args)
}
