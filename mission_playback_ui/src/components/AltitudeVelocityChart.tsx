import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { TelemetryState } from '../hooks/useTelemetry';

interface AltitudeVelocityChartProps {
  history: TelemetryState[];
}

/**
 * Altitude vs Velocity Chart Component
 * Compares raw noisy sensor data against Kalman-filtered truth
 */
export const AltitudeVelocityChart: React.FC<AltitudeVelocityChartProps> = ({ history }) => {
  // Prepare data for Recharts
  const chartData = history.map((state, index) => ({
    time: index,
    altitude: state.position[2],
    velocity: Math.sqrt(
      state.velocity[0] ** 2 + state.velocity[1] ** 2 + state.velocity[2] ** 2
    ),
    altitudeUncertainty: state.uncertainty[2],
    velocityUncertainty: state.uncertainty[3],
  }));

  return (
    <div style={{ width: '100%', height: '300px', backgroundColor: '#f5f5f5', padding: '10px' }}>
      <h3>Altitude & Velocity Profile</h3>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="time" />
          <YAxis yAxisId="left" label={{ value: 'Altitude (m)', angle: -90, position: 'insideLeft' }} />
          <YAxis yAxisId="right" orientation="right" label={{ value: 'Velocity (m/s)', angle: 90, position: 'insideRight' }} />
          <Tooltip />
          <Legend />
          <Line yAxisId="left" type="monotone" dataKey="altitude" stroke="#0088ff" dot={false} />
          <Line yAxisId="right" type="monotone" dataKey="velocity" stroke="#ff6600" dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};
