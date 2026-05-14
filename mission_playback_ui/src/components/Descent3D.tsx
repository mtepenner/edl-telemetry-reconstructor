import React, { useRef, useEffect, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, PerspectiveCamera, Grid, Line } from '@react-three/drei';
import { Group, Euler } from 'three';
import { TelemetryState } from '../hooks/useTelemetry';

interface Descent3DProps {
  state: TelemetryState | null;
}

/**
 * Spacecraft 3D model
 */
const Spacecraft: React.FC<{ position: [number, number, number]; quaternion: [number, number, number, number] }> = ({
  position,
  quaternion,
}) => {
  const groupRef = useRef<Group>(null);

  useEffect(() => {
    if (groupRef.current) {
      groupRef.current.position.set(position[0] / 1000, -position[2] / 1000, position[1] / 1000);

      // Convert quaternion to Euler angles for Three.js
      const q = quaternion;
      const euler = new Euler();
      euler.setFromQuaternion({
        x: q[0],
        y: q[1],
        z: q[2],
        w: q[3],
      } as any);
      groupRef.current.rotation.copy(euler);
    }
  }, [position, quaternion]);

  return (
    <group ref={groupRef}>
      {/* Heat shield (cone) */}
      <mesh>
        <coneGeometry args={[0.3, 0.5, 32]} />
        <meshStandardMaterial color="#ff6600" />
      </mesh>

      {/* Lander body (box) */}
      <mesh position={[0, 0.3, 0]}>
        <boxGeometry args={[0.2, 0.3, 0.2]} />
        <meshStandardMaterial color="#cccccc" />
      </mesh>

      {/* Solar panels (thin boxes) */}
      <mesh position={[-0.25, 0.3, 0]}>
        <boxGeometry args={[0.1, 0.05, 0.3]} />
        <meshStandardMaterial color="#3366ff" />
      </mesh>
      <mesh position={[0.25, 0.3, 0]}>
        <boxGeometry args={[0.1, 0.05, 0.3]} />
        <meshStandardMaterial color="#3366ff" />
      </mesh>
    </group>
  );
};

/**
 * Martian terrain
 */
const MarsTerrain: React.FC = () => {
  return (
    <mesh position={[0, -5, 0]} rotation={[-Math.PI / 2, 0, 0]}>
      <planeGeometry args={[20, 20, 64, 64]} />
      <meshStandardMaterial color="#d4720c" wireframe={false} />
    </mesh>
  );
};

/**
 * Trajectory trail
 */
const TrajectoryTrail: React.FC<{ positions: [number, number, number][] }> = ({ positions }) => {
  const points: [number, number, number][] = positions.map((p) => [p[0] / 1000, -p[2] / 1000, p[1] / 1000]);

  return <Line points={points} color="#00ff00" lineWidth={2} />;
};

/**
 * Descent 3D Visualization Component
 */
const Descent3DContent: React.FC<Descent3DProps> = ({ state }) => {
  const [positions, setPositions] = useState<[number, number, number][]>([]);

  useEffect(() => {
    if (state) {
      setPositions((prev) => [...prev.slice(-500), state.position]);
    }
  }, [state]);

  return (
    <>
      <PerspectiveCamera makeDefault position={[3, 3, 3]} />
      <OrbitControls autoRotate autoRotateSpeed={1} />

      {/* Lighting */}
      <ambientLight intensity={0.6} />
      <directionalLight position={[10, 10, 10]} intensity={1} />

      {/* Scene objects */}
      {state && <Spacecraft position={state.position} quaternion={state.quaternion} />}
      <MarsTerrain />
      {positions.length > 1 && <TrajectoryTrail positions={positions} />}

      {/* Grid */}
      <Grid args={[20, 20]} cellSize={1} />
    </>
  );
};

export const Descent3D: React.FC<Descent3DProps> = (props) => {
  return (
    <div style={{ width: '100%', height: '100%', backgroundColor: '#000000' }}>
      <Canvas>
        <Descent3DContent {...props} />
      </Canvas>
    </div>
  );
};
