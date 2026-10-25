// src/controllers/sse.controller.ts
import { Context } from "hono";
import { streamSSE } from "hono/streaming";
import { eventBuffer, AppEvent } from "../services/event-buffer.service";

export const sseController = {
    handleStream: async (c: Context) => {
        c.header("Content-Type", "text/event-stream");
        c.header("Cache-Control", "no-cache");
        c.header("Connection", "keep-alive");
        c.header("X-Accel-Buffering", "no"); // Disables Nginx proxy buffering if deployed behind one

        return streamSSE(c, async (stream) => {
            const history = eventBuffer.getHistory();
            for (const oldEvent of history) {
                await stream.writeSSE({
                    data: JSON.stringify(oldEvent.payload),
                    event: "message_log",
                    id: oldEvent.id,
                });
            }

            const handleIncoming = async (event: AppEvent) => {
                await stream.writeSSE({
                    data: JSON.stringify(event.payload),
                    event: "message_log",
                    id: event.id,
                });
            };

            // Hook up this specific client's network stream to the buffer emitter
            eventBuffer.on("new_event", handleIncoming);

            stream.onAbort(() => {
                eventBuffer.off("new_event", handleIncoming);
                console.log(
                    "client disconnected, cleaned up SSE event listener loop",
                );
            });

            while (true) {
                // Sends an empty comment line every 30 seconds as a heartbeat ping
                await stream.write(": heartbeat\n\n");
                await stream.sleep(30000);
            }
        });
    },
};
