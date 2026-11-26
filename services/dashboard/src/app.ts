import "./instrumentation";
import * as dotenv from "dotenv";
import * as path from "path";
import rmq from "./messaging/rmq.client";
import db from "./db/client";
import { apiRouter } from "./routes";
import { Hono } from "hono";
import { serve } from "@hono/node-server";
import { sseController } from "./controllers/sse.controller";
import { DeviceEvent, CompositeEvent } from "../proto/events";
import { cors } from "hono/cors"
import { rmqMessageController } from "./controllers/rmq-message.controller";
import { logger } from "./logger";

// dotenv.config({ path: path.resolve(process.cwd(), "./../../.env_dev") });

logger.info({"position": "app"})

const getEnvVar = (envVar: string): string => {
    if (process.env[envVar] == null) {
        console.error(`error: environment variable ${envVar} is missing`);
        process.exit(1);
    }
    return process.env[envVar];
};

const RMQ_HOST = getEnvVar("RMQ_HOST");
const RMQ_PORT = getEnvVar("RMQ_PORT");
const RMQ_USER = getEnvVar("DASHBOARD_RMQ_USER");
const RMQ_PASS = getEnvVar("DASHBOARD_RMQ_PASS");
const RMQ_VHOST = getEnvVar("RMQ_VHOST");
const RMQ_QUEUE_DASHBOARD_DEVICE_EVENTS_TRIGGERED = getEnvVar(
    "RMQ_QUEUE_DASHBOARD_DEVICE_EVENTS_TRIGGERED",
);
const RMQ_QUEUE_DASHBOARD_COMPOSITE_EVENTS_TRIGGERED = getEnvVar(
    "RMQ_QUEUE_DASHBOARD_COMPOSITE_EVENTS_TRIGGERED",
);

async function bootstrap() {
    const rmqClient = new rmq();
    const opts = {
        protocol: "amqp",
        hostname: RMQ_HOST,
        port: parseInt(RMQ_PORT),
        username: RMQ_USER,
        password: RMQ_PASS,
        vhost: RMQ_VHOST,
    };
    await rmqClient.initialize(opts);
    await rmqClient.consumeFromQueue<DeviceEvent>(
        RMQ_QUEUE_DASHBOARD_DEVICE_EVENTS_TRIGGERED,
        rmqMessageController.handleDeviceEventMessage,
        DeviceEvent,
    );
    await rmqClient.consumeFromQueue<CompositeEvent>(
        RMQ_QUEUE_DASHBOARD_COMPOSITE_EVENTS_TRIGGERED,
        rmqMessageController.handleCompositeEventMessage,
        CompositeEvent,
    );

    // database wiring
    const [{ version }] = await db<{ version: string }[]>`select version();`;
    console.log("db version:", version);

    // api endpoints wiring
    const app = new Hono();
    app.use('/api/*', cors())

    if (process.env.BFF_API_PREFIX == null) {
        throw new Error("BFF_API_PREFIX not set");
    }
    app.route(process.env.BFF_API_PREFIX, apiRouter);

    if (process.env.BFF_PORT == null) {
        throw new Error("BFF_PORT not set");
    }
    console.log(
        "port:",
        process.env.BFF_PORT,
        "api_prefix:",
        process.env.BFF_API_PREFIX,
    );
    serve({ fetch: app.fetch, port: parseInt(process.env.BFF_PORT) });
}

bootstrap();
