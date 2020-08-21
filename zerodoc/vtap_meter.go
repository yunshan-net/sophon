package zerodoc

import (
	"strconv"

	"gitlab.x.lan/yunshan/droplet-libs/app"
	"gitlab.x.lan/yunshan/droplet-libs/codec"
)

type VTAPUsageMeter struct {
	PacketTx uint64 `db:"packet_tx"`
	PacketRx uint64 `db:"packet_rx"`
	ByteTx   uint64 `db:"byte_tx"`
	ByteRx   uint64 `db:"byte_rx"`
	L3ByteTx uint64 `db:"l3_byte_tx"`
	L3ByteRx uint64 `db:"l3_byte_rx"`
}

func (m *VTAPUsageMeter) Reverse() {
	m.PacketTx, m.PacketRx = m.PacketRx, m.PacketTx
	m.ByteTx, m.ByteRx = m.ByteRx, m.ByteTx
	m.L3ByteTx, m.L3ByteRx = m.L3ByteRx, m.L3ByteTx
}

func (m *VTAPUsageMeter) ID() uint8 {
	return PACKET_ID
}

func (m *VTAPUsageMeter) Name() string {
	return MeterVTAPNames[m.ID()]
}

func (m *VTAPUsageMeter) VTAPName() string {
	return MeterVTAPNames[m.ID()]
}

func (m *VTAPUsageMeter) Encode(encoder *codec.SimpleEncoder) {
	encoder.WriteVarintU64(m.PacketTx)
	encoder.WriteVarintU64(m.PacketRx)
	encoder.WriteVarintU64(m.ByteTx)
	encoder.WriteVarintU64(m.ByteRx)
	encoder.WriteVarintU64(m.L3ByteTx)
	encoder.WriteVarintU64(m.L3ByteRx)
}

func (m *VTAPUsageMeter) Decode(decoder *codec.SimpleDecoder) {
	m.PacketTx = decoder.ReadVarintU64()
	m.PacketRx = decoder.ReadVarintU64()
	m.ByteTx = decoder.ReadVarintU64()
	m.ByteRx = decoder.ReadVarintU64()
	m.L3ByteTx = decoder.ReadVarintU64()
	m.L3ByteRx = decoder.ReadVarintU64()
}

func (m *VTAPUsageMeter) SortKey() uint64 {
	return uint64(m.ByteTx) + uint64(m.ByteRx)
}

func (m *VTAPUsageMeter) ToKVString() string {
	buffer := make([]byte, app.MAX_DOC_STRING_LENGTH)
	size := m.MarshalTo(buffer)
	return string(buffer[:size])
}

func (m *VTAPUsageMeter) MarshalTo(b []byte) int {
	offset := 0
	offset += copy(b[offset:], "packet_tx=")
	offset += copy(b[offset:], strconv.FormatUint(m.PacketTx, 10))
	offset += copy(b[offset:], "i,packet_rx=")
	offset += copy(b[offset:], strconv.FormatUint(m.PacketRx, 10))
	offset += copy(b[offset:], "i,byte_tx=")
	offset += copy(b[offset:], strconv.FormatUint(m.ByteTx, 10))
	offset += copy(b[offset:], "i,byte_rx=")
	offset += copy(b[offset:], strconv.FormatUint(m.ByteRx, 10))
	offset += copy(b[offset:], "i,l3_byte_tx=")
	offset += copy(b[offset:], strconv.FormatUint(m.L3ByteTx, 10))
	offset += copy(b[offset:], "i,l3_byte_rx=")
	offset += copy(b[offset:], strconv.FormatUint(m.L3ByteRx, 10))
	b[offset] = 'i'
	offset++

	return offset
}

func (m *VTAPUsageMeter) Merge(other *VTAPUsageMeter) {
	m.PacketTx += other.PacketTx
	m.PacketRx += other.PacketRx
	m.ByteTx += other.ByteTx
	m.ByteRx += other.ByteRx
	m.L3ByteTx += other.L3ByteTx
	m.L3ByteRx += other.L3ByteRx
}

func (m *VTAPUsageMeter) ConcurrentMerge(other app.Meter) {
	if other, ok := other.(*VTAPUsageMeter); ok {
		m.Merge(other)
	}
}

func (m *VTAPUsageMeter) SequentialMerge(other app.Meter) {
	m.ConcurrentMerge(other)
}

func (m *VTAPUsageMeter) Fill(ids []uint8, values []interface{}) {
	for i, id := range ids {
		if id <= _METER_INVALID_ || id >= _METER_MAX_ID_ || values[i] == nil {
			continue
		}
		v, ok := values[i].(int64)
		if !ok {
			continue
		}
		switch id {
		case _METER_PACKET_TX:
			m.PacketTx = uint64(v)
		case _METER_PACKET_RX:
			m.PacketRx = uint64(v)
		case _METER_BYTE_TX:
			m.ByteTx = uint64(v)
		case _METER_BYTE_RX:
			m.ByteRx = uint64(v)
		case _METER_L3_BYTE_TX:
			m.L3ByteTx = uint64(v)
		case _METER_L3_BYTE_RX:
			m.L3ByteRx = uint64(v)
		default:
			log.Warningf("unsupport meter id=%d", id)
		}
	}
}
