/*
Kafka consumer for telemetry ingestion.
Buffers incoming packets and handles network jitter.
*/

package ingestion

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// TelemetryMessage represents a telemetry packet
type TelemetryMessage struct {
	Timestamp float64                `json:"timestamp"`
	SimTime   float64                `json:"sim_time"`
	TrueState map[string]interface{} `json:"true_state"`
	IMU       map[string]interface{} `json:"imu"`
	Radar     map[string]interface{} `json:"radar,omitempty"`
	RawJSON   []byte                 `json:"-"`
}

// KafkaConsumer consumes telemetry from Kafka broker
type KafkaConsumer struct {
	reader     *kafka.Reader
	msgBuffer  chan *TelemetryMessage
	bufferSize int
	stopChan   chan bool
	wg         sync.WaitGroup
	isRunning  bool
	mu         sync.RWMutex
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers []string, topic string, bufferSize int) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		CommitInterval: 100 * time.Millisecond,
		MaxBytes:       10e6, // 10MB
	})

	return &KafkaConsumer{
		reader:     reader,
		msgBuffer:  make(chan *TelemetryMessage, bufferSize),
		bufferSize: bufferSize,
		stopChan:   make(chan bool),
		isRunning:  false,
	}
}

// Start begins consuming messages from Kafka
func (kc *KafkaConsumer) Start() error {
	kc.mu.Lock()
	if kc.isRunning {
		kc.mu.Unlock()
		return nil
	}
	kc.isRunning = true
	kc.mu.Unlock()

	kc.wg.Add(1)
	go func() {
		defer kc.wg.Done()
		kc.consume()
	}()

	log.Println("Kafka consumer started")
	return nil
}

// consume reads messages from Kafka
func (kc *KafkaConsumer) consume() {
	ctx := context.Background()
	for {
		select {
		case <-kc.stopChan:
			log.Println("Kafka consumer stopped")
			return
		default:
		}

		msg, err := kc.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("Kafka fetch error: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Parse telemetry message
		var telemetry TelemetryMessage
		if err := json.Unmarshal(msg.Value, &telemetry); err != nil {
			log.Printf("Failed to parse telemetry: %v", err)
			continue
		}

		telemetry.RawJSON = msg.Value

		// Non-blocking send to buffer (drop oldest if full)
		select {
		case kc.msgBuffer <- &telemetry:
		default:
			// Buffer full, drop oldest message
			select {
			case <-kc.msgBuffer:
				kc.msgBuffer <- &telemetry
			default:
			}
		}
	}
}

// GetMessage returns the next telemetry message (non-blocking)
func (kc *KafkaConsumer) GetMessage() (*TelemetryMessage, bool) {
	select {
	case msg := <-kc.msgBuffer:
		return msg, true
	default:
		return nil, false
	}
}

// GetMessageBlocking returns the next telemetry message (blocking)
func (kc *KafkaConsumer) GetMessageBlocking(timeout time.Duration) (*TelemetryMessage, bool) {
	select {
	case msg := <-kc.msgBuffer:
		return msg, true
	case <-time.After(timeout):
		return nil, false
	}
}

// BufferSize returns current buffer occupancy
func (kc *KafkaConsumer) BufferSize() int {
	return len(kc.msgBuffer)
}

// Stop stops the consumer
func (kc *KafkaConsumer) Stop() {
	kc.mu.Lock()
	if !kc.isRunning {
		kc.mu.Unlock()
		return
	}
	kc.isRunning = false
	kc.mu.Unlock()

	kc.stopChan <- true
	kc.wg.Wait()
	kc.reader.Close()
}
