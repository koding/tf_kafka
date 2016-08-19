package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// channel variables
var (
	done = make(chan bool)
	msgs = make(chan string)
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

}

var (
	producer, consumer int
)

func produce(x time.Duration, n int, c chan struct{}) {
	defer func() {
		close(msgs)
		done <- true
	}()

	for i := 0; i < n; i++ {
		select {
		case <-c:
			return
		case <-time.After(x):
			str := fmt.Sprintf("/key-%d", i)
			msgs <- str
			// we dont need to lock process coz there is single process
			producer++
		}
	}
}

func consume(c *zk.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		msg, ok := <-msgs
		if !ok {
			return
		}
		_, err := c.Create(msg, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				continue
			}
			fmt.Println(fmt.Errorf("%v", err.Error()))
			return
		}

		if err == nil {
			mu.Lock()
			consumer++
			mu.Unlock()
		}
	}
}

// cleanUp cleans the all znodes in zookeper data structure
func cleanUp(c *zk.Conn) error {
	children, _, _, err := c.ChildrenW("/")
	if err != nil {
		return err
	}

	for _, child := range children {
		// we can't delete /zookeeper path from the znodes
		if child == "zookeeper" {
			continue
		}
		// delete operation takes path parameter
		// then it should have '/' as first index
		if err := c.Delete("/"+child, 0); err != nil {
			return err
		}
	}

	return nil
}

func handleCtrlC(c chan os.Signal, cc chan struct{}) {
	// handle ctrl+c event here
	<-c
	close(cc)
}




import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"strconv"

	"github.com/Shopify/sarama"
)

func producer()  {
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
		// Should not reach here
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			// Should not reach here
			panic(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var enqueued, errors int
	doneCh := make(chan struct{})
	go func() {
		for {

			time.Sleep(500 * time.Millisecond)

			strTime := strconv.Itoa(int(time.Now().Unix()))
			msg := &sarama.ProducerMessage{
				Topic: "important",
				Key:   sarama.StringEncoder(strTime),
				Value: sarama.StringEncoder("Something Cool"),
			}
			select {
			case producer.Input() <- msg:
				enqueued++
				fmt.Println("Produce message")
			case err := <-producer.Errors():
				errors++
				fmt.Println("Failed to produce message:", err)
			case <-signals:
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
}
