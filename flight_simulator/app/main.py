"""
Flight Simulator Main Entry Point
Generates realistic noisy sensor data from spacecraft descent simulation.
Streams data via UDP or Kafka.
"""
import json
import socket
import time
import argparse
import logging
from typing import Optional
import numpy as np

from dynamics.descent_profile import DescentProfile, DescentState
from sensors.imu_model import IMUModel
from sensors.radar_model import RadarModel

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class FlightSimulator:
    """Simulates spacecraft descent and generates sensor telemetry"""

    def __init__(self, use_kafka: bool = False, kafka_broker: str = None, 
                 udp_host: str = "127.0.0.1", udp_port: int = 5005):
        """
        Initialize the flight simulator.
        
        Args:
            use_kafka: whether to use Kafka (True) or UDP (False)
            kafka_broker: Kafka broker address (if using Kafka)
            udp_host: UDP host address
            udp_port: UDP port
        """
        self.descent_profile = DescentProfile()
        self.imu_model = IMUModel(random_seed=42)
        self.radar_model = RadarModel(random_seed=43)
        
        self.use_kafka = use_kafka
        self.kafka_broker = kafka_broker
        self.udp_host = udp_host
        self.udp_port = udp_port
        
        self.udp_socket: Optional[socket.socket] = None
        self.kafka_producer = None
        
        # Simulation parameters
        self.dt = 0.01  # 10 ms time step (100 Hz)
        self.sim_time = 0.0
        self.target_realtime_factor = 1.0  # Run at real-time speed
        
        self._initialize_transport()

    def _initialize_transport(self):
        """Initialize UDP socket or Kafka producer"""
        if self.use_kafka:
            try:
                from kafka import KafkaProducer
                self.kafka_producer = KafkaProducer(
                    bootstrap_servers=self.kafka_broker,
                    value_serializer=lambda v: json.dumps(v).encode('utf-8')
                )
                logger.info(f"Connected to Kafka broker at {self.kafka_broker}")
            except Exception as e:
                logger.error(f"Failed to connect to Kafka: {e}")
                logger.info("Falling back to UDP")
                self.use_kafka = False
        
        if not self.use_kafka:
            self.udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            logger.info(f"UDP socket initialized, will send to {self.udp_host}:{self.udp_port}")

    def _send_telemetry(self, data: dict):
        """
        Send telemetry data to transport (UDP or Kafka)
        
        Args:
            data: telemetry data as dictionary
        """
        try:
            if self.use_kafka and self.kafka_producer:
                self.kafka_producer.send('telemetry', value=data)
            elif self.udp_socket:
                message = json.dumps(data).encode('utf-8')
                self.udp_socket.sendto(message, (self.udp_host, self.udp_port))
        except Exception as e:
            logger.error(f"Error sending telemetry: {e}")

    def _package_telemetry(self, state: DescentState, imu_data, radar_data) -> dict:
        """
        Package current state and measurements into telemetry message.
        
        Args:
            state: current descent state
            imu_data: IMU measurement
            radar_data: radar measurement
            
        Returns:
            dictionary with telemetry data
        """
        altitude = max(0, state.position[2])  # Prevent negative altitude
        
        telemetry = {
            "timestamp": state.time,
            "sim_time": self.sim_time,
            "true_state": {
                "position": state.position.tolist(),
                "velocity": state.velocity.tolist(),
                "acceleration": state.acceleration.tolist(),
                "quaternion": state.quaternion.tolist(),
                "altitude": altitude
            },
            "imu": {
                "timestamp": imu_data.timestamp,
                "acceleration": imu_data.acceleration.tolist(),
                "angular_velocity": imu_data.angular_velocity.tolist()
            }
        }
        
        if radar_data:
            telemetry["radar"] = {
                "timestamp": radar_data.timestamp,
                "altitude": radar_data.altitude,
                "signal_strength": radar_data.signal_strength
            }
        
        return telemetry

    def run(self, duration: float = 600.0, verbose: bool = False):
        """
        Run the flight simulation.
        
        Args:
            duration: maximum simulation time in seconds
            verbose: print detailed state information
        """
        logger.info(f"Starting descent simulation for {duration} seconds")
        
        state = self.descent_profile.get_initial_state()
        iteration = 0
        last_wall_time = time.time()
        
        try:
            while state.time < duration and state.position[2] > 0:
                # Get measurements
                imu_data = self.imu_model.measure(
                    state.acceleration,
                    state.angular_velocity,
                    state.time
                )
                
                radar_data = self.radar_model.measure(
                    max(0, state.position[2]),
                    state.time
                )
                
                # Package and send telemetry
                telemetry = self._package_telemetry(state, imu_data, radar_data)
                self._send_telemetry(telemetry)
                
                # Logging
                if verbose and iteration % 100 == 0:
                    logger.info(
                        f"T={state.time:7.2f}s | Alt={state.position[2]:8.1f}m | "
                        f"Vel={np.linalg.norm(state.velocity):7.1f}m/s | "
                        f"Accel={np.linalg.norm(state.acceleration):6.2f}m/s²"
                    )
                
                # Simulation step
                state = self.descent_profile.step(state, self.dt)
                self.sim_time += self.dt
                
                # Real-time synchronization
                elapsed_wall_time = time.time() - last_wall_time
                desired_wall_time = self.dt * self.target_realtime_factor
                
                if elapsed_wall_time < desired_wall_time:
                    time.sleep(desired_wall_time - elapsed_wall_time)
                
                last_wall_time = time.time()
                iteration += 1
            
            logger.info(f"Descent complete at T={state.time:.2f}s, Final altitude={state.position[2]:.1f}m")
            
        except KeyboardInterrupt:
            logger.info("Simulation interrupted by user")
        finally:
            self._cleanup()

    def _cleanup(self):
        """Clean up resources"""
        if self.udp_socket:
            self.udp_socket.close()
        if self.kafka_producer:
            self.kafka_producer.close()
        logger.info("Simulator shutdown complete")


def main():
    parser = argparse.ArgumentParser(description='EDL Telemetry Flight Simulator')
    parser.add_argument('--kafka', action='store_true', help='Use Kafka instead of UDP')
    parser.add_argument('--kafka-broker', default='localhost:9092', 
                       help='Kafka broker address (default: localhost:9092)')
    parser.add_argument('--udp-host', default='127.0.0.1', help='UDP host')
    parser.add_argument('--udp-port', type=int, default=5005, help='UDP port')
    parser.add_argument('--duration', type=float, default=600.0, help='Simulation duration (seconds)')
    parser.add_argument('--verbose', action='store_true', help='Verbose output')
    
    args = parser.parse_args()
    
    simulator = FlightSimulator(
        use_kafka=args.kafka,
        kafka_broker=args.kafka_broker,
        udp_host=args.udp_host,
        udp_port=args.udp_port
    )
    
    simulator.run(duration=args.duration, verbose=args.verbose)


if __name__ == '__main__':
    main()
