package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StockMessage struct {
	Company   string  `json:"company"`
	EventType string  `json:"eventType"`
	Price     float64 `json:"price"`
}

type StockDocument struct {
	Company  string  `bson:"company"`
	AvgPrice float64 `bson:"avgPrice"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Get environment variables
	rabbitMQURL := getEnvOrDefault("RABBITMQ_URL", "amqp://stockmarket:supersecret123@rabbitmq:5672/")
	mongoDBURL := getEnvOrDefault("MONGODB_URL", "mongodb://stockmarket1:27017,stockmarket2:27018,stockmarket3:27019/mydatabase?replicaSet=rs0")
	queueName := getEnvOrDefault("QUEUE_NAME", "AAPL")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDBURL))
	failOnError(err, "Failed to connect to MongoDB")
	defer client.Disconnect(ctx)

	collection := client.Database("stockmarket").Collection("stocks")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		var messages []StockMessage
		var totalPrice float64

		for d := range msgs {
			var stockEvent StockMessage
			err := json.Unmarshal(d.Body, &stockEvent)
			if err != nil {
				log.Printf("Error parsing message: %s", err)
				continue
			}

			messages = append(messages, stockEvent)
			totalPrice += stockEvent.Price

			// Process batch when we have 1000 messages
			if len(messages) >= 1000 {
				avgPrice := totalPrice / float64(len(messages))

				// Update document with new average price
				update := bson.D{
					{"$set", bson.D{
						{"company", queueName},
						{"avgPrice", avgPrice},
					}},
				}

				opts := options.Update().SetUpsert(true)
				filter := bson.D{{"company", queueName}}

				_, err = collection.UpdateOne(context.Background(), filter, update, opts)
				if err != nil {
					log.Printf("Error updating MongoDB: %s", err)
				} else {
					log.Printf("Updated average price for %s: %.2f", queueName, avgPrice)
				}

				// Reset batch
				messages = []StockMessage{}
				totalPrice = 0
			}
		}
	}()

	log.Printf(" [*] Waiting for messages in %s queue. To exit press CTRL+C", queueName)
	wg.Wait()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
