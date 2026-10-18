import * as amqp from "amqplib";

export default class RMQClient {
    queue: string = "";
    connectionOptions: amqp.Options.Connect;

    EXCHANGE_NAME = "vms.events";

    constructor() {
        this.connectionOptions = {};
    }

    async connect(opts: amqp.Options.Connect, queueName: string) {
        this.connectionOptions = opts;
        this.queue = queueName;
        return this._connect();
    }

    private async _connect() {
        try {
            console.log(
                "connectingto RabbitMQ Broker!",
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

            await channel.checkQueue(this.queue);

            await channel.consume(
                this.queue,
                (message) => {
                    if (!message) {
                        console.warn("Consumer cancelled by server interface.");
                        return;
                    }

                    try {
                        const payloadString = message.content.toString();
                        const routingInfo = message.fields.routingKey;

                        console.log(
                            `\n[NEW ALERT] Intercepted on Routing Key: ${routingInfo}`,
                        );
                        console.log(`Payload Data: ${payloadString}`);

                        channel.ack(message);
                        console.log(
                            "Message successfully processed and Acknowledged (ACK).",
                        );
                    } catch (processingError) {
                        console.error(
                            "Error parsing message content:",
                            processingError,
                        );
                        channel.nack(message, false, true);
                    }
                },
                {
                    noAck: false,
                },
            );
        } catch (connectionError) {
            console.error(
                "Failed to establish communication with RabbitMQ engine:",
                (connectionError as Error).message,
            );
            console.log("retrying in 10s ...");
            setTimeout(() => this._connect(), 10000);
        }
    }
}
