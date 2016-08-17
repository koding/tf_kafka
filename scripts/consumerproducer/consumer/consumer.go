package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

var (
	msgCount int
)

func main() {

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Specify brokers address. This is default one
	brokers := []string{"localhost:9092"}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "important"
	// How to decide partition, is it fixed value...?
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	// msgCount := 0

	// Get signnal for finish
	doneCh := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(3)

	// warn to return all goroutines when Interrupt signal triggered
	go handleCtrlC(signals, doneCh)

	for index := 0; index < 3; index++ {
		go consume(consumer, doneCh, &wg)
	}

	// <-doneCh
	wg.Wait()

	fmt.Println("Processed", msgCount, "messages")
}

func consume(consumer sarama.PartitionConsumer, doneCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	msgCount = 0

	for {

		select {
		case err := <-consumer.Errors():
			fmt.Println(err)
		case msg := <-consumer.Messages():
			msgCount++
			fmt.Println("Received messages", string(msg.Key), string(msg.Value))
			// case <-signals:
			// 	fmt.Println("Interrupt is detected")
			// 	doneCh <- struct{}{}
		case <-doneCh:
			fmt.Println("DONE CHANNEL TRIGGERED")
			return
		case <-time.After(10 * time.Second):
			fmt.Println("TIME AFTER WORKS")
			return
		}
	}
}

func handleCtrlC(c chan os.Signal, cc chan struct{}) {
	// handle ctrl+c event here
	<-c
	close(cc)
}
