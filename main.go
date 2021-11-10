package main

import (
	"log"
	"time"
	"fmt"
	"os"
	"net/http"

	"github.com/joho/godotenv"
	amqp "github.com/streadway/amqp"
	"github.com/gorilla/mux"
	

)

const queueName string = "jobQueue"
const hostString string = "127.0.0.1:8000"

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Handler takes income request and tries to create an instant Job ID
// Place the job in the queue, returns the Job ID to the caller

func getServer(name string) JobServer {
	/*
		Create server object and initiates
		the channel and queue details to publish messages
	*/
	err := godotenv.Load()
	amqp_connection_string := fmt.Sprintf("%s", os.Getenv("AMQP"))
	conn, err := amqp.Dial(amqp_connection_string)
	handleError(err, "Dialing failed to RabbitMQ broker")

	channel, err := conn.Channel()
	handleError(err, "Fetching channel failed")

	jobQueue, err := channel.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)

	handleError(err, "Job queue creation failed")
	return JobServer{Conn: conn, Channel: channel, Queue: jobQueue}
}

func main() {
	jobServer := getServer(queueName)

	// Start Workers
	go func(conn *amqp.Connection) {
		workerProcess := Workers{
			conn: jobServer.Conn,
		}
		workerProcess.run()
	}(jobServer.Conn)

	router := mux.NewRouter()
	router.HandleFunc("/job/database", job.Server.asyncDBHandler)
	router.HandleFunc("/job/mail", jobServer.asyncMailHandler)
	router.HandleFunc("/job/callback", jobServer.asyncCallbackHandler)

	httpServer := &http.Server{
		Handler: router,
		Addr: hostString,
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
	}

	log.Fatal(httpServer.ListenAndServe())

	defer jobServer.Channel.Close()
	defer jobServer.Conn.Close()
}