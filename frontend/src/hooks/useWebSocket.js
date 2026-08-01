import { useState, useEffect, useRef, useCallback } from 'react';

export function useWebSocket(url) {
  const [status, setStatus] = useState('disconnected');
  const socketRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const messageListenersRef = useRef(new Set());

  const connect = useCallback(() => {
    if (socketRef.current && (socketRef.current.readyState === WebSocket.CONNECTING || socketRef.current.readyState === WebSocket.OPEN)) {
      return;
    }

    setStatus('connecting');
    const ws = new WebSocket(url);
    socketRef.current = ws;

    ws.onopen = () => {
      setStatus('connected');
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };

    ws.onmessage = (event) => {
      try {
        const lines = event.data.split('\n');
        for (const line of lines) {
          const trimmed = line.trim();
          if (trimmed) {
            const data = JSON.parse(trimmed);
            messageListenersRef.current.forEach((listener) => listener(data));
          }
        }
      } catch (err) {
        console.error('Error parsing WebSocket message:', err);
      }
    };

    ws.onclose = () => {
      setStatus('disconnected');
      socketRef.current = null;
      // Attempt reconnection after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, 3000);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      ws.close();
    };
  }, [url]);

  const send = useCallback((type, payload) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ type, payload }));
      return true;
    }
    return false;
  }, []);

  const addListener = useCallback((listener) => {
    messageListenersRef.current.add(listener);
    return () => {
      messageListenersRef.current.delete(listener);
    };
  }, []);

  useEffect(() => {
    connect();
    return () => {
      if (socketRef.current) {
        socketRef.current.onclose = null; // Prevent reconnect on unmount
        socketRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [connect]);

  return { status, send, addListener };
}
