import React, { useState } from 'react';
import { TelemetryState } from '../hooks/useTelemetry';

interface TimelineScrubberProps {
  history: TelemetryState[];
  onSeek: (index: number) => void;
}

/**
 * Timeline Scrubber Component
 * Allows rewind and replay of specific descent events
 */
export const TimelineScrubber: React.FC<TimelineScrubberProps> = ({ history, onSeek }) => {
  const [position, setPosition] = useState(0);
  const [isPlaying, setIsPlaying] = useState(true);

  const handleSliderChange = (value: number) => {
    setPosition(value);
    onSeek(value);
  };

  return (
    <div style={{ padding: '20px', backgroundColor: '#f5f5f5', borderTop: '1px solid #ccc' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '10px' }}>
        <button onClick={() => setIsPlaying(!isPlaying)}>
          {isPlaying ? '⏸ Pause' : '▶ Play'}
        </button>
        <span>Position: {position} / {history.length}</span>
      </div>
      <input
        type="range"
        min="0"
        max={Math.max(1, history.length - 1)}
        value={position}
        onChange={(e) => handleSliderChange(parseInt(e.target.value))}
        style={{ width: '100%' }}
      />
      <div style={{ marginTop: '10px', fontSize: '12px', color: '#666' }}>
        {history.length > 0 && (
          <>
            <div>Duration: {(history.length * 0.01).toFixed(1)}s</div>
            <div>Sample Rate: 100 Hz</div>
          </>
        )}
      </div>
    </div>
  );
};
