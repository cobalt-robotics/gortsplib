package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	"github.com/cobalt-robotics/gortsplib/pkg/h264"
	"github.com/asticode/go-astits"
)

// mpegtsEncoder allows to encode H264 NALUs into MPEG-TS.
type mpegtsEncoder struct {
	sps []byte
	pps []byte

	f                *os.File
	b                *bufio.Writer
	mux              *astits.Muxer
	dtsExtractor     *h264.DTSExtractor
	firstIDRReceived bool
	startDTS         time.Duration
}

// newMPEGTSEncoder allocates a mpegtsEncoder.
func newMPEGTSEncoder(sps []byte, pps []byte) (*mpegtsEncoder, error) {
	f, err := os.Create("mystream.ts")
	if err != nil {
		return nil, err
	}
	b := bufio.NewWriter(f)

	mux := astits.NewMuxer(context.Background(), b)
	mux.AddElementaryStream(astits.PMTElementaryStream{
		ElementaryPID: 256,
		StreamType:    astits.StreamTypeH264Video,
	})
	mux.SetPCRPID(256)

	return &mpegtsEncoder{
		sps: sps,
		pps: pps,
		f:   f,
		b:   b,
		mux: mux,
	}, nil
}

// close closes all the mpegtsEncoder resources.
func (e *mpegtsEncoder) close() {
	e.b.Flush()
	e.f.Close()
}

// encode encodes H264 NALUs into MPEG-TS.
func (e *mpegtsEncoder) encode(nalus [][]byte, pts time.Duration) error {
	// prepend an AUD. This is required by some players
	filteredNALUs := [][]byte{
		{byte(h264.NALUTypeAccessUnitDelimiter), 240},
	}

	nonIDRPresent := false
	idrPresent := false

	for _, nalu := range nalus {
		typ := h264.NALUType(nalu[0] & 0x1F)
		switch typ {
		case h264.NALUTypeSPS:
			e.sps = append([]byte(nil), nalu...)
			continue

		case h264.NALUTypePPS:
			e.pps = append([]byte(nil), nalu...)
			continue

		case h264.NALUTypeAccessUnitDelimiter:
			continue

		case h264.NALUTypeIDR:
			idrPresent = true

			// add SPS and PPS before every IDR
			if e.sps != nil && e.pps != nil {
				filteredNALUs = append(filteredNALUs, e.sps, e.pps)
			}

		case h264.NALUTypeNonIDR:
			nonIDRPresent = true
		}

		filteredNALUs = append(filteredNALUs, nalu)
	}

	if !nonIDRPresent && !idrPresent {
		return nil
	}

	var dts time.Duration

	if !e.firstIDRReceived {
		// skip samples silently until we find one with a IDR
		if !idrPresent {
			return nil
		}

		e.firstIDRReceived = true
		e.dtsExtractor = h264.NewDTSExtractor()

		var err error
		dts, err = e.dtsExtractor.Extract(filteredNALUs, pts)
		if err != nil {
			return err
		}

		e.startDTS = dts
		dts = 0
		pts -= e.startDTS

	} else {
		var err error
		dts, err = e.dtsExtractor.Extract(filteredNALUs, pts)
		if err != nil {
			return err
		}

		dts -= e.startDTS
		pts -= e.startDTS
	}

	oh := &astits.PESOptionalHeader{
		MarkerBits: 2,
	}

	if dts == pts {
		oh.PTSDTSIndicator = astits.PTSDTSIndicatorOnlyPTS
		oh.PTS = &astits.ClockReference{Base: int64(pts.Seconds() * 90000)}
	} else {
		oh.PTSDTSIndicator = astits.PTSDTSIndicatorBothPresent
		oh.DTS = &astits.ClockReference{Base: int64(dts.Seconds() * 90000)}
		oh.PTS = &astits.ClockReference{Base: int64(pts.Seconds() * 90000)}
	}

	// encode into Annex-B
	annexb, err := h264.AnnexBMarshal(filteredNALUs)
	if err != nil {
		return err
	}

	// write TS packet
	_, err = e.mux.WriteData(&astits.MuxerData{
		PID: 256,
		AdaptationField: &astits.PacketAdaptationField{
			RandomAccessIndicator: idrPresent,
		},
		PES: &astits.PESData{
			Header: &astits.PESHeader{
				OptionalHeader: oh,
				StreamID:       224, // video
			},
			Data: annexb,
		},
	})
	if err != nil {
		return err
	}

	log.Println("wrote TS packet")
	return nil
}
