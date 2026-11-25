import {
    createContext,
    useContext,
    useEffect,
    useReducer,
    type Dispatch,
    type ReactNode,
} from "react";
import SSEClient from "./SSEClient";
import { DeviceEvent, CompositeEvent } from "../../../proto/events";
import { env } from "../config/env";

interface DashboardState {
    recentEvents: DeviceEvent[];
    latestDetectionsByEventId: Record<string, CompositeEvent>;
    isConnected: boolean;
}

type DashboardAction =
    | { type: "SSE_STATUS"; connected: boolean }
    | { type: "DEVICE_EVENT"; payload: DeviceEvent }
    | { type: "COMPOSITE_EVENT"; payload: CompositeEvent }
    | { type: "CLEAR_EVENTS" };

const MAX_RECENT_EVENTS = 100;

const initialState: DashboardState = {
    recentEvents: [],
    latestDetectionsByEventId: {},
    isConnected: false,
};

function dashboardReducer(
    state: DashboardState,
    action: DashboardAction,
): DashboardState {
    switch (action.type) {
        case "SSE_STATUS":
            return { ...state, isConnected: action.connected };

        case "DEVICE_EVENT":
            return {
                ...state,
                recentEvents: [action.payload, ...state.recentEvents].slice(
                    0,
                    MAX_RECENT_EVENTS,
                ),
            };

        case "COMPOSITE_EVENT":
            return {
                ...state,
                latestDetectionsByEventId: {
                    ...state.latestDetectionsByEventId,
                    [action.payload.uuid]: action.payload,
                },
            };

        case "CLEAR_EVENTS":
            return {
                ...state,
                recentEvents: [],
                latestDetectionsByEventId: {},
            };

        default:
            return state;
    }
}

const DashboardStateContext = createContext<DashboardState | null>(null);
const DashboardDispatchContext =
    createContext<Dispatch<DashboardAction> | null>(null);

export function DashboardProvider({ children }: { children: ReactNode }) {
    const [state, dispatch] = useReducer(dashboardReducer, initialState);

    useEffect(() => {
        const client = SSEClient.getInstance();
        
        const sseSource = new URL(env.BFF_HOST);
        sseSource.port = env.BFF_PORT;
        const prefix = (env.BFF_API_PREFIX ?? "").replace(/\/+$/, "");
        sseSource.pathname = `${prefix}/sse`;
        client.connect(sseSource);

        const offCameraEvent = client.on("device_event", (raw) => {
            dispatch({
                type: "DEVICE_EVENT",
                payload: DeviceEvent.fromJSON(raw),
            });
        });

        const offDetection = client.on("detection_result", (raw) => {
            dispatch({
                type: "COMPOSITE_EVENT",
                payload: CompositeEvent.fromJSON(raw),
            });
        });

        return () => {
            offCameraEvent();
            offDetection();
            client.shutDown();
        };
    }, []);

    return (
        <DashboardStateContext.Provider value={state}>
            <DashboardDispatchContext.Provider value={dispatch}>
                {children}
            </DashboardDispatchContext.Provider>
        </DashboardStateContext.Provider>
    );
}

export function useDashboardState(): DashboardState {
    const ctx = useContext(DashboardStateContext);
    if (!ctx)
        throw new Error("useDashboardState must be inside DashboardProvider");
    return ctx;
}

export function useDashboardDispatch(): Dispatch<DashboardAction> {
    const ctx = useContext(DashboardDispatchContext);
    if (!ctx)
        throw new Error(
            "useDashboardDispatch must be inside DashboardProvider",
        );
    return ctx;
}
