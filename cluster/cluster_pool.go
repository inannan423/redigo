package cluster

import (
	"context"
	"errors"
	"redigo/resp/client"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type connectionFactory struct {
	Peer string // peer node id
}

// MakeObject creates a new connection object
func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	c, err := client.MakeClient(f.Peer)
	if err != nil {
		return nil, err
	}
	c.Start()
	return pool.NewPooledObject(c), nil
}

// DestroyObject destroys a connection object
func (f *connectionFactory) DestroyObject(ctx context.Context, pooledObject *pool.PooledObject) error {
	c, ok := pooledObject.Object.(*client.Client)
	if !ok {
		return errors.New("invalid connection type")
	}
	c.Close()
	return nil
}

// ValidateObject validates a connection object
// We don't need it, just return true
func (f *connectionFactory) ValidateObject(ctx context.Context, pooledObject *pool.PooledObject) bool {
	return true
}

// PassivateObject passivates a connection object
// We don't need it, just return nil
func (f *connectionFactory) PassivateObject(ctx context.Context, pooledObject *pool.PooledObject) error {
	return nil
}

// ActivateObject activates a connection object
// We don't need it, just return nil
func (f *connectionFactory) ActivateObject(ctx context.Context, pooledObject *pool.PooledObject) error {
	return nil
}
