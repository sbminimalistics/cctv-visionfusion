// src/services/event-buffer.service.ts
import { EventEmitter } from 'node:events'
import crypto from 'node:crypto'

// Define the shape of the event that will be broadcasted to SSE clients
export interface AppEvent {
  id: string
  payload: any
  timestamp: string
}

class EventBufferService extends EventEmitter {
  private buffer: AppEvent[] = []
  private readonly MAX_SIZE = 50 // The sliding threshold window size

  constructor() {
    super()
    // Prevent memory leak warning messages from Node runtime if you have many SSE connections
    this.setMaxListeners(0)
  }

  /**
   * Pushes a new incoming RabbitMQ message payload into the buffer, 
   * trims the oldest entries (head), and triggers an event to all active streams.
   */
  public push(payload: any): void {
    const event: AppEvent = {
      id: crypto.randomUUID(),
      payload,
      timestamp: new Date().toISOString()
    }

    this.buffer.push(event)

    // SLIDING WINDOW ENGINE: Trim the head (forget oldest entries)
    if (this.buffer.length > this.MAX_SIZE) {
      this.buffer.shift()
    }

    // Emit event instantly to all active Hono SSE subscribers
    this.emit('new_event', event)
  }

  /**
   * Retrieves a copy of the currently cached historical data.
   * This allows newly connected SSE clients to immediately see the latest timeline history.
   */
  public getHistory(): AppEvent[] {
    return [...this.buffer]
  }

  /**
   * Helper utility to clear memory cache if needed during application runtimes
   */
  public clear(): void {
    this.buffer = []
  }
}

// Export as a single Instance (Singleton design pattern)
// This guarantees both the consumers.ts and sse.controller.ts share the same memory loop.
export const eventBuffer = new EventBufferService()
