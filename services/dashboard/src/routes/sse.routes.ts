import { Hono } from "hono";
import { sseController } from "../controllers/sse.controller";

const sseRoutes = new Hono();

sseRoutes.get('/', sseController.handleStream)

export { sseRoutes };
