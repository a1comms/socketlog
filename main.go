package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/logging"
)

var workerDataChannel = make(chan []byte, 4096)

func main() {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "a1comms-legacy"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		log.Fatalf("PING to logging service failed: %s", err)
	}

	// Sets the name of the log to write to.
	logName := "dev-log"

	// Selects the log to write to.
	logger := client.Logger(logName)

	if len(os.Args) != 2 {
		log.Fatalf("no SOCKET file specified")
	}

	l, err := net.Listen("unix", os.Args[1])
	if err != nil {
		log.Fatalf("UNIX SOCKET ERROR: %s", err)
	}
	// Just work with defer here; this works as long as the signal handling
	// happens in the main Go routine.
	defer l.Close()

	// chmod the socket so everyone can connect.
	if err := os.Chmod(os.Args[1], 0777); err != nil {
		log.Fatal(err)
	}

	// Make sure the server does not block the main
	go func() {
		for i := 0; i < 10; i++ {
			go logIngestionWorker(ctx, logger, workerDataChannel)
		}

		for {
			fd, err := l.Accept()
			if err != nil {
				log.Printf("Error while accepting connection: %s", err)
			}

			go acceptMessages(fd)
		}
	}()

	// Use a buffered channel so we don't miss any signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	// ...and exit, running all the defer statements
}

func acceptMessages(fd net.Conn) {
	defer fd.Close()

	bfd := bufio.NewScanner(fd)

	for bfd.Scan() {
		workerDataChannel <- bfd.Bytes()
	}

	err := bfd.Err()
	if err != nil {
		log.Printf("Error while reading from SOCKET: %s", err)
	}
}

type JSONLogEntry struct {
	Severity string          `json:"severity"`
	Message  string          `json:"message"`
	Context  json.RawMessage `json:"context"`
}

func logIngestionWorker(ctx context.Context, logger *logging.Logger, dataChannel <-chan []byte) {
	for {
		data, more := <-dataChannel
		if more {
			logData := &JSONLogEntry{}
			err := json.Unmarshal(data, logData)
			if err != nil {
				log.Printf("Failed to unmarshal JSON in log message: %s", err)
				continue
			}

			logEntry := logging.Entry{
				Payload: logData,
			}

			switch logData.Severity {
			case "Debug":
				logEntry.Severity = logging.Debug
			case "Info":
				logEntry.Severity = logging.Info
			case "Notice":
				logEntry.Severity = logging.Notice
			case "Warning":
				logEntry.Severity = logging.Warning
			case "Error":
				logEntry.Severity = logging.Error
			case "Critical":
				logEntry.Severity = logging.Critical
			case "Alert":
				logEntry.Severity = logging.Alert
			case "Emergency":
				logEntry.Severity = logging.Emergency
			default:
				logEntry.Severity = logging.Default
			}

			logger.Log(logEntry)
		} else {
			return
		}
	}
}
