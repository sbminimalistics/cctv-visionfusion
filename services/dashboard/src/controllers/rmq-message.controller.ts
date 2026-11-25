import { CompositeEvent, DeviceEvent } from "../../proto/events";
import { eventBuffer } from "../services/event-buffer.service";

export const rmqMessageController = {
    handleDeviceEventMessage: (data: any) => {
        let ev = DeviceEvent.toJSON(data);
        eventBuffer.push(ev);
    },

    handleCompositeEventMessage: (data: any) => {
        let ev = CompositeEvent.toJSON(data);
        eventBuffer.push(ev);
    }
}