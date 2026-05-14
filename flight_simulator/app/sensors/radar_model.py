"""
Radar Altimeter Model
Simulates altitude measurements with noise based on Martian terrain.
"""
import numpy as np
from dataclasses import dataclass


@dataclass
class RadarMeasurement:
    """Radar altimeter measurement"""
    timestamp: float
    altitude: float  # meters (altitude above ground)
    signal_strength: float  # 0.0 to 1.0


class RadarModel:
    """Simulates radar altimeter with realistic noise"""

    def __init__(self, random_seed: int = None):
        """
        Initialize radar model.
        
        Args:
            random_seed: for reproducible randomness
        """
        if random_seed is not None:
            np.random.seed(random_seed)
        
        # Radar noise characteristics
        self.altitude_noise_std_m = 0.5  # Base noise in meters
        self.altitude_noise_std_percent = 0.01  # 1% of measured altitude
        
        # Measurement update rate
        self.update_rate_hz = 10.0  # 10 Hz update rate
        self.last_update_time = -1.0 / self.update_rate_hz
        
        # Martian terrain model (simple)
        self.terrain_roughness = 2.0  # meters RMS roughness
        self.terrain_frequency = 0.001  # wavelength parameter

    def measure(self, true_altitude: float, timestamp: float) -> RadarMeasurement:
        """
        Generate noisy radar measurement.
        
        Simulates:
        - Range measurement noise (proportional to altitude)
        - Terrain roughness effects
        - Update rate constraints
        
        Args:
            true_altitude: true altitude above ground in meters
            timestamp: current timestamp
            
        Returns:
            RadarMeasurement with noisy altitude
        """
        # Check if we should generate a new measurement
        # (simulate discrete update rate)
        time_since_update = timestamp - self.last_update_time
        if time_since_update < (1.0 / self.update_rate_hz):
            # Return last measurement or skip
            return None
        
        self.last_update_time = timestamp
        
        # Terrain roughness model
        terrain_noise = np.random.normal(0, self.terrain_roughness)
        
        # Range measurement noise (increases with altitude)
        altitude_measurement_noise = np.random.normal(
            0, 
            max(self.altitude_noise_std_m, true_altitude * self.altitude_noise_std_percent)
        )
        
        # Combined noise
        total_noise = terrain_noise + altitude_measurement_noise
        
        # Noisy measurement
        measured_altitude = true_altitude + total_noise
        
        # Radar can't measure negative altitudes
        if measured_altitude < 0:
            measured_altitude = 0.0
        
        # Signal strength decreases with altitude (realistic for radar)
        max_range = 10000.0  # meters
        signal_strength = max(0.1, 1.0 - (true_altitude / max_range))
        
        return RadarMeasurement(
            timestamp=timestamp,
            altitude=measured_altitude,
            signal_strength=signal_strength
        )
