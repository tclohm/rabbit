package main

import (
	"log"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	ampq "github.com/streadway/amqp"
)

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	err := godotenv.Load()
	amqp_connection_string := fmt.Sprintf("%s", os.Getenv("AMQP"))

	conn, err := ampq.Dial(amqp_connection_string)
	handleError(err, "Dialing failed to connect to RabbitMQ broker")
	defer conn.Close()

	channel, err := conn.Channel()

	testQueue, err := channel.QueueDeclare(
		"test", // Name of the queue
		false, 	// Message is persisted or not
		false, 	// Delete message when unused
		false,	// Exclusive
		false,	// No Waiting time
		nil,	// Extra args
	)
	
	handleError(err, "Queue creation failed")

	messages, err := channel.Consume(
		testQueue.Name, // queue
		"", // consumer
		true, // auto-acknowledge
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil, // args
	)

	handleError(err, "Failed to register a consumer")

	// go-routine spawns a function that runs an infinite loop to collect
	// messages and process them
	go func() {
		for message := range messages {
			log.Printf("Received a message from the queue: %s", message.Body)
		}
	}()


	log.Println("Worker has started")
	wait := make(chan bool)
	<-wait

}