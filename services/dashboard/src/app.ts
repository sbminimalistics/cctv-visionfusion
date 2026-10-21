import * as dotenv from "dotenv";
import * as path from "path";
import rmq from "./messaging/rmq.ts";
import db from "./db/database.ts";

dotenv.config({ path: path.resolve(process.cwd(), "../.env") });

const RMQ_HOST = process.env.RMQ_HOST || "localhost";
const RMQ_PORT = process.env.RMQ_PORT || "5672";
const RMQ_USER = process.env.DASHBOARD_RMQ_USER || "dashboard_worker";
const RMQ_PASS = process.env.DASHBOARD_RMQ_PASS;
const RMQ_VHOST = process.env.RMQ_VHOST || "vms";
const RMQ_QUEUE_DASHBOARD_FEED = process.env.RMQ_QUEUE_DASHBOARD_FEED || "dashboard_feed";

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
    consumer.connect(opts, RMQ_QUEUE_DASHBOARD_FEED);

    const res = await db`select version();`;
    console.log("db version:", res);
}

bootstrap();
