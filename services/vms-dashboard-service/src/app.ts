import * as dotenv from "dotenv";
import * as path from "path";
import rmq from "./messaging/rmq.ts";

dotenv.config({ path: path.resolve(process.cwd(), "../.env") });

const RMQ_HOST = process.env.RMQ_HOST || "localhost";
const RMQ_PORT = process.env.RMQ_PORT || "5672";
const RMQ_USER = process.env.DASHBOARD_RMQ_USER || "consumer_worker";
const RMQ_PASS = process.env.DASHBOARD_RMQ_PASS;
const RMQ_VHOST = process.env.RMQ_VHOST || "/vms";

async function bootstrap() {
    if (!RMQ_PASS) {
        console.error("Error: NODE_RMQ_PASS environment variable is missing.");
        process.exit(1);
    }

    const connectionUrl = `amqp://${encodeURIComponent(RMQ_USER)}:${encodeURIComponent(RMQ_PASS)}@${RMQ_HOST}:${RMQ_PORT}${RMQ_VHOST}`;

    // connectMessanger(connectionUrl);
    console.log(
        `Attempting connection to RabbitMQ at ${RMQ_HOST}:${RMQ_PORT}${RMQ_VHOST}...`,
    );
    const consumer = new rmq();
    const opts = {
        protocol: "amqp",
        hostname: RMQ_HOST,
        port: parseInt(RMQ_PORT),
        username: RMQ_USER,
        password: RMQ_PASS,
        vhost: RMQ_VHOST,
    };
    consumer.connect(opts, "vms.alerts");
}

/* async function connectMessanger(url: string) {
    try {
        const connection = await amqp.connect(url);
        const channel = await connection.createChannel();
        console.log("Connected successfully to RabbitMQ Broker!");

        process.once("SIGINT", async () => {
            console.log("Closing consumer channels safely...");
            await channel.close();
            await connection.close();
            process.exit(0);
        });

        await channel.assertExchange(EXCHANGE_NAME, "topic", { durable: true });

        await channel.assertQueue(QUEUE_NAME, {
            durable: true,
            arguments: {
                "x-max-length": 5000,
                "x-overflow": "drop-head",
                "x-message-ttl": 86400000,
            },
        });

        await channel.bindQueue(QUEUE_NAME, EXCHANGE_NAME, ROUTING_KEY);
        console.log(
            `Subscribed to queue: [${QUEUE_NAME}]. Awaiting stream notifications...`,
        );

        await channel.consume(
            QUEUE_NAME,
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
        console.log("retrying in 5s ...");
        setTimeout(connectMessanger, 5000);
    }
} */

bootstrap();
