package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

const (
	topic = "important"
)

// channel variables
var (
	doneCh                       = make(chan bool)
	done                         = make(chan bool)
	msgs                         = make(chan string)
	msgCount                     int
	producedValue, consumedValue int
)

// holds the flag variables
var (
	element    = flag.Int("element", 10, "default znode number")
	interval   = flag.Duration("interval", 1*time.Second, "default duration to generate strings")
	goroutines = flag.Int("goroutines", 2, "default go routines count")
	address    = flag.String("address", "0.0.0.0", "addresses for zookeeper")
)

var mu = &sync.Mutex{}

func main() {
	flag.Parse()

	if *goroutines > *element {
		log.Fatal(fmt.Errorf("err: goroutines should be less than element number"))
	}

	// Setup configuration
	config := sarama.NewConfig()
	// Return specifies what channels will be populated.
	// If they are set to true, you must read from
	// config.Producer.Return.Successes = true
	// The total number of times to retry sending a message (default 3).
	config.Producer.Retry.Max = 5

	// The level of acknowledgement reliability needed from the broker.
	config.Producer.RequiredAcks = sarama.WaitForAll
	brokers := []string{"localhost:9092"}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var errors int

	doneCh := make(chan struct{})

	go func() {
		defer func() {
			time.Sleep(*interval / 2)
			close(doneCh)
		}()

		for i := 0; i < *element; i++ {

			time.Sleep(*interval)

			strTime := strconv.Itoa(int(time.Now().Unix()))
			val := fmt.Sprintf("Something Cool: %d", i)

			msg := &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(strTime),
				Value: sarama.StringEncoder(val),
			}

			select {
			case producer.Input() <- msg:
				producedValue++
				fmt.Println("Produce message:", i)
			case err := <-producer.Errors():
				errors++
				fmt.Println("Failed to produce message:", err)
			case <-signals:
				doneCh <- struct{}{}
			}
		}
	}()

	// CONSUMER GOES HERE

	config.Consumer.Return.Errors = true

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}
	str, _ := master.Topics()
	fmt.Println("MASTER CONSUMER IS :", str)
	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	// How to decide partition, is it fixed value...?
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(*goroutines)

	// warn to return all goroutines when Interrupt signal triggered
	go handleCtrlC(signals, doneCh)

	for index := 0; index < *goroutines; index++ {
		go consume(consumer, doneCh, &wg)
	}

	wg.Wait()
	// <-doneCh
	fmt.Println("Processed", msgCount, "messages")
	fmt.Println("CONSUMED: %d and PRODUCED: %d", consumedValue, producedValue)
}

func consume(consumer sarama.PartitionConsumer, doneCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case err := <-consumer.Errors():
			fmt.Println(err)
		case msg := <-consumer.Messages():
			msgCount++
			fmt.Println("Received messages", string(msg.Key), string(msg.Value))
			// use mutex here to prevent from race condition
			mu.Lock()
			consumedValue++
			mu.Unlock()
		case <-doneCh:
			fmt.Println("DONE CHANNEL TRIGGERED")
			return
			// case <-time.After(*interval * 2):
			// 	fmt.Println("TIME AFTER WORKS")
			// 	return
		}
	}
}

func handleCtrlC(c chan os.Signal, cc chan struct{}) {
	// handle ctrl+c event here
	<-c
	close(cc)
}
