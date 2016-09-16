package main

import (
	"errors"
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
	"github.com/bsm/sarama-cluster"
)

var (
	doneCh                       = make(chan struct{})
	msgs                         = make(chan string)
	msgCount                     int
	producedValue, consumedValue int
)

var (
	groupID = flag.String("group", "may_group", "REQUIRED: The shared consumer group name")
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

	config := cluster.NewConfig()

	client, err := cluster.NewClient(brokers, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	producer, err := sarama.NewSyncProducerFromClient(client.Client)
	if err != nil {
		log.Fatal(err.Error())
	}

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

			if _, _, err = producer.SendMessage(msg); err != nil {
				log.Fatal(fmt.Errorf(err.Error()))
			} else {
				producedValue++
			}
		}
		if err = producer.Close(); err != nil {
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

	if *verbose {
		sarama.Logger = logger
	} else {
		config.Consumer.Return.Errors = true
		config.Group.Return.Notifications = true
	}

	consumer, err := cluster.NewConsumer(brokers, *groupID, strings.Split(*topic, ","), config)
	// consumer, err := cluster.NewConsumerFromClient(client, *groupID, strings.Split(*topic, ","))
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

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var wg sync.WaitGroup
	wg.Add(*goroutines)

	// warn to return all goroutines when Interrupt signal triggered
	go handleCtrlC(signals, doneCh)
	offset := &Offset{}
	for index := 0; index < *goroutines; index++ {
		go consume(consumer, doneCh, &wg, offset)
	}

	wg.Wait()

	count, err := offset.getTotalOffset()
	if err != nil {
		log.Fatal(err.Error())
	}

	if producedValue == int(count) {
		fmt.Println("produced value and total count is:", producedValue, count)
	}

	if producedValue == consumedValue {
		fmt.Println(fmt.Sprintf("producer:%d and consumer:%d checking is finihed as successfully", producedValue, consumedValue))
	} else {
		fmt.Println(fmt.Sprintf("producer:%d and consumer:%d checking is failed", producedValue, consumedValue))
	}

	if err := consumer.Close(); err != nil {
		logger.Println("Failed to close consumer: ", err)
	}
}

type Offset struct {
	count map[int32][]int64
	// min   int64
	// max   int64
	mu sync.Mutex
}

func (o *Offset) setPartitionAndOffset(partition int32, offset int64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.count == nil {
		o.count = make(map[int32][]int64)
	}

	if _, ok := o.count[partition]; !ok {
		// if we init array first time
		// then we dont have another offset value to assign
		o.count[partition] = make([]int64, 2)
	}

	o.setOffset(partition, offset)

	return
}

var (
	ErrNilOffsetCount = errors.New("offset count is nil")
)

func (o *Offset) getTotalOffset() (int64, error) {
	var totalCount int64
	if o.count == nil {
		return 0, ErrNilOffsetCount
	}
	for _, v := range o.count {
		totalCount += v[1] - v[0] + 1
	}

	return totalCount, nil
}

func (o *Offset) setOffset(partition int32, offset int64) {
	v, ok := o.count[partition]
	if !ok {
		o.count[partition] = make([]int64, 2)
	}

	min, max := v[0], v[1]

	if min == 0 && max == 0 {
		min = offset
		max = offset
	}

	if offset < min {
		min = offset
	}
	if offset > max {
		max = offset
	}

	// update min and max value for offsets
	v[0], v[1] = min, max
	// set updated offset values into count map
	o.count[partition] = v

	return
}

func consume(consumer *cluster.Consumer, doneCh chan struct{}, wg *sync.WaitGroup, o *Offset) {
	defer wg.Done()

	for {
		select {
		case msg, ok := <-consumer.Messages():
			if !ok {
				return
			}

			consumer.MarkOffset(msg, "")
			o.setPartitionAndOffset(msg.Partition, msg.Offset)
			// use mutex here to prevent from race condition
			mu.Lock()
			consumedValue++
			mu.Unlock()
		case <-doneCh:
			return
		}
	}
}

func handleCtrlC(c chan os.Signal, cc chan struct{}) {
	<-c
	close(cc)
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
