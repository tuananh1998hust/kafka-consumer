package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User :
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func getMongoCollection(mongoURL, dbName, collectionName string) *mongo.Collection {
	clientOptions := options.Client().ApplyURI(mongoURL)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Panic("Connect Err", err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Panic("Check Connection err", err)
	}

	fmt.Println("Connected to MongoDB ... !!")

	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	return collection
}

// func getKafkaReader(kafkaURL, topic string, partition int) *kafka.Conn {
// 	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, topic, partition)

// 	if err != nil {
// 		log.Fatal("Connection Kafka Err", err)
// 	}

// 	return conn
// }

func getKafkaReader(kafkaURL, topic, group string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		Topic:   topic,
		GroupID: group,
		// Partition: 0,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

// WriteToDB :
func WriteToDB(reader *kafka.Reader, collection *mongo.Collection) {
	// conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	// batch := conn.ReadBatch(10e3, 1e6)
	// b := make([]byte, 10e3)

	// for {
	// 	_, err := batch.Read(b)
	// 	if err != nil {
	// 		break
	// 	}

	// 	fmt.Println(string(b))
	// }

	// batch.Close()
	// msg, err := conn.ReadMessage(1e6)

	msg, err := reader.ReadMessage(context.Background())

	if err != nil {
		return
	}

	var user User

	json.Unmarshal(msg.Value, &user)

	result, err := collection.InsertOne(context.TODO(), bson.M{
		"name":       user.Name,
		"age":        user.Age,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})

	if err != nil {
		log.Println("Insert err ", err)
	}

	log.Println("Insert ID", result.InsertedID)
}

func main() {
	// get Mongo db Collection using environment variables.
	mongoURL := os.Getenv("MONGO_URL")
	// mongoURL := "mongodb://localhost:27017"
	dbName := "gokafka"
	// dbName := os.Getenv("DB_NAME")
	// collectionName := os.Getenv("collectionName")
	userCollection := getMongoCollection(mongoURL, dbName, "user")

	// get kafka reader using environment variables.
	kafkaURL := os.Getenv("KAFKA_URL")
	topic := os.Getenv("KAFKA_TOPIC")
	// kafkaURL := "localhost:29092"
	// topic := "user-topic"
	// partition := 0
	group := "group"
	// conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, topic, partition)

	// if err != nil {
	// 	log.Panic("connection kafka err", err)
	// }

	newReader := getKafkaReader(kafkaURL, topic, group)

	// conn.SetReadDeadline(time.Now().Add(time.Second))
	// batch := conn.ReadBatch(10e3, 1e6)
	// b := make([]byte, 10e3)

	// for {
	// 	_, err := batch.Read(b)
	// 	if err != nil {
	// 		break
	// 	}
	// 	fmt.Println(string(b))
	// }

	// batch.Close()

	// defer conn.Close()
	defer newReader.Close()

	var wg sync.WaitGroup

	for {
		wg.Add(1)

		go WriteToDB(newReader, userCollection)

		time.Sleep(100 * time.Millisecond)
	}

	// wg.Wait()
}
