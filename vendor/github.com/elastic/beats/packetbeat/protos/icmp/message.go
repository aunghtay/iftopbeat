package icmp

import (
	"encoding/binary"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/tsg/gopacket/layers"
)

// TODO: more types (that are not provided as constants in gopacket)

// ICMPv4 types that represent a response (all other types represent a request)
var icmp4ResponseTypes = map[uint8]bool{
	layers.ICMPv4TypeEchoReply:        true,
	layers.ICMPv4TypeTimestampReply:   true,
	layers.ICMPv4TypeInfoReply:        true,
	layers.ICMPv4TypeAddressMaskReply: true,
}

// ICMPv6 types that represent a response (all other types represent a request)
var icmp6ResponseTypes = map[uint8]bool{
	layers.ICMPv6TypeEchoReply: true,
}

// ICMPv4 types that represent an error
var icmp4ErrorTypes = map[uint8]bool{
	layers.ICMPv4TypeDestinationUnreachable: true,
	layers.ICMPv4TypeSourceQuench:           true,
	layers.ICMPv4TypeTimeExceeded:           true,
	layers.ICMPv4TypeParameterProblem:       true,
}

// ICMPv6 types that represent an error
var icmp6ErrorTypes = map[uint8]bool{
	layers.ICMPv6TypeDestinationUnreachable: true,
	layers.ICMPv6TypePacketTooBig:           true,
	layers.ICMPv6TypeTimeExceeded:           true,
	layers.ICMPv6TypeParameterProblem:       true,
}

// ICMPv4 types that require a request & a response
var icmp4PairTypes = map[uint8]bool{
	layers.ICMPv4TypeEchoRequest:        true,
	layers.ICMPv4TypeEchoReply:          true,
	layers.ICMPv4TypeTimestampRequest:   true,
	layers.ICMPv4TypeTimestampReply:     true,
	layers.ICMPv4TypeInfoRequest:        true,
	layers.ICMPv4TypeInfoReply:          true,
	layers.ICMPv4TypeAddressMaskRequest: true,
	layers.ICMPv4TypeAddressMaskReply:   true,
}

// ICMPv6 types that require a request & a response
var icmp6PairTypes = map[uint8]bool{
	layers.ICMPv6TypeEchoRequest: true,
	layers.ICMPv6TypeEchoReply:   true,
}

// Contains all used information from the ICMP message on the wire.
type icmpMessage struct {
	Ts     time.Time
	Type   uint8
	Code   uint8
	Length int
}

func isRequest(tuple *icmpTuple, msg *icmpMessage) bool {
	if tuple.IcmpVersion == 4 {
		return !icmp4ResponseTypes[msg.Type]
	}
	if tuple.IcmpVersion == 6 {
		return !icmp6ResponseTypes[msg.Type]
	}
	logp.WTF("icmp", "Invalid ICMP version[%d]", tuple.IcmpVersion)
	return true
}

func isError(tuple *icmpTuple, msg *icmpMessage) bool {
	if tuple.IcmpVersion == 4 {
		return icmp4ErrorTypes[msg.Type]
	}
	if tuple.IcmpVersion == 6 {
		return icmp6ErrorTypes[msg.Type]
	}
	logp.WTF("icmp", "Invalid ICMP version[%d]", tuple.IcmpVersion)
	return true
}

func requiresCounterpart(tuple *icmpTuple, msg *icmpMessage) bool {
	if tuple.IcmpVersion == 4 {
		return icmp4PairTypes[msg.Type]
	}
	if tuple.IcmpVersion == 6 {
		return icmp6PairTypes[msg.Type]
	}
	logp.WTF("icmp", "Invalid ICMP version[%d]", tuple.IcmpVersion)
	return false
}

func extractTrackingData(icmpVersion uint8, msgType uint8, baseLayer *layers.BaseLayer) (uint16, uint16) {
	if icmpVersion == 4 {
		if icmp4PairTypes[msgType] {
			id := binary.BigEndian.Uint16(baseLayer.Contents[4:6])
			seq := binary.BigEndian.Uint16(baseLayer.Contents[6:8])
			return id, seq
		}
		return 0, 0
	}
	if icmpVersion == 6 {
		if icmp6PairTypes[msgType] {
			id := binary.BigEndian.Uint16(baseLayer.Contents[4:6])
			seq := binary.BigEndian.Uint16(baseLayer.Contents[6:8])
			return id, seq
		}
		return 0, 0
	}
	logp.WTF("icmp", "Invalid ICMP version[%d]", icmpVersion)
	return 0, 0
}

func humanReadable(tuple *icmpTuple, msg *icmpMessage) string {
	if tuple.IcmpVersion == 4 {
		return layers.ICMPv4TypeCode(binary.BigEndian.Uint16([]byte{msg.Type, msg.Code})).String()
	}
	if tuple.IcmpVersion == 6 {
		return layers.ICMPv6TypeCode(binary.BigEndian.Uint16([]byte{msg.Type, msg.Code})).String()
	}
	logp.WTF("icmp", "Invalid ICMP version[%d]", tuple.IcmpVersion)
	return ""
}