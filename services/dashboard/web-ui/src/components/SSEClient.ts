import {
    EventStreamContentType,
    fetchEventSource,
    type EventSourceMessage,
} from "@microsoft/fetch-event-source";

type EventCallback<T = unknown> = (data: T) => void;

class SSEClient {
    static #instance: SSEClient;
    #ctrl: AbortController | null = null;
    #callbacks = new Map<string, Set<EventCallback<any>>>();

    private constructor() {}

    public static getInstance(): SSEClient {
        if (!SSEClient.#instance) {
            SSEClient.#instance = new SSEClient();
        }
        return SSEClient.#instance;
    }

    public connect(endpointURL: string | URL): void {
        // prevent duplicate streams
        if (this.#ctrl && !this.#ctrl.signal.aborted) {
            return;
        }

        this.#ctrl = new AbortController();

        fetchEventSource(endpointURL.toString(), {
            method: "GET",
            signal: this.#ctrl.signal,
            async onopen(response) {
                const contentType = response.headers.get("content-type");
                if (
                    response.ok &&
                    contentType?.startsWith(EventStreamContentType)
                ) {
                    return;
                }
                // stop retrying on client errors
                if (
                    response.status >= 400 &&
                    response.status < 500 &&
                    response.status !== 429
                ) {
                    throw new Error(`SSE fatal error: ${response.status}`);
                }
            },
            onmessage: this.#onmessage,
            onerror(err) {
                if (
                    err instanceof Error &&
                    err.message.startsWith("SSE fatal error:")
                ) {
                    throw err; // rethrowing stops automatic retries
                }
                // returning nothing lets fetchEventSource retry automatically
            },
        });
    }

    public on<T = unknown>(
        messageType: string,
        cb: EventCallback<T>,
    ): () => void {
        if (!this.#callbacks.has(messageType)) {
            this.#callbacks.set(messageType, new Set());
        }
        this.#callbacks.get(messageType)!.add(cb);

        // returning an unsubscribe function makes React useEffect cleanup trivial
        return () => this.off(messageType, cb);
    }

    public off<T = unknown>(messageType: string, cb: EventCallback<T>): this {
        this.#callbacks.get(messageType)?.delete(cb);
        return this;
    }

    public shutDown(): void {
        this.#ctrl?.abort();
        this.#ctrl = null;
    }

    #onmessage = (msg: EventSourceMessage): void => {
        const callbacks = this.#callbacks.get(msg.event);
        if (!callbacks || callbacks.size === 0) return;

        let parsed: unknown = msg.data;
        try {
            parsed = JSON.parse(msg.data);
        } catch {
            // keep raw string if payload isn't JSON (e.g. plain ping/text)
        }

        for (const cb of callbacks) {
            cb(parsed);
        }
    };
}

export default SSEClient;
