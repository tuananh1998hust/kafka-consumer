package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getMongoCollection(mongoURL, dbName, collectionName string) *mongo.Collection {
	var dbHost = os.Getenv("DBHOST")
	clientOptions := options.Client().ApplyURI(dbHost)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal("Connect Err", err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal("Check Connection err", err)
	}

	fmt.Println("Connected to MongoDB ... !!")

	db := client.Database(dbName)
	collection := db.Collection(collectionName)
	return collection
}

func getKafkaReader(kafkaURL, topic string, partition int) *kafka.Conn {
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, topic, partition)

	if err != nil {
		log.Fatal("Connection Kafka Err", err)
	}

	return conn
}

func main() {
	// get Mongo db Collection using environment variables.
	// mongoURL := os.Getenv("MONGO_URL")
	// dbName := os.Getenv("DB_NAME")
	// collectionName := os.Getenv("collectionName")
	// collection := getMongoCollection(mongoURL, dbName, collectionName)

	// get kafka reader using environment variables.
	kafkaURL := os.Getenv("KAFKA_URL")
	topic := os.Getenv("KAFKA_TOPIC")
	partition := 0

	conn, _ := kafka.DialLeader(context.Background(), "tcp", kafkaURL, topic, partition)

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	batch := conn.ReadBatch(10e3, 1e6)
	b := make([]byte, 10e3)

	for {
		_, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b))
	}

	batch.Close()
	conn.Close()
}
