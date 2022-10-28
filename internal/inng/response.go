package inng

import (
	"encoding/binary"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

type ResponseImageProcessing struct {
	ReqMessageId string  `cbor:"1,keyasint"` // `cbor:"req_message_id"`
	Success      bool    `cbor:"2,keyasint"` // `cbor:"success"`
	SrcWidth     uint32  `cbor:"3,keyasint"` // `cbor:"src_width"`
	SrcHeight    uint32  `cbor:"4,keyasint"` // `cbor:"src_height"`
	DstWidth     uint32  `cbor:"5,keyasint"` // `cbor:"dst_width"`
	DstHeight    uint32  `cbor:"6,keyasint"` // `cbor:"dst_height"`
	Message      *string `cbor:"7,keyasint"` // `cbor:"message"`
}

func (r *ResponseImageProcessing) MarshalCborWithPayload(payload []byte) ([]byte, error) {
	var payloadSz int
	if payload != nil {
		payloadSz = len(payload)
	}

	cborBytes, err := cbor.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal CBOR message: %v", err)
	}

	resBytes := make([]byte, 4+len(cborBytes)+payloadSz)
	binary.LittleEndian.PutUint32(resBytes, uint32(payloadSz))

	copy(resBytes[4:], cborBytes)
	if payload != nil {
		copy(resBytes[4+len(cborBytes):], payload)
	}

	return resBytes, nil
}

func (r *ResponseImageProcessing) UnmarshalCborWithPayload(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		return nil, fmt.Errorf("Empty message")
	}

	resBytesSize := len(msg)
	if resBytesSize < 4 {
		return nil, fmt.Errorf("Incorrect message size")
	}

	resPayloadSize := binary.LittleEndian.Uint32(msg[:4])
	if (resPayloadSize < 0) || (resPayloadSize > (uint32(resBytesSize) - 4)) {
		return nil, fmt.Errorf("Incorrect message payload part size")
	}

	resCborEnd := resBytesSize - int(resPayloadSize)
	if resCborEnd > 4 {
		if err := cbor.Unmarshal(msg[4:resCborEnd], r); err != nil {
			return nil, err
		}
	}

	var resPayload []byte
	if resPayloadSize > 0 {
		resPayload = msg[resCborEnd:]
	}

	return resPayload, nil
}
