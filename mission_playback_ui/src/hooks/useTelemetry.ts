/*
 * useTelemetry Hook
 * Manages WebSocket connection and state buffering
 */

import { useState, useEffect, useRef, useCallback } from 'react';

export interface TelemetryState {
  timestamp: number;
  position: [number, number, number];
  velocity: [number, number, number];
  quaternion: [number, number, number, number];
  uncertainty: [number, number, number, number, number, number, number, number, number, number];
}

interface UseTelemetryReturn {
  state: TelemetryState | null;
  isConnected: boolean;
  error: string | null;
  history: TelemetryState[];
}

export const useTelemetry = (wsUrl: string = 'ws://localhost:8080/ws'): UseTelemetryReturn => {
  const [state, setState] = useState<TelemetryState | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<TelemetryState[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const historyRef = useRef<TelemetryState[]>([]);

  useEffect(() => {
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
      setError(null);
    };

    ws.onmessage = (event) => {
      try {
        const telemetry = JSON.parse(event.data) as TelemetryState;
        setState(telemetry);

        // Keep history of last 1000 measurements
        historyRef.current.push(telemetry);
        if (historyRef.current.length > 1000) {
          historyRef.current.shift();
        }
        setHistory([...historyRef.current]);
      } catch (err) {
        console.error('Failed to parse telemetry:', err);
        setError('Failed to parse telemetry data');
      }
    };

    ws.onerror = () => {
      setError('WebSocket connection error');
      setIsConnected(false);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
      // Attempt reconnect after 3 seconds
      setTimeout(() => {
        window.location.reload();
      }, 3000);
    };

    wsRef.current = ws;

    return () => {
      ws.close();
    };
  }, [wsUrl]);

  return {
    state,
    isConnected,
    error,
    history,
  };
};
