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
	producedValue, consumedValue int
)

// flag variables for kafka script
var (
	groupID    = flag.String("group", "may_group", "REQUIRED: The shared consumer group name")
	offset     = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	verbose    = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	element    = flag.Int("element", 10, "default message number")
	interval   = flag.Duration("interval", 1*time.Second, "default duration to generate strings")
	goroutines = flag.Int("goroutines", 2, "default go routines count")
	address    = flag.String("address", "localhost:9092", "addresses for kafka")
	topic      = flag.String("topic", "defaultTopic", "topic name for kafka")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

// Error variables for kafka
var (
	ErrNilOffsetCount = errors.New("offset count is nil")
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
		fmt.Println(fmt.Sprintf("producer:%d and consumer:%d checking is finihed as successfully", producedValue, count))
	} else {
		fmt.Println(fmt.Sprintf("producer:%d and consumer:%d checking is failed", producedValue, count))
	}

	if err := consumer.Close(); err != nil {
		logger.Println("Failed to close consumer: ", err)
	}
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

// Offset holds the partition-offset data for kafka consumers
type Offset struct {
	// count holds the partition-offset values
	// count[partition][]int64{minimum , maximum}
	// e.g :
	// assume that message's partition is 2 , minimum offset is 100 and maximum offset is 240
	// then we will hold these data as
	// count[2][]int64{100,240}
	// first value of array is the minimum value of map
	// second value of array is the maximum value of map
	count map[int32][]int64

	// mu is the mutex for consumer counting and set offsets operations
	mu sync.Mutex
}

// setPartitionAndOffset sets the partition and the offset
func (o *Offset) setPartitionAndOffset(partition int32, offset int64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.count == nil {
		o.count = make(map[int32][]int64)
	}

	if _, ok := o.count[partition]; !ok {
		o.count[partition] = make([]int64, 2)
	}

	o.setOffset(partition, offset)

	return
}

// getTotalOffset sums all of offsets for each partition
// and return total number of operated offset
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

// setOffset set the minumum and maximum offset according to partition of consumer
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

// handleCtrlC handles the CTRL^ C keys and then closes the struct channel
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
