/*
Extended Kalman Filter (EKF) Implementation
Fuses IMU and Radar measurements for state estimation during descent.

The EKF predicts state using IMU (acceleration, angular velocity) and
updates the estimate using Radar (altitude) measurements.
*/

package estimation

import (
	"math"
	"time"
)

// State represents the spacecraft state vector
// [x, y, z, vx, vy, vz, qx, qy, qz, qw]
type State struct {
	Position   [3]float64 // x, y, z in meters
	Velocity   [3]float64 // vx, vy, vz in m/s
	Quaternion [4]float64 // qx, qy, qz, qw (orientation)
	Timestamp  time.Time
}

// IMUMeasurement is inertial measurement data
type IMUMeasurement struct {
	Acceleration    [3]float64 // ax, ay, az in m/s²
	AngularVelocity [3]float64 // wx, wy, wz in rad/s
	Timestamp       time.Time
}

// RadarMeasurement is altitude measurement data
type RadarMeasurement struct {
	Altitude       float64 // altitude in meters
	SignalStrength float64 // 0.0 to 1.0
	Timestamp      time.Time
}

// ExtendedKalmanFilter for descent state estimation
type ExtendedKalmanFilter struct {
	// State: [px, py, pz, vx, vy, vz, qx, qy, qz, qw] = 10 states
	x *Vector // State estimate (10x1)
	P *Matrix // Covariance matrix (10x10)

	// Process noise covariance (10x10)
	Q *Matrix
	// Measurement noise for radar (1x1)
	R float64

	// Mars gravity
	g float64

	// Last measurement time for dt calculation
	lastTime time.Time
}

// NewExtendedKalmanFilter creates a new EKF instance
func NewExtendedKalmanFilter(initialState State) *ExtendedKalmanFilter {
	ekf := &ExtendedKalmanFilter{
		g:        3.71, // Mars gravity
		lastTime: time.Now(),
	}

	// Initialize state vector [px, py, pz, vx, vy, vz, qx, qy, qz, qw]
	stateData := []float64{
		initialState.Position[0],
		initialState.Position[1],
		initialState.Position[2],
		initialState.Velocity[0],
		initialState.Velocity[1],
		initialState.Velocity[2],
		initialState.Quaternion[0],
		initialState.Quaternion[1],
		initialState.Quaternion[2],
		initialState.Quaternion[3],
	}
	ekf.x = NewVectorFromSlice(stateData)

	// Initialize covariance matrix (relatively high initial uncertainty)
	ekf.P = NewMatrix(10, 10)
	for i := 0; i < 10; i++ {
		if i < 3 {
			ekf.P.Set(i, i, 100.0) // Position uncertainty: 10m
		} else if i < 6 {
			ekf.P.Set(i, i, 100.0) // Velocity uncertainty: 10 m/s
		} else {
			ekf.P.Set(i, i, 0.1) // Quaternion uncertainty
		}
	}

	// Process noise covariance
	ekf.Q = NewMatrix(10, 10)
	for i := 0; i < 10; i++ {
		if i < 3 {
			ekf.Q.Set(i, i, 0.01) // Position process noise
		} else if i < 6 {
			ekf.Q.Set(i, i, 0.1) // Velocity process noise
		} else {
			ekf.Q.Set(i, i, 0.001) // Quaternion process noise
		}
	}

	// Measurement noise (radar altitude)
	ekf.R = 0.25 // 0.5m standard deviation

	return ekf
}

// Predict performs the EKF prediction step using IMU data
func (ekf *ExtendedKalmanFilter) Predict(imu IMUMeasurement) {
	now := time.Now()
	dt := now.Sub(ekf.lastTime).Seconds()
	if dt <= 0 || dt > 0.1 { // Skip if dt invalid or too large
		dt = 0.01
	}
	ekf.lastTime = now

	// Extract state
	px, py, pz := ekf.x.At(0, 0), ekf.x.At(1, 0), ekf.x.At(2, 0)
	vx, vy, vz := ekf.x.At(3, 0), ekf.x.At(4, 0), ekf.x.At(5, 0)
	qx, qy, qz, qw := ekf.x.At(6, 0), ekf.x.At(7, 0), ekf.x.At(8, 0), ekf.x.At(9, 0)

	ax := imu.Acceleration[0]
	ay := imu.Acceleration[1]
	az := imu.Acceleration[2]

	// Predict new state using simple kinematic model
	// Position: x_new = x + v*dt + 0.5*a*dt²
	newPx := px + vx*dt + 0.5*ax*dt*dt
	newPy := py + vy*dt + 0.5*ay*dt*dt
	newPz := pz + vz*dt + 0.5*az*dt*dt

	// Velocity: v_new = v + a*dt - g*dt (gravity acts downward)
	newVx := vx + ax*dt
	newVy := vy + ay*dt
	newVz := vz + (az-ekf.g)*dt

	// Quaternion: simple update (in practice, would use proper quaternion integration)
	// For now, keep quaternion fixed
	newQx, newQy, newQz, newQw := qx, qy, qz, qw

	// Prevent altitude from going negative
	if newPz < 0 {
		newPz = 0
		if newVz < 0 {
			newVz = 0
		}
	}

	// Update state vector
	ekf.x.SetVec(0, newPx)
	ekf.x.SetVec(1, newPy)
	ekf.x.SetVec(2, newPz)
	ekf.x.SetVec(3, newVx)
	ekf.x.SetVec(4, newVy)
	ekf.x.SetVec(5, newVz)
	ekf.x.SetVec(6, newQx)
	ekf.x.SetVec(7, newQy)
	ekf.x.SetVec(8, newQz)
	ekf.x.SetVec(9, newQw)

	// Predict covariance: P_new = A*P*A^T + Q
	// Using simplified linear approximation
	A := ekf.computeStateTransitionMatrix(dt)
	AT := Transpose(A)
	APAT := Multiply(Multiply(A, ekf.P), AT)
	ekf.P = Add(APAT, ekf.Q)
}

// Update performs the EKF update step using radar measurement
func (ekf *ExtendedKalmanFilter) Update(radar RadarMeasurement) {
	// Observation matrix H: we only measure altitude (z position)
	// H = [0, 0, 1, 0, 0, 0, 0, 0, 0, 0]
	H := NewMatrix(1, 10)
	H.Set(0, 2, 1.0)

	// Innovation: z - h(x)
	// Measured altitude - predicted altitude
	z := radar.Altitude
	zPred := ekf.x.At(2, 0)
	innovation := z - zPred

	// Innovation covariance: S = H*P*H^T + R
	HT := Transpose(H)
	PHT := Multiply(ekf.P, HT)
	HPHT := Multiply(H, PHT)
	S := HPHT.At(0, 0) + ekf.R

	// Kalman gain: K = P*H^T / S
	K := NewMatrix(10, 1)
	for i := 0; i < 10; i++ {
		K.Set(i, 0, PHT.At(i, 0)/S)
	}

	// Update state: x = x + K * innovation
	for i := 0; i < 10; i++ {
		ekf.x.SetVec(i, ekf.x.At(i, 0)+K.At(i, 0)*innovation)
	}

	// Update covariance: P = (I - K*H)*P
	I := Identity(10)
	KH := Multiply(K, H)
	IKH := Subtract(I, KH)
	ekf.P = Multiply(IKH, ekf.P)

	// Ensure altitude doesn't go negative
	if ekf.x.At(2, 0) < 0 {
		ekf.x.SetVec(2, 0)
		if ekf.x.At(5, 0) < 0 {
			ekf.x.SetVec(5, 0)
		}
	}
}

// GetState returns the current estimated state
func (ekf *ExtendedKalmanFilter) GetState() State {
	return State{
		Position: [3]float64{
			ekf.x.At(0, 0),
			ekf.x.At(1, 0),
			ekf.x.At(2, 0),
		},
		Velocity: [3]float64{
			ekf.x.At(3, 0),
			ekf.x.At(4, 0),
			ekf.x.At(5, 0),
		},
		Quaternion: [4]float64{
			ekf.x.At(6, 0),
			ekf.x.At(7, 0),
			ekf.x.At(8, 0),
			ekf.x.At(9, 0),
		},
		Timestamp: ekf.lastTime,
	}
}

// computeStateTransitionMatrix creates the Jacobian for state prediction
func (ekf *ExtendedKalmanFilter) computeStateTransitionMatrix(dt float64) *Matrix {
	A := Identity(10)

	// Position depends on velocity
	A.Set(0, 3, dt)
	A.Set(1, 4, dt)
	A.Set(2, 5, dt)

	// Velocity depends on acceleration (dt dependency)
	// This is simplified; full version would use IMU acceleration

	return A
}

// GetCovariance returns the state covariance matrix
func (ekf *ExtendedKalmanFilter) GetCovariance() *Matrix {
	return ekf.P
}

// GetUncertainty returns the uncertainty (standard deviation) for each state variable
func (ekf *ExtendedKalmanFilter) GetUncertainty() [10]float64 {
	var uncertainty [10]float64
	for i := 0; i < 10; i++ {
		uncertainty[i] = math.Sqrt(ekf.P.At(i, i))
	}
	return uncertainty
}
