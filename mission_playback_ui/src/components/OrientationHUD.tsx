import React, { useEffect, useState } from 'react';
import { TelemetryState } from '../hooks/useTelemetry';

interface OrientationHUDProps {
  state: TelemetryState | null;
}

/**
 * Orientation HUD Component
 * Displays navigation ball / artificial horizon
 */
export const OrientationHUD: React.FC<OrientationHUDProps> = ({ state }) => {
  const [pitch, setPitch] = useState(0);
  const [roll, setRoll] = useState(0);
  const [yaw, setYaw] = useState(0);

  useEffect(() => {
    if (state) {
      // Convert quaternion to Euler angles
      const q = state.quaternion;
      const [qx, qy, qz, qw] = q;

      // Roll (x-axis rotation)
      const roll = Math.atan2(2 * (qw * qx + qy * qz), 1 - 2 * (qx * qx + qy * qy));
      // Pitch (y-axis rotation)
      const pitch = Math.asin(2 * (qw * qy - qz * qx));
      // Yaw (z-axis rotation)
      const yaw = Math.atan2(2 * (qw * qz + qx * qy), 1 - 2 * (qy * qy + qz * qz));

      setPitch((pitch * 180) / Math.PI);
      setRoll((roll * 180) / Math.PI);
      setYaw((yaw * 180) / Math.PI);
    }
  }, [state]);

  return (
    <div style={{ padding: '20px', backgroundColor: '#1a1a1a', color: '#00ff00', fontFamily: 'monospace' }}>
      <h3>Orientation (Artificial Horizon)</h3>
      <svg width="300" height="300" viewBox="0 0 300 300" style={{ border: '2px solid #00ff00' }}>
        {/* Sky */}
        <rect x="0" y="0" width="300" height="150" fill="#0033aa" />
        {/* Ground */}
        <rect x="0" y="150" width="300" height="150" fill="#8b6914" />

        {/* Horizon line (rotated by roll) */}
        <g transform={`translate(150, 150) rotate(${roll})`}>
          <line x1="-150" y1="0" x2="150" y2="0" stroke="#00ff00" strokeWidth="2" />
        </g>

        {/* Pitch lines */}
        <g transform={`translate(150, 150) rotate(${roll})`}>
          {[-60, -30, 0, 30, 60].map((p) => (
            <g key={p}>
              <line x1="-20" y1={-p * 2.5} x2="20" y2={-p * 2.5} stroke="#00ff00" strokeWidth="1" />
              <text x="30" y={-p * 2.5 + 4} fill="#00ff00" fontSize="12">
                {p}°
              </text>
            </g>
          ))}
        </g>

        {/* Center reticle */}
        <circle cx="150" cy="150" r="40" fill="none" stroke="#00ff00" strokeWidth="2" />
        <line x1="110" y1="150" x2="190" y2="150" stroke="#00ff00" strokeWidth="2" />
        <line x1="150" y1="110" x2="150" y2="190" stroke="#00ff00" strokeWidth="2" />
      </svg>

      <div style={{ marginTop: '20px' }}>
        <div>Pitch: {pitch.toFixed(1)}°</div>
        <div>Roll: {roll.toFixed(1)}°</div>
        <div>Yaw: {yaw.toFixed(1)}°</div>
      </div>
    </div>
  );
};
