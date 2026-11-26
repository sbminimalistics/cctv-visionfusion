import * as amqp from "amqplib";
import { MessageFns } from "../../proto/events";
import { SpanStatusCode, trace } from "@opentelemetry/api";
import { logger } from "../logger";
// import { XMLParser } from "fast-xml-parser";

const tracer = trace.getTracer("cctv-dashboard-bff");

export default class RMQClient {
    private connection: amqp.ChannelModel | null = null;
    private connectionOptions: amqp.Options.Connect;
    EXCHANGE_NAME = "events";

    constructor() {
        this.connectionOptions = {};
    }

    async initialize(opts: amqp.Options.Connect) {
        this.connectionOptions = opts;
        try {
            console.log("establishing connection to RabbitMQ broker...");
            this.connection = await amqp.connect(this.connectionOptions, {
                timeout: 10000,
            });
            console.log("RabbitMQ connection successfully started");

            // global shutdown hook for the entire shared connection
            process.once("SIGINT", async () => {
                console.log("closing global RabbitMQ connection...");
                await this.connection?.close();
                process.exit(0);
            });
        } catch (err) {
            console.error(
                "failed to connect to RabbitMQ, retrying in 10s...",
                (err as Error).message,
            );
            setTimeout(() => this.initialize(opts), 10000);
        }
    }

    async consumeFromQueue<T>(
        queueName: string,
        onMessage: (payload: any) => void,
        msgF: MessageFns<T>,
    ) {
        if (!this.connection) {
            throw new Error(
                "Cannot consume: RMQClient has not been initialized yet.",
            );
        }
        try {
            console.log("consume RabbitMQ queue", queueName);
            const connection = await amqp.connect(this.connectionOptions, {
                timeout: 10000,
            });
            const channel = await connection.createChannel();
            console.log("Connected successfully to RabbitMQ Broker!");

            process.once("SIGINT", async () => {
                console.log("Closing consumer channels safely...");
                await channel.close();
                await connection.close();
                process.exit(0);
            });

            // Assumes queue exists externally - throws error if missing
            await channel.checkQueue(queueName);

            await channel.consume(
                queueName,
                async (message) => {
                    if (!message) {
                        console.warn("consumer cancelled by server interface");
                        return;
                    }

                    await tracer.startActiveSpan(
                        `rmq.consume.${queueName}`,
                        async (span) => {
                            try {
                                const routingInfo = message.fields.routingKey;
                                span.setAttribute(
                                    "messaging.rabbitmq.routing_key",
                                    routingInfo,
                                );
                                logger.info({queueName, routingKey: routingInfo})

                                const msg = msgF.decode(message.content);
                                await onMessage(msg); // awaits async SSE broadcast before acking!

                                channel.ack(message);
                                span.setStatus({ code: SpanStatusCode.OK });
                            } catch (processingError) {
                                span.recordException(processingError as Error);
                                span.setStatus({
                                    code: SpanStatusCode.ERROR,
                                    message: (processingError as Error).message,
                                });
                                console.error(
                                    "Error parsing message content:",
                                    processingError,
                                );
                                channel.nack(message, false, false); // see note #2 below!
                            } finally {
                                span.end();
                            }
                        },
                    );
                },
                { noAck: false },
            );
        } catch (connectionError) {
            console.error(
                "Failed to establish communication with RabbitMQ engine:",
                (connectionError as Error).message,
            );
            console.log("Retrying in 10s ...");
            setTimeout(() => this.initialize(this.connectionOptions), 10000);
        }
    }
}
