import { Hono } from "hono";
import { sseController } from "../controllers/sse.controller";

const sseRoutes = new Hono();

sseRoutes.get('/stream', sseController.handleStream)

export { sseRoutes };
