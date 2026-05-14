import React from 'react';
import './App.css';
import { useTelemetry } from './hooks/useTelemetry';
import { Descent3D } from './components/Descent3D';
import { AltitudeVelocityChart } from './components/AltitudeVelocityChart';
import { OrientationHUD } from './components/OrientationHUD';
import { TimelineScrubber } from './components/TimelineScrubber';

function App() {
  const { state, isConnected, error, history } = useTelemetry('ws://localhost:8080/ws');

  const handleSeek = (index: number) => {
    // This would implement playback seeking
    console.log('Seeking to index:', index);
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>🚀 EDL Mission Playback Dashboard</h1>
        <div className="connection-status">
          <span className={isConnected ? 'status-connected' : 'status-disconnected'}>
            {isConnected ? '● Connected' : '● Disconnected'}
          </span>
          {error && <span className="error">{error}</span>}
        </div>
      </header>

      <main className="app-main">
        <div className="layout">
          {/* Left panel: 3D visualization */}
          <div className="panel-3d">
            <Descent3D state={state} />
          </div>

          {/* Right panel: Telemetry readout */}
          <div className="panel-right">
            <div className="telemetry-display">
              <h3>Telemetry Readout</h3>
              {state ? (
                <div className="telemetry-values">
                  <div className="value-group">
                    <div className="label">Position</div>
                    <div className="value">
                      X: {state.position[0].toFixed(1)} m
                    </div>
                    <div className="value">
                      Y: {state.position[1].toFixed(1)} m
                    </div>
                    <div className="value">
                      Z: {state.position[2].toFixed(1)} m (Altitude)
                    </div>
                  </div>

                  <div className="value-group">
                    <div className="label">Velocity</div>
                    <div className="value">
                      VX: {state.velocity[0].toFixed(2)} m/s
                    </div>
                    <div className="value">
                      VY: {state.velocity[1].toFixed(2)} m/s
                    </div>
                    <div className="value">
                      VZ: {state.velocity[2].toFixed(2)} m/s
                    </div>
                  </div>

                  <div className="value-group">
                    <div className="label">Uncertainty (1σ)</div>
                    <div className="value">
                      Pos: ±{Math.sqrt(state.uncertainty[0] ** 2 + state.uncertainty[1] ** 2 + state.uncertainty[2] ** 2).toFixed(1)} m
                    </div>
                    <div className="value">
                      Vel: ±{Math.sqrt(state.uncertainty[3] ** 2 + state.uncertainty[4] ** 2 + state.uncertainty[5] ** 2).toFixed(2)} m/s
                    </div>
                  </div>
                </div>
              ) : (
                <div className="no-data">Waiting for telemetry...</div>
              )}
            </div>

            <OrientationHUD state={state} />
          </div>
        </div>

        {/* Bottom panel: Charts */}
        <div className="panel-bottom">
          <AltitudeVelocityChart history={history} />
        </div>
      </main>

      {/* Timeline */}
      <TimelineScrubber history={history} onSeek={handleSeek} />
    </div>
  );
}

export default App;
