package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig نود MongoDB
type MongoDBConfig struct {
	URI        string `json:"uri"`        // MongoDB connection URI
	Database   string `json:"database"`   // Database name
	Collection string `json:"collection"` // Default collection
}

// MongoDBExecutor اجراکننده نود MongoDB
type MongoDBExecutor struct {
	config MongoDBConfig
	client *mongo.Client
	mu     sync.RWMutex
}

// NewMongoDBExecutor ایجاد MongoDBExecutor
func NewMongoDBExecutor() node.Executor {
	return &MongoDBExecutor{}
}

// Init initializes the MongoDB node with configuration
func (e *MongoDBExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var mongoConfig MongoDBConfig
	if err := json.Unmarshal(configJSON, &mongoConfig); err != nil {
		return fmt.Errorf("invalid mongodb config: %w", err)
	}

	// Validate
	if mongoConfig.URI == "" {
		return fmt.Errorf("uri is required")
	}
	if mongoConfig.Database == "" {
		return fmt.Errorf("database is required")
	}

	e.config = mongoConfig

	// Initialize connection
	if err := e.connect(); err != nil {
		return err
	}

	return nil
}

// connect اتصال به MongoDB
func (e *MongoDBExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(e.config.URI))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}

	e.client = client
	return nil
}

// Execute اجرای نود
func (e *MongoDBExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.RLock()
	client := e.client
	e.mu.RUnlock()

	if client == nil {
		return node.Message{}, fmt.Errorf("not connected to database")
	}

	var operation, collection string
	var filter, document, update interface{}
	var limit int64 = 0

	// Get parameters from message
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}
	if c, ok := msg.Payload["collection"].(string); ok {
		collection = c
	}
	if f, ok := msg.Payload["filter"]; ok {
		filter = f
	}
	if d, ok := msg.Payload["document"]; ok {
		document = d
	}
	if u, ok := msg.Payload["update"]; ok {
		update = u
	}
	if l, ok := msg.Payload["limit"].(float64); ok {
		limit = int64(l)
	}

	// Use default collection if not provided
	if collection == "" {
		collection = e.config.Collection
	}
	if collection == "" {
		return node.Message{}, fmt.Errorf("collection is required")
	}
	if operation == "" {
		return node.Message{}, fmt.Errorf("operation is required")
	}

	coll := client.Database(e.config.Database).Collection(collection)

	switch operation {
	case "find":
		return e.find(ctx, coll, filter, limit)
	case "findOne":
		return e.findOne(ctx, coll, filter)
	case "insertOne":
		return e.insertOne(ctx, coll, document)
	case "insertMany":
		return e.insertMany(ctx, coll, document)
	case "updateOne":
		return e.updateOne(ctx, coll, filter, update)
	case "updateMany":
		return e.updateMany(ctx, coll, filter, update)
	case "deleteOne":
		return e.deleteOne(ctx, coll, filter)
	case "deleteMany":
		return e.deleteMany(ctx, coll, filter)
	case "count":
		return e.count(ctx, coll, filter)
	default:
		return node.Message{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (e *MongoDBExecutor) find(ctx context.Context, coll *mongo.Collection, filter interface{}, limit int64) (node.Message, error) {
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return node.Message{}, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"results": results,
			"count":   len(results),
		},
	}, nil
}

func (e *MongoDBExecutor) findOne(ctx context.Context, coll *mongo.Collection, filter interface{}) (node.Message, error) {
	var result map[string]interface{}
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return node.Message{
				Payload: map[string]interface{}{
					"found":  false,
					"result": nil,
				},
			}, nil
		}
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"found":  true,
			"result": result,
		},
	}, nil
}

func (e *MongoDBExecutor) insertOne(ctx context.Context, coll *mongo.Collection, document interface{}) (node.Message, error) {
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"insertedId": result.InsertedID,
		},
	}, nil
}

func (e *MongoDBExecutor) insertMany(ctx context.Context, coll *mongo.Collection, documents interface{}) (node.Message, error) {
	docs, ok := documents.([]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("documents must be an array")
	}

	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"insertedIds": result.InsertedIDs,
			"count":       len(result.InsertedIDs),
		},
	}, nil
}

func (e *MongoDBExecutor) updateOne(ctx context.Context, coll *mongo.Collection, filter, update interface{}) (node.Message, error) {
	result, err := coll.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"matchedCount":  result.MatchedCount,
			"modifiedCount": result.ModifiedCount,
			"upsertedId":    result.UpsertedID,
		},
	}, nil
}

func (e *MongoDBExecutor) updateMany(ctx context.Context, coll *mongo.Collection, filter, update interface{}) (node.Message, error) {
	result, err := coll.UpdateMany(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"matchedCount":  result.MatchedCount,
			"modifiedCount": result.ModifiedCount,
		},
	}, nil
}

func (e *MongoDBExecutor) deleteOne(ctx context.Context, coll *mongo.Collection, filter interface{}) (node.Message, error) {
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"deletedCount": result.DeletedCount,
		},
	}, nil
}

func (e *MongoDBExecutor) deleteMany(ctx context.Context, coll *mongo.Collection, filter interface{}) (node.Message, error) {
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"deletedCount": result.DeletedCount,
		},
	}, nil
}

func (e *MongoDBExecutor) count(ctx context.Context, coll *mongo.Collection, filter interface{}) (node.Message, error) {
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"count": count,
		},
	}, nil
}

// Cleanup پاکسازی منابع
func (e *MongoDBExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return e.client.Disconnect(ctx)
	}
	return nil
}
