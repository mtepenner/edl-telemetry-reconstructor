"""
Descent Profile Module
Simulates true physics: aerodynamic drag, gravity, and thrust during EDL sequence.
"""
import numpy as np
from dataclasses import dataclass
from typing import Tuple


@dataclass
class DescentState:
    """Represents the spacecraft's state during descent"""
    time: float
    position: np.ndarray  # [x, y, z] in meters
    velocity: np.ndarray  # [vx, vy, vz] in m/s
    acceleration: np.ndarray  # [ax, ay, az] in m/s²
    quaternion: np.ndarray  # [qx, qy, qz, qw] orientation
    angular_velocity: np.ndarray  # [wx, wy, wz] in rad/s


class DescentProfile:
    """Simulates Mars EDL descent dynamics"""

    def __init__(self):
        """Initialize descent profile parameters"""
        # Mars atmosphere and gravity parameters
        self.mars_gravity = 3.71  # m/s²
        self.mars_atmosphere_density_sea_level = 0.020  # kg/m³
        self.scale_height = 11500  # meters (Mars atmosphere scale height)
        
        # Spacecraft parameters
        self.mass = 2500  # kg (typical Mars rover + lander)
        self.drag_coefficient = 1.2  # for blunt body
        self.reference_area = 10.0  # m² (heat shield area)
        
        # Initial conditions (typical for Mars entry)
        self.altitude_initial = 120000  # meters
        self.velocity_initial = 5800  # m/s (entry velocity)
        
        # Thrust parameters
        self.max_thrust = 50000  # Newtons (retro-rockets)
        self.thrust_start_altitude = 15000  # meters
        self.thrust_end_altitude = 500  # meters

    def atmosphere_density(self, altitude: float) -> float:
        """
        Calculate Mars atmospheric density at given altitude.
        Uses exponential model.
        
        Args:
            altitude: altitude in meters
            
        Returns:
            density in kg/m³
        """
        if altitude < 0:
            altitude = 0
        return self.mars_atmosphere_density_sea_level * np.exp(-altitude / self.scale_height)

    def calculate_drag_force(self, velocity_magnitude: float, altitude: float) -> float:
        """
        Calculate drag force magnitude.
        F_drag = 0.5 * rho * v^2 * Cd * A
        
        Args:
            velocity_magnitude: speed in m/s
            altitude: altitude in meters
            
        Returns:
            drag force in Newtons
        """
        rho = self.atmosphere_density(altitude)
        drag = 0.5 * rho * velocity_magnitude**2 * self.drag_coefficient * self.reference_area
        return drag

    def calculate_thrust_vector(self, altitude: float, velocity: np.ndarray) -> np.ndarray:
        """
        Calculate thrust vector (opposing velocity direction).
        Thrust is active only between specific altitude bands.
        
        Args:
            altitude: current altitude in meters
            velocity: velocity vector [vx, vy, vz]
            
        Returns:
            thrust vector [fx, fy, fz] in Newtons
        """
        if altitude > self.thrust_start_altitude or altitude < self.thrust_end_altitude:
            return np.array([0.0, 0.0, 0.0])
        
        velocity_magnitude = np.linalg.norm(velocity)
        if velocity_magnitude < 1e-6:
            return np.array([0.0, 0.0, 0.0])
        
        # Thrust opposes velocity direction
        thrust_unit = -velocity / velocity_magnitude
        thrust = self.max_thrust * thrust_unit
        return thrust

    def step(self, state: DescentState, dt: float) -> DescentState:
        """
        Simulate one time step of descent using simple Euler integration.
        
        Args:
            state: current descent state
            dt: time step in seconds
            
        Returns:
            new descent state
        """
        # Current altitude
        altitude = state.position[2]
        velocity_magnitude = np.linalg.norm(state.velocity)
        
        # Forces
        # Gravity (pointing down, z-negative on Mars)
        gravity_force = np.array([0.0, 0.0, -self.mars_gravity * self.mass])
        
        # Drag force (opposing velocity)
        if velocity_magnitude > 1e-6:
            drag_force = -0.5 * self.atmosphere_density(altitude) * \
                        velocity_magnitude**2 * self.drag_coefficient * \
                        self.reference_area * (state.velocity / velocity_magnitude)
        else:
            drag_force = np.array([0.0, 0.0, 0.0])
        
        # Thrust force
        thrust_force = self.calculate_thrust_vector(altitude, state.velocity)
        
        # Total force and acceleration
        total_force = gravity_force + drag_force + thrust_force
        acceleration = total_force / self.mass
        
        # Update state (Euler integration)
        new_position = state.position + state.velocity * dt
        new_velocity = state.velocity + acceleration * dt
        
        # Ensure we don't go below ground (altitude = 0)
        if new_position[2] < 0:
            new_position[2] = 0
            new_velocity[2] = 0  # Stop downward motion
        
        # Simple attitude dynamics (slowly stabilizing to vertical)
        # In real scenario, this would use quaternion integration
        new_quaternion = state.quaternion
        new_angular_velocity = state.angular_velocity * 0.95  # Damping
        
        return DescentState(
            time=state.time + dt,
            position=new_position,
            velocity=new_velocity,
            acceleration=acceleration,
            quaternion=new_quaternion,
            angular_velocity=new_angular_velocity
        )

    def get_initial_state(self) -> DescentState:
        """Get initial state for descent simulation"""
        return DescentState(
            time=0.0,
            position=np.array([0.0, 0.0, self.altitude_initial]),
            velocity=np.array([0.0, 0.0, -self.velocity_initial]),
            acceleration=np.array([0.0, 0.0, 0.0]),
            quaternion=np.array([0.0, 0.0, 0.0, 1.0]),  # Identity quaternion
            angular_velocity=np.array([0.0, 0.0, 0.0])
        )
