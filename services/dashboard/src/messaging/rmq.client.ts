import * as amqp from "amqplib";
import { MessageFns } from "../../proto/events";
// import { XMLParser } from "fast-xml-parser";

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
            console.log(
                "Connecting to RabbitMQ Broker!",
                this.connectionOptions,
            );
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
                (message) => {
                    if (!message) {
                        console.warn("Consumer cancelled by server interface.");
                        return;
                    }

                    try {
                        const payloadString = message.content.toString();
                        const routingInfo = message.fields.routingKey;

                        console.log(
                            `rmq message from queue: ${queueName}, routing key: ${routingInfo}`,
                        );

                        const msg = msgF.decode(message.content);
                        // const { rawPayload, ...msgWithoutRaw } = msg;
                        // const { deviceId, eventType, manufacturer} = msgWithoutRaw;
                        // const originalEvent = msg.rawPayload.toString();
                        // const parser = new XMLParser();
                        // const doc = parser.parse(originalEvent);

                        // console.log("rmq message:", msgWithoutRaw, doc);
                        // console.log("rmq message:", deviceId, eventType, "manufacturer:", manufacturer);

                        onMessage(msg);

                        channel.ack(message);
                        // console.log(
                        //     "Message successfully processed and Acknowledged (ACK).",
                        // );
                    } catch (processingError) {
                        console.error(
                            "Error parsing message content:",
                            processingError,
                        );
                        channel.nack(message, false, true);
                    }
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
