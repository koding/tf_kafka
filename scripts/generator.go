package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

func as() {
	io.Copy()

}

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

	config.Producer.Partitioner = NewRandomPartitioner

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

//
///
/// implement this new consumer
//
//

func main() {
	flag.Parse()

	if *groupID == "" {
		printUsageErrorAndExit("You have to provide a -group name.")
	} else if *brokerList == "" {
		printUsageErrorAndExit("You have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable.")
	} else if *topicList == "" {
		printUsageErrorAndExit("You have to provide -topics as a comma-separated list.")
	}

	// Init config
	config := cluster.NewConfig()
	if *verbose {
		sarama.Logger = logger
	} else {
		config.Consumer.Return.Errors = true
		config.Group.Return.Notifications = true
	}

	switch *offset {
	case "oldest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		printUsageErrorAndExit("-offset should be `oldest` or `newest`")
	}

	// Init consumer, consume errors & messages
	consumer, err := cluster.NewConsumer(strings.Split(*brokerList, ","), *groupID, strings.Split(*topicList, ","), config)
	if err != nil {
		printErrorAndExit(69, "Failed to start consumer: %s", err)
	}

	go func() {
		for err := range consumer.Errors() {
			logger.Printf("Error: %s\n", err.Error())
		}
	}()

	go func() {
		for note := range consumer.Notifications() {
			logger.Printf("Rebalanced: %+v\n", note)
		}
	}()

	go func() {
		for msg := range consumer.Messages() {
			fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Value)
			consumer.MarkOffset(msg, "")
		}
	}()

	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-wait

	if err := consumer.Close(); err != nil {
		logger.Println("Failed to close consumer: ", err)
	}
}

func printErrorAndExit(code int, format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	os.Exit(code)
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}
