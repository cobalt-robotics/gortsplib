package headers

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cobalt-robotics/gortsplib/pkg/base"
)

var casesTransport = []struct {
	name string
	vin  base.HeaderValue
	vout base.HeaderValue
	h    Transport
}{
	{
		"udp unicast play request",
		base.HeaderValue{`RTP/AVP;unicast;client_port=3456-3457;mode="PLAY"`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=3456-3457;mode=play`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			ClientPorts: &[2]int{3456, 3457},
			Mode: func() *TransportMode {
				v := TransportModePlay
				return &v
			}(),
		},
	},
	{
		"udp unicast play response",
		base.HeaderValue{`RTP/AVP/UDP;unicast;client_port=3056-3057;server_port=5000-5001`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=3056-3057;server_port=5000-5001`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			ClientPorts: &[2]int{3056, 3057},
			ServerPorts: &[2]int{5000, 5001},
		},
	},
	{
		"udp multicast play request / response",
		base.HeaderValue{`RTP/AVP;multicast;destination=225.219.201.15;port=7000-7001;ttl=127`},
		base.HeaderValue{`RTP/AVP;multicast;destination=225.219.201.15;port=7000-7001;ttl=127`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryMulticast
				return &v
			}(),
			Destination: func() *net.IP {
				v := net.ParseIP("225.219.201.15")
				return &v
			}(),
			TTL: func() *uint {
				v := uint(127)
				return &v
			}(),
			Ports: &[2]int{7000, 7001},
		},
	},
	{
		"tcp play request / response",
		base.HeaderValue{`RTP/AVP/TCP;interleaved=0-1`},
		base.HeaderValue{`RTP/AVP/TCP;interleaved=0-1`},
		Transport{
			Protocol:       TransportProtocolTCP,
			InterleavedIDs: &[2]int{0, 1},
		},
	},
	{
		"udp unicast play response with a single port and ssrc",
		base.HeaderValue{`RTP/AVP/UDP;unicast;server_port=8052;client_port=14186;ssrc=0B6020AD;mode=PLAY`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=14186-14187;server_port=8052-8053;ssrc=0B6020AD;mode=play`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			Mode: func() *TransportMode {
				v := TransportModePlay
				return &v
			}(),
			ClientPorts: &[2]int{14186, 14187},
			ServerPorts: &[2]int{8052, 8053},
			SSRC: func() *uint32 {
				v := uint32(0x0B6020AD)
				return &v
			}(),
		},
	},
	{
		"udp record response with receive",
		base.HeaderValue{`RTP/AVP/UDP;unicast;mode=receive;source=127.0.0.1;client_port=14186-14187;server_port=5000-5001`},
		base.HeaderValue{`RTP/AVP;unicast;source=127.0.0.1;client_port=14186-14187;server_port=5000-5001;mode=record`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			Mode: func() *TransportMode {
				v := TransportModeRecord
				return &v
			}(),
			ClientPorts: &[2]int{14186, 14187},
			ServerPorts: &[2]int{5000, 5001},
			Source: func() *net.IP {
				v := net.ParseIP("127.0.0.1")
				return &v
			}(),
		},
	},
	{
		"unsorted udp unicast play request headers",
		base.HeaderValue{`client_port=3456-3457;RTP/AVP;mode="PLAY";unicast`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=3456-3457;mode=play`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			ClientPorts: &[2]int{3456, 3457},
			Mode: func() *TransportMode {
				v := TransportModePlay
				return &v
			}(),
		},
	},
	{
		"ssrc odd",
		base.HeaderValue{`RTP/AVP/UDP;unicast;client_port=14186;server_port=8052;ssrc=4317f;mode=play`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=14186-14187;server_port=8052-8053;ssrc=0004317F;mode=play`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			Mode: func() *TransportMode {
				v := TransportModePlay
				return &v
			}(),
			ClientPorts: &[2]int{14186, 14187},
			ServerPorts: &[2]int{8052, 8053},
			SSRC: func() *uint32 {
				v := uint32(0x04317f)
				return &v
			}(),
		},
	},
	{
		"hikvision ssrc with initial spaces",
		base.HeaderValue{`RTP/AVP/UDP;unicast;client_port=14186;server_port=8052;ssrc= 4317f;mode=play`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=14186-14187;server_port=8052-8053;ssrc=0004317F;mode=play`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			Mode: func() *TransportMode {
				v := TransportModePlay
				return &v
			}(),
			ClientPorts: &[2]int{14186, 14187},
			ServerPorts: &[2]int{8052, 8053},
			SSRC: func() *uint32 {
				v := uint32(0x04317f)
				return &v
			}(),
		},
	},
	{
		"dahua rtsp server ssrc with initial spaces",
		base.HeaderValue{`RTP/AVP/TCP;unicast;interleaved=0-1;ssrc=     D93FF`},
		base.HeaderValue{`RTP/AVP/TCP;unicast;interleaved=0-1;ssrc=000D93FF`},
		Transport{
			Protocol: TransportProtocolTCP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			InterleavedIDs: &[2]int{0, 1},
			SSRC: func() *uint32 {
				v := uint32(0xD93FF)
				return &v
			}(),
		},
	},
	{
		"empty source",
		base.HeaderValue{`RTP/AVP/UDP;unicast;source=;client_port=32560-32561;server_port=3046-3047;ssrc=45dcb578`},
		base.HeaderValue{`RTP/AVP;unicast;client_port=32560-32561;server_port=3046-3047;ssrc=45DCB578`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			SSRC: func() *uint32 {
				v := uint32(0x45dcb578)
				return &v
			}(),
			ClientPorts: &[2]int{32560, 32561},
			ServerPorts: &[2]int{3046, 3047},
		},
	},
	{
		"invalid ssrc",
		base.HeaderValue{`RTP/AVP;unicast;client_port=14236;source=172.16.8.2;server_port=56002;ssrc=1449463210`},
		base.HeaderValue{`RTP/AVP;unicast;source=172.16.8.2;client_port=14236-14237;server_port=56002-56003`},
		Transport{
			Protocol: TransportProtocolUDP,
			Delivery: func() *TransportDelivery {
				v := TransportDeliveryUnicast
				return &v
			}(),
			Source: func() *net.IP {
				v := net.ParseIP("172.16.8.2")
				return &v
			}(),
			ClientPorts: &[2]int{14236, 14237},
			ServerPorts: &[2]int{56002, 56003},
		},
	},
}

func TestTransportUnmarshal(t *testing.T) {
	for _, ca := range casesTransport {
		t.Run(ca.name, func(t *testing.T) {
			var h Transport
			err := h.Unmarshal(ca.vin)
			require.NoError(t, err)
			require.Equal(t, ca.h, h)
		})
	}
}

func TestTransportUnmarshalErrors(t *testing.T) {
	for _, ca := range []struct {
		name string
		hv   base.HeaderValue
		err  string
	}{
		{
			"empty",
			base.HeaderValue{},
			"value not provided",
		},
		{
			"2 values",
			base.HeaderValue{"a", "b"},
			"value provided multiple times ([a b])",
		},
		{
			"invalid keys",
			base.HeaderValue{`key1="k`},
			"apexes not closed (key1=\"k)",
		},
		{
			"protocol not found",
			base.HeaderValue{`invalid;unicast;client_port=14186-14187`},
			"protocol not found (invalid;unicast;client_port=14186-14187)",
		},
		{
			"invalid interleaved port",
			base.HeaderValue{`RTP/AVP;unicast;interleaved=aa-14187`},
			"invalid ports (aa-14187)",
		},
		{
			"invalid ttl",
			base.HeaderValue{`RTP/AVP;unicast;ttl=aa`},
			"strconv.ParseUint: parsing \"aa\": invalid syntax",
		},
		{
			"invalid destination",
			base.HeaderValue{`RTP/AVP;unicast;destination=aa`},
			"invalid destination (aa)",
		},
		{
			"invalid ports 1",
			base.HeaderValue{`RTP/AVP;unicast;port=aa`},
			"invalid ports (aa)",
		},
		{
			"invalid ports 2",
			base.HeaderValue{`RTP/AVP;unicast;port=aa-bb-cc`},
			"invalid ports (aa-bb-cc)",
		},
		{
			"invalid ports 3",
			base.HeaderValue{`RTP/AVP;unicast;port=aa-14187`},
			"invalid ports (aa-14187)",
		},
		{
			"invalid ports 4",
			base.HeaderValue{`RTP/AVP;unicast;port=14186-aa`},
			"invalid ports (14186-aa)",
		},
		{
			"invalid client port",
			base.HeaderValue{`RTP/AVP;unicast;client_port=aa-14187`},
			"invalid ports (aa-14187)",
		},
		{
			"invalid server port",
			base.HeaderValue{`RTP/AVP;unicast;server_port=aa-14187`},
			"invalid ports (aa-14187)",
		},
		{
			"invalid mode",
			base.HeaderValue{`RTP/AVP;unicast;mode=aa`},
			"invalid transport mode: 'aa'",
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			var h Transport
			err := h.Unmarshal(ca.hv)
			require.EqualError(t, err, ca.err)
		})
	}
}

func TestTransportMarshal(t *testing.T) {
	for _, ca := range casesTransport {
		t.Run(ca.name, func(t *testing.T) {
			req := ca.h.Marshal()
			require.Equal(t, ca.vout, req)
		})
	}
}
