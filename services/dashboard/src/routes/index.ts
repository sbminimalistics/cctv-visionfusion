import { Hono } from "hono";
import { sseRoutes } from "./sse.routes";

export const apiRouter = new Hono();

apiRouter.route("/sse", sseRoutes);

export default apiRouter;