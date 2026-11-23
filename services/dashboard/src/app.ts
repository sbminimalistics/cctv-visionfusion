import * as dotenv from "dotenv";
import * as path from "path";
import rmq from "./messaging/rmq.client";
import db from "./db/client";
import { apiRouter } from "./routes";
import { Hono } from "hono";
import { serve } from "@hono/node-server";
import { sseController } from "./controllers/sse.controller";
import { DeviceEvent, CompositionEvent } from "../proto/events";

dotenv.config({ path: path.resolve(process.cwd(), "../.env") });

const getEnvVar = (envVar: string): string => {
    if (process.env[envVar] == null) {
        console.error(`error: environment variable ${envVar} is missing`);
        process.exit(1)
    }
    return process.env[envVar];
}

const RMQ_HOST = getEnvVar("RMQ_HOST");
const RMQ_PORT = getEnvVar("RMQ_PORT");
const RMQ_USER = getEnvVar("DASHBOARD_RMQ_USER");
const RMQ_PASS = getEnvVar("DASHBOARD_RMQ_PASS");
const RMQ_VHOST = getEnvVar("RMQ_VHOST");
const RMQ_QUEUE_DASHBOARD_DEVICE_EVENTS_TRIGGERED = getEnvVar("RMQ_QUEUE_DASHBOARD_DEVICE_EVENTS_TRIGGERED");
const RMQ_QUEUE_DASHBOARD_COMPOSITION_EVENTS_TRIGGERED = getEnvVar("RMQ_QUEUE_DASHBOARD_COMPOSITION_EVENTS_TRIGGERED");

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
        (msg) => {
            // console.log("content:", msg);
            // sseController.handleStream,
        },
        DeviceEvent,
    );
    await rmqClient.consumeFromQueue<CompositionEvent>(
        RMQ_QUEUE_DASHBOARD_COMPOSITION_EVENTS_TRIGGERED,
        (msg) => {
            // console.log("content:", msg);

            // sseController.handleStream,
        },
        CompositionEvent,
    );

    // database wiring
    const res = await db`select version();`;
    console.log("db version:", res);

    // api endpoints wiring
    const app = new Hono();
    app.route("/api", apiRouter);
    serve({ fetch: app.fetch, port: 3000 });
}

bootstrap();
