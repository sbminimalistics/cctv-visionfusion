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

const RMQ_HOST = process.env.RMQ_HOST || "localhost";
const RMQ_PORT = process.env.RMQ_PORT || "5672";
const RMQ_USER = process.env.DASHBOARD_RMQ_USER || "dashboard_worker";
const RMQ_PASS = process.env.DASHBOARD_RMQ_PASS;
const RMQ_VHOST = process.env.RMQ_VHOST || "cctv-vf";
const RMQ_QUEUE_DASHBOARD_FEED =
    process.env.RMQ_QUEUE_DASHBOARD_FEED || "dashboard_feed";
const RMQ_QUEUE_DASHBOARD_TRIGGERS =
    process.env.RMQ_QUEUE_DASHBOARD_TRIGGERS || "dashboard_triggers";

async function bootstrap() {
    // messaging wiring
    if (!RMQ_PASS) {
        console.error("Error: NODE_RMQ_PASS environment variable is missing.");
        process.exit(1);
    }

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
        RMQ_QUEUE_DASHBOARD_FEED,
        (msg) => {
            // console.log("content:", msg);
            // sseController.handleStream,
        },
        DeviceEvent,
    );
    await rmqClient.consumeFromQueue<CompositionEvent>(
        RMQ_QUEUE_DASHBOARD_TRIGGERS,
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
