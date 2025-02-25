package rtpreorderer

import (
	"github.com/pion/rtp"
)

const (
	bufferSize = 64
)

// Reorderer filters incoming RTP packets, in order to
// - order packets
// - remove duplicate packets
type Reorderer struct {
	initialized    bool
	expectedSeqNum uint16
	buffer         []*rtp.Packet
	absPos         uint16
}

// New allocates a Reorderer.
func New() *Reorderer {
	return &Reorderer{
		buffer: make([]*rtp.Packet, bufferSize),
	}
}

// Process processes a RTP packet.
func (r *Reorderer) Process(pkt *rtp.Packet) []*rtp.Packet {
	if !r.initialized {
		r.initialized = true
		r.expectedSeqNum = pkt.SequenceNumber + 1
		return []*rtp.Packet{pkt}
	}

	relPos := pkt.SequenceNumber - r.expectedSeqNum

	// packet is a duplicate or has been sent
	// before the first packet processed by Reorderer.
	// discard.
	if relPos > 0xFFF {
		return nil
	}

	// there's a missing packet and buffer is full.
	// return entire buffer and clear it.
	if relPos >= bufferSize {
		n := 1
		for i := uint16(0); i < bufferSize; i++ {
			p := (r.absPos + i) & (bufferSize - 1)
			if r.buffer[p] != nil {
				n++
			}
		}

		ret := make([]*rtp.Packet, n)
		pos := 0

		for i := uint16(0); i < bufferSize; i++ {
			p := (r.absPos + i) & (bufferSize - 1)
			if r.buffer[p] != nil {
				ret[pos] = r.buffer[p]
				pos++
			}
		}

		ret[pos] = pkt

		for i := 0; i < bufferSize; i++ {
			r.buffer[i] = nil
		}

		r.expectedSeqNum = pkt.SequenceNumber + 1
		return ret
	}

	// there's a missing packet
	if relPos != 0 {
		p := (r.absPos + relPos) & (bufferSize - 1)

		// current packet is a duplicate. discard
		if r.buffer[p] != nil {
			return nil
		}

		// put current packet in buffer
		r.buffer[p] = pkt
		return nil
	}

	// all packets have been received correctly.
	// return them

	n := uint16(1)
	for {
		p := (r.absPos + n) & (bufferSize - 1)
		if r.buffer[p] == nil {
			break
		}
		n++
	}

	ret := make([]*rtp.Packet, n)
	ret[0] = pkt

	r.absPos++
	r.absPos &= (bufferSize - 1)

	for i := uint16(1); i < n; i++ {
		ret[i], r.buffer[r.absPos] = r.buffer[r.absPos], nil
		r.absPos++
		r.absPos &= (bufferSize - 1)
	}

	r.expectedSeqNum = pkt.SequenceNumber + n

	return ret
}
