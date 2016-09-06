package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/kr/pretty"
)

var (
	doneCh                       = make(chan bool)
	msgs                         = make(chan string)
	msgCount                     int
	producedValue, consumedValue int
)

var (
	groupID = flag.String("group", "", "REQUIRED: The shared consumer group name")
	// brokerList = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	// topicList  = flag.String("topics", "", "REQUIRED: The comma separated list of topics to consume")
	offset  = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	verbose = flag.Bool("verbose", false, "Whether to turn on sarama logging")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

var (
	element    = flag.Int("element", 10, "default message number")
	interval   = flag.Duration("interval", 1*time.Second, "default duration to generate strings")
	goroutines = flag.Int("goroutines", 2, "default go routines count")
	address    = flag.String("address", "localhost:9092", "addresses for kafka")
	topic      = flag.String("topic", "defaultTopic", "topic name for kafka")
)

func main() {
	flag.Parse()

	if *element <= 0 {
		log.Fatal(fmt.Errorf("element should be greater than zero"))
	}

	if *goroutines > *element {
		log.Fatal(fmt.Errorf("err: goroutines should be less than element number"))
	}

	brokers := strings.Split(*address, ",")

	config := cluster.NewConfig()

	client, err := cluster.NewClient(brokers, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// client, err := sarama.NewClient(brokers, &config.Config)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	producer, err := sarama.NewSyncProducerFromClient(client.Client)
	if err != nil {
		log.Fatal(err.Error())
	}

	go func() {
		defer func() {
			time.Sleep(*interval / 10)
			close(doneCh)
		}()
		fmt.Println("BURAYA1")
		for i := 0; i < *element; i++ {
			fmt.Println("BURAYA2")
			strTime := strconv.Itoa(int(time.Now().Unix()))
			val := fmt.Sprintf("Something Cool: %d", i)

			msg := &sarama.ProducerMessage{
				Topic: *topic,
				Key:   sarama.StringEncoder(strTime),
				Value: sarama.StringEncoder(val),
			}

			// use ticker as time.Sleep
			<-time.Tick(*interval)

			fmt.Println("BURAYA3")
			if _, _, err := producer.SendMessage(msg); err != nil {
				log.Fatal(fmt.Errorf(err.Error()))
			} else {
				fmt.Println("Message Gonderildi:", msg)
			}

		}
		if err := producer.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	switch *offset {
	case "oldest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		printUsageErrorAndExit("-offset should be `oldest` or `newest`")
	}

	config.Consumer.Return.Errors = true

	// consumer, err := cluster.NewConsumer(brokers, *groupID, strings.Split(*topic, ","), config)
	consumer, err := cluster.NewConsumerFromClient(client, *groupID, strings.Split(*topic, ","))
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
			// fmt.Println("Message geldi", msg)
			fmt.Printf("msg %# v", pretty.Formatter(msg))
			fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Value)
			consumer.MarkOffset(msg, "")
		}
	}()

	wait := make(chan os.Signal)
	signal.Notify(wait, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-wait

	if err := consumer.Close(); err != nil {
		logger.Println("Failed to close consumer: ", err)
	}

}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}

func printErrorAndExit(code int, format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	os.Exit(code)
}
