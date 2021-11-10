package main

import (
	"log"
	"fmt"
	"os"
	"time"
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
	handleError(err, "Fetching channel failed")
	defer channel.Close()

	testQueue, err := channel.QueueDeclare(
		"test", // queue name
		false, // message to be persisted
		false, // delete message when unused
		false, // exclusive
		false, // no wait time
		nil, // extra args
	)

	handleError(err, "Queue creation failed")

	serverTime := time.Now()
	message := ampq.Publishing{
		ContentType: "text/plain",
		Body: []byte(serverTime.String()),
	}

	err = channel.Publish(
		"", // exchange
		testQueue.Name, // routing key (queue)
		false, // mandatory
		false, // immediate
		message,
	)

	handleError(err, "Failed to publish a message")
	log.Println("Successfully published a message to the queue")
}