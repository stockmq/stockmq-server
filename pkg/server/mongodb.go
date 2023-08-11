package server

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Configuration.
type MongoDBConfig struct {
	Enabled    bool   `xml:"Enabled"`
	URL        string `xml:"URL"`
	RetryDelay int    `xml:"RetryDelay"`
	Database   string `xml:"Database"`
	Candles    string `xml:"Candles"`
	Quotes     string `xml:"Quotes"`
}

// DefaultMongoDBConfig returns default MongoDB config.
func DefaultMongoDBConfig() MongoDBConfig {
	return MongoDBConfig{
		Enabled:    false,
		URL:        "mongodb://localhost:27017",
		RetryDelay: 5,
		Database:   "stockmq",
		Candles:    "candles",
		Quotes:     "quotes",
	}
}

// InfluxDBConfig returns InfluxDB configuration.
func (s *Server) MongoDBConfig() MongoDBConfig {
	return s.ServerConfig().MongoDB
}

// IsMongoDBReconnecting returns whether MongoDB is scheduled to reconnect.
func (s *Server) IsMongoDBReconnecting() bool {
	return s.mongoReconn.Load()
}

// CloseMongoDB closes the MongoDB connection.
func (s *Server) CloseMongoDB() {
	s.mongoMu.Lock()
	defer s.mongoMu.Unlock()
	if s.mongoClient != nil {
		if err := s.mongoClient.Disconnect(context.TODO()); err != nil {
			s.logger.Error("MongoDB: cannot Disconnect()")
		}

		s.mongoClient = nil
	}
}

// StartMongoDB starts MongoDB client.
func (s *Server) StartMongoDB() {
	cfg := s.MongoDBConfig()
	s.logger.Info("Starting MongoDB connection", "url", cfg.URL)

	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.URL))
	if err != nil {
		s.HandleMongoDBError(err)
	}

	s.mongoClient = client
}

// StartMongoDB starts MongoDB client.
func (s *Server) HandleMongoDBError(err error) {
	// Do nothing if the server is shutting down or MongoDB is reconnecting
	if s.IsShutdown() || s.IsMongoDBReconnecting() {
		return
	}

	// Close MongoDB connection
	s.logger.Error("MongoDB error closing connection", "error", err)
	s.CloseMongoDB()

	// Runs goroutine to restart MongoDB connection after RetryDelay
	go func() {
		cfg := s.MongoDBConfig()
		s.mongoReconn.Store(true)
		s.logger.Info("MongoDB: Reconnecting", "url", cfg.URL, "delay", cfg.RetryDelay)

		select {
		case <-s.quitCh:
			return
		case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			s.mongoReconn.Store(false)
			s.StartMongoDB()
		}
	}()
}

// MongoDBStore persists the object in MongoDB collection
func (s *Server) MongoDBStore(object interface{}) {
	cfg := s.MongoDBConfig()

	s.mongoMu.Lock()
	client := s.mongoClient
	s.mongoMu.Unlock()

	var c string

	switch object.(type) {
	case *Candle:
		c = cfg.Candles
	case *Quote:
		c = cfg.Quotes
	default:
		return
	}

	if client != nil {
		collection := client.Database(cfg.Database).Collection(c)
		if _, err := collection.InsertOne(context.TODO(), object); err != nil {
			s.HandleMongoDBError(err)
		}
	}
}
