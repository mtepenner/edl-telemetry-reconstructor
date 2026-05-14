/*
Fusion Server Main Entry Point
Consumes raw data streams and broadcasts filtered state estimates.
*/

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	estimation "edl-telemetry-reconstructor/internal/estimation"
	ingestion "edl-telemetry-reconstructor/internal/ingestion"
	publisher "edl-telemetry-reconstructor/internal/publisher"
)

func main() {
	// Command line flags
	kafkaBroker := flag.String("kafka-broker", "localhost:9092", "Kafka broker address")
	kafkaTopic := flag.String("kafka-topic", "telemetry", "Kafka topic name")
	httpPort := flag.String("http-port", "8080", "HTTP server port")
	verbose := flag.Bool("verbose", false, "Verbose logging")

	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Println("=== EDL Telemetry Fusion Engine ===")
	log.Printf("Kafka broker: %s, Topic: %s", *kafkaBroker, *kafkaTopic)

	// Initialize Kafka consumer
	consumer := ingestion.NewKafkaConsumer(
		[]string{*kafkaBroker},
		*kafkaTopic,
		100, // buffer size
	)

	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	// Initialize EKF (wait for first measurement)
	log.Println("Waiting for first telemetry message to initialize EKF...")
	var ekf *estimation.ExtendedKalmanFilter
	timeout := time.Now().Add(30 * time.Second)

	for ekf == nil && time.Now().Before(timeout) {
		msg, ok := consumer.GetMessageBlocking(1 * time.Second)
		if ok {
			imuAccel := msg.IMU["acceleration"].([]interface{})
			imuAngvel := msg.IMU["angular_velocity"].([]interface{})
			trueState := msg.TrueState
			truePos := trueState["position"].([]interface{})
			trueVel := trueState["velocity"].([]interface{})
			trueQuat := trueState["quaternion"].([]interface{})

			initialState := estimation.State{
				Position: [3]float64{
					truePos[0].(float64),
					truePos[1].(float64),
					truePos[2].(float64),
				},
				Velocity: [3]float64{
					trueVel[0].(float64),
					trueVel[1].(float64),
					trueVel[2].(float64),
				},
				Quaternion: [4]float64{
					trueQuat[0].(float64),
					trueQuat[1].(float64),
					trueQuat[2].(float64),
					trueQuat[3].(float64),
				},
				Timestamp: time.Now(),
			}

			ekf = estimation.NewExtendedKalmanFilter(initialState)
			log.Println("EKF initialized")

			// Process this first message
			imuMeasurement := estimation.IMUMeasurement{
				Acceleration: [3]float64{
					imuAccel[0].(float64),
					imuAccel[1].(float64),
					imuAccel[2].(float64),
				},
				AngularVelocity: [3]float64{
					imuAngvel[0].(float64),
					imuAngvel[1].(float64),
					imuAngvel[2].(float64),
				},
				Timestamp: time.Now(),
			}
			ekf.Predict(imuMeasurement)
		}
	}

	if ekf == nil {
		log.Fatalf("Failed to receive initial telemetry within timeout")
	}

	// Initialize WebSocket broadcaster
	broadcaster := publisher.NewWebSocketBroadcaster(60.0) // 60 Hz
	broadcaster.Start()

	// HTTP server
	http.HandleFunc("/ws", broadcaster.HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go func() {
		log.Printf("Starting HTTP server on :%s", *httpPort)
		if err := http.ListenAndServe(":"+*httpPort, nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Main processing loop
	ticker := time.NewTicker(time.Second / 60) // 60 Hz
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Fusion engine running, processing telemetry...")

	for {
		select {
		case <-ticker.C:
			// Check for new telemetry message
			msg, ok := consumer.GetMessage()
			if ok {
				// Extract IMU data
				imuData := msg.IMU
				imuAccel := imuData["acceleration"].([]interface{})
				imuAngvel := imuData["angular_velocity"].([]interface{})

				imuMeasurement := estimation.IMUMeasurement{
					Acceleration: [3]float64{
						imuAccel[0].(float64),
						imuAccel[1].(float64),
						imuAccel[2].(float64),
					},
					AngularVelocity: [3]float64{
						imuAngvel[0].(float64),
						imuAngvel[1].(float64),
						imuAngvel[2].(float64),
					},
					Timestamp: time.Now(),
				}

				// Prediction step
				ekf.Predict(imuMeasurement)

				// Update step if radar available
				if msg.Radar != nil {
					radarData := msg.Radar
					if altitude, ok := radarData["altitude"].(float64); ok {
						radarMeasurement := estimation.RadarMeasurement{
							Altitude:  altitude,
							Timestamp: time.Now(),
						}
						ekf.Update(radarMeasurement)
					}
				}

				// Publish state
				state := ekf.GetState()
				uncertainty := ekf.GetUncertainty()
				broadcaster.PublishState(state, uncertainty)

				if *verbose {
					log.Printf("[T=%.2fs] Alt=%.1fm, Vel=%.1f m/s, Clients=%d, Buffer=%d",
						msg.Timestamp, state.Position[2],
						-state.Velocity[2], // Negative because z-velocity is negative
						broadcaster.GetClientCount(),
						consumer.BufferSize())
				}
			}

		case <-sigChan:
			log.Println("Shutdown signal received")
			goto cleanup
		}
	}

cleanup:
	log.Println("Shutting down fusion engine...")
	consumer.Stop()
	broadcaster.Stop()
	log.Println("Shutdown complete")
}
