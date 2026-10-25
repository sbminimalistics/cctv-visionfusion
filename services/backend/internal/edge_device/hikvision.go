package edge_device

import (
	"encoding/xml"
	"errors"
	"fmt"
)

const (
	IO            DeviceEventType = "IO"
	VMD           DeviceEventType = "VMD"
	LineDetection DeviceEventType = "linedetection"
	VideoLoss     DeviceEventType = "videoloss"
	ANY           DeviceEventType = "*"
)

type HikvisionParsedMessage struct {
	XMLName xml.Name `xml:"EventNotificationAlert"`

	Name             string `xml:"id"`
	IPAddress        string `xml:"ipAddress"`
	EventType        string `xml:"eventType"`
	EventDescription string `xml:"eventDescription"`
	ActivePostCount  int    `xml:"activePostCount"`
}

func (r *HikvisionParsedMessage) GetEventType() string {
	return r.EventType
}

type HikvisionPacketParser struct {
	PacketParser
}

func (h *HikvisionPacketParser) Parse(packet []byte) (BasicParsedMessage, error) {
	var ev HikvisionParsedMessage
	if err := xml.Unmarshal(packet, &ev); err == nil {
		fmt.Printf("@camera EVENT MATCHED: [%s] Description: %s ip: %s \n", ev.EventType, ev.EventDescription, ev.IPAddress)

		return &ev, nil
	} else {
		msg := fmt.Sprintf("unreadable Hikvision XML payload: %v\n", err)
		return nil, errors.New(msg)
	}
}
