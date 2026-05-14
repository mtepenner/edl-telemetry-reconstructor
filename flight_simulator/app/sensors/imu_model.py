"""
IMU (Inertial Measurement Unit) Model
Simulates accelerometer and gyroscope with realistic noise, bias, and random walk.
"""
import numpy as np
from dataclasses import dataclass


@dataclass
class IMUMeasurement:
    """IMU sensor measurement"""
    timestamp: float
    acceleration: np.ndarray  # [ax, ay, az] in m/s²
    angular_velocity: np.ndarray  # [wx, wy, wz] in rad/s


class IMUModel:
    """Simulates IMU with noise and biases"""

    def __init__(self, random_seed: int = None):
        """
        Initialize IMU model with noise parameters.
        
        Args:
            random_seed: for reproducible randomness
        """
        if random_seed is not None:
            np.random.seed(random_seed)
        
        # Accelerometer noise parameters
        self.accel_noise_std = 0.01  # m/s² (Gaussian white noise)
        self.accel_bias_instability = 0.001  # m/s² (bias random walk)
        self.accel_initial_bias = np.random.normal(0, 0.05, 3)  # Initial bias
        
        # Gyroscope noise parameters
        self.gyro_noise_std = 0.001  # rad/s (Gaussian white noise)
        self.gyro_bias_instability = 0.0001  # rad/s (bias random walk)
        self.gyro_initial_bias = np.random.normal(0, 0.01, 3)  # Initial bias
        
        # Current biases (evolve over time)
        self.accel_bias = self.accel_initial_bias.copy()
        self.gyro_bias = self.gyro_initial_bias.copy()

    def measure(self, true_acceleration: np.ndarray, 
                true_angular_velocity: np.ndarray,
                timestamp: float) -> IMUMeasurement:
        """
        Generate noisy IMU measurement from true values.
        
        Adds:
        - Gaussian white noise
        - Bias (constant + random walk component)
        - Potential scale factor errors (not implemented here)
        
        Args:
            true_acceleration: true acceleration [ax, ay, az]
            true_angular_velocity: true angular velocity [wx, wy, wz]
            timestamp: current timestamp
            
        Returns:
            IMUMeasurement with noisy data
        """
        # Update biases (random walk)
        self.accel_bias += np.random.normal(0, self.accel_bias_instability, 3)
        self.gyro_bias += np.random.normal(0, self.gyro_bias_instability, 3)
        
        # Add Gaussian white noise
        accel_noise = np.random.normal(0, self.accel_noise_std, 3)
        gyro_noise = np.random.normal(0, self.gyro_noise_std, 3)
        
        # Noisy measurements = true value + bias + noise
        noisy_acceleration = true_acceleration + self.accel_bias + accel_noise
        noisy_angular_velocity = true_angular_velocity + self.gyro_bias + gyro_noise
        
        return IMUMeasurement(
            timestamp=timestamp,
            acceleration=noisy_acceleration,
            angular_velocity=noisy_angular_velocity
        )

    def reset_biases(self):
        """Reset biases to initial values (for simulation restarts)"""
        self.accel_bias = self.accel_initial_bias.copy()
        self.gyro_bias = self.gyro_initial_bias.copy()
