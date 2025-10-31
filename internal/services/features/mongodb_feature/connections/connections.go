package connections

import (
	"context"
	"log/slog"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connections struct {
	ctx     context.Context
	mutex   sync.Mutex
	clients map[string]*mongo.Client
}

func NewConnections(ctx context.Context) *Connections {
	return &Connections{
		ctx:     ctx,
		clients: make(map[string]*mongo.Client),
	}
}

func (c *Connections) Get(url string, database string, collection string) (*mongo.Collection, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	client, exists := c.clients[url]

	if !exists {
		clientOptions := options.Client().ApplyURI(url)

		var err error

		client, err = mongo.Connect(context.Background(), clientOptions)

		if err != nil {
			return nil, err
		}

		c.clients[url] = client
	}

	return client.Database(database).Collection(collection), nil
}

func (c *Connections) Close() error {
	slog.Warn("Closing mongodb connections")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, client := range c.clients {
		if err := client.Disconnect(c.ctx); err != nil {
			return err
		}
	}

	return nil
}
