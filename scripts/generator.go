package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

// channel variables
var (
	doneCh                       = make(chan bool)
	msgs                         = make(chan string)
	msgCount                     int
	producedValue, consumedValue int
)

// holds the flag variables
var (
	element    = flag.Int("element", 10, "default message number")
	interval   = flag.Duration("interval", 1*time.Second, "default duration to generate strings")
	goroutines = flag.Int("goroutines", 2, "default go routines count")
	address    = flag.String("address", "localhost:9092", "addresses for kafka")
	topic      = flag.String("topic", "defaultTopic", "topic name for kafka")
)

var mu = &sync.Mutex{}

func main() {
	flag.Parse()

	if *element <= 0 {
		log.Fatal(fmt.Errorf("element should be greater than zero"))
	}

	if *goroutines > *element {
		log.Fatal(fmt.Errorf("err: goroutines should be less than element number"))
	}

	brokers := strings.Split(*address, ",")

	config := sarama.NewConfig()
	producer := initProducer(config, brokers)

	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var errors int

	doneCh := make(chan struct{})

	go func() {
		defer func() {
			time.Sleep(*interval / 10)
			close(doneCh)
		}()
		for i := 0; i < *element; i++ {

			strTime := strconv.Itoa(int(time.Now().Unix()))
			val := fmt.Sprintf("Something Cool: %d", i)

			msg := &sarama.ProducerMessage{
				Topic: *topic,
				Key:   sarama.StringEncoder(strTime),
				Value: sarama.StringEncoder(val),
			}

			// use ticker as time.Sleep
			<-time.Tick(*interval)

			select {
			case producer.Input() <- msg:
				producedValue++
			case err := <-producer.Errors():
				errors++
				fmt.Println("Failed to produce message:", err)
			case <-signals:
				doneCh <- struct{}{}
			}

		}
	}()

	master, consumer := initConsumer(config, brokers, *topic)
	defer func() {
		if err := master.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	var wg sync.WaitGroup
	wg.Add(*goroutines)

	// warn to return all goroutines when Interrupt signal triggered
	go handleCtrlC(signals, doneCh)

	for index := 0; index < *goroutines; index++ {
		go consume(consumer, doneCh, &wg)
	}

	wg.Wait()

	if producedValue == consumedValue {
		fmt.Println(fmt.Sprintf("producer:%d and consumer:%d checking is finihed as successfully", producedValue, consumedValue))
	} else {
		fmt.Println(fmt.Sprintf("producer:%d and consumer:%d checking is failed", producedValue, consumedValue))
	}

	os.Exit(0)
}

func consume(consumer sarama.PartitionConsumer, doneCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case err := <-consumer.Errors():
			fmt.Println(err)
		case <-consumer.Messages():
			msgCount++
			// use mutex here to prevent from race condition
			mu.Lock()
			consumedValue++
			mu.Unlock()
		case <-doneCh:
			return
		}
	}
}

// handle ctrl+c event here
// when CTRL^C is triggered, it closes the channels
func handleCtrlC(c chan os.Signal, cc chan struct{}) {
	<-c
	close(cc)
}

// initProducer gets the existing config, and then creates the producer
func initProducer(config *sarama.Config, brokers []string) sarama.AsyncProducer {
	// Return specifies what channels will be populated.
	// If they are set to true, you must read from
	// config.Producer.Return.Successes = true
	// The total number of times to retry sending a message (default 3).
	config.Producer.Retry.Max = 5

	// The level of acknowledgement reliability needed from the broker.
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return producer
}

// initConsumer gets existing config , then creates the consumer inside the function
func initConsumer(config *sarama.Config, brokers []string, topic string) (sarama.Consumer, sarama.PartitionConsumer) {
	// CONSUMER GOES HERE
	config.Consumer.Return.Errors = true

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal(err.Error())
	}

	return master, consumer
}
