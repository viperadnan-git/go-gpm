package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protowire"
)

const libraryStateEndpoint = "https://photosdata-pa.googleapis.com/6439526531001121323/18047484249733410717"

// Helper functions to build protobuf messages manually

// appendEmptyMessage appends an empty sub-message at the given field number
func appendEmptyMessage(b []byte, fieldNum protowire.Number) []byte {
	b = protowire.AppendTag(b, fieldNum, protowire.BytesType)
	b = protowire.AppendBytes(b, nil)
	return b
}

// appendStringField appends a string field
func appendStringField(b []byte, fieldNum protowire.Number, value string) []byte {
	if value == "" {
		return b
	}
	b = protowire.AppendTag(b, fieldNum, protowire.BytesType)
	b = protowire.AppendString(b, value)
	return b
}

// appendVarintField appends a varint field
func appendVarintField(b []byte, fieldNum protowire.Number, value int64) []byte {
	b = protowire.AppendTag(b, fieldNum, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(value))
	return b
}

// appendBytesField appends a bytes field
func appendBytesField(b []byte, fieldNum protowire.Number, value []byte) []byte {
	b = protowire.AppendTag(b, fieldNum, protowire.BytesType)
	b = protowire.AppendBytes(b, value)
	return b
}

// appendRepeatedVarint appends a repeated varint field (packed)
func appendRepeatedVarint(b []byte, fieldNum protowire.Number, values []int64) []byte {
	if len(values) == 0 {
		return b
	}
	// Build packed repeated field
	var packed []byte
	for _, v := range values {
		packed = protowire.AppendVarint(packed, uint64(v))
	}
	b = protowire.AppendTag(b, fieldNum, protowire.BytesType)
	b = protowire.AppendBytes(b, packed)
	return b
}

// appendMessage appends a sub-message at the given field number
func appendMessage(b []byte, fieldNum protowire.Number, msg []byte) []byte {
	b = protowire.AppendTag(b, fieldNum, protowire.BytesType)
	b = protowire.AppendBytes(b, msg)
	return b
}

// buildField1Inner builds the deeply nested field1.field1.field1 structure (media field masks)
func buildField1Inner() []byte {
	var b []byte
	// Fields 1, 3, 4, 6 are empty messages
	b = appendEmptyMessage(b, 1)
	b = appendEmptyMessage(b, 3)
	b = appendEmptyMessage(b, 4)

	// Field 5 has nested structure
	var f5 []byte
	f5 = appendEmptyMessage(f5, 1)
	f5 = appendEmptyMessage(f5, 2)
	f5 = appendEmptyMessage(f5, 3)
	f5 = appendEmptyMessage(f5, 4)
	f5 = appendEmptyMessage(f5, 5)
	f5 = appendEmptyMessage(f5, 7)
	b = appendMessage(b, 5, f5)

	b = appendEmptyMessage(b, 6)

	// Field 7 has field2 inside
	var f7 []byte
	f7 = appendEmptyMessage(f7, 2)
	b = appendMessage(b, 7, f7)

	// Empty messages for many fields
	b = appendEmptyMessage(b, 15)
	b = appendEmptyMessage(b, 16)
	b = appendEmptyMessage(b, 17)
	b = appendEmptyMessage(b, 19)
	b = appendEmptyMessage(b, 20)

	// Field 21 has nested structure
	var f21 []byte
	var f21_5 []byte
	f21_5 = appendEmptyMessage(f21_5, 3)
	f21 = appendMessage(f21, 5, f21_5)
	f21 = appendEmptyMessage(f21, 6)
	b = appendMessage(b, 21, f21)

	b = appendEmptyMessage(b, 25)

	// Field 30 has field2 inside
	var f30 []byte
	f30 = appendEmptyMessage(f30, 2)
	b = appendMessage(b, 30, f30)

	b = appendEmptyMessage(b, 31)
	b = appendEmptyMessage(b, 32)

	// Field 33 has field1 inside
	var f33 []byte
	f33 = appendEmptyMessage(f33, 1)
	b = appendMessage(b, 33, f33)

	b = appendEmptyMessage(b, 34)
	b = appendEmptyMessage(b, 36)
	b = appendEmptyMessage(b, 37)
	b = appendEmptyMessage(b, 38)
	b = appendEmptyMessage(b, 39)
	b = appendEmptyMessage(b, 40)
	b = appendEmptyMessage(b, 41)

	return b
}

// buildField5Type builds the field5 structure within field1.field1
func buildField5Type() []byte {
	var b []byte

	// Field 2
	var f2 []byte
	var f2_2 []byte
	var f2_2_3 []byte
	f2_2_3 = appendEmptyMessage(f2_2_3, 2)
	f2_2 = appendMessage(f2_2, 3, f2_2_3)
	var f2_2_4 []byte
	f2_2_4 = appendEmptyMessage(f2_2_4, 2)
	f2_2_4 = appendEmptyMessage(f2_2_4, 4)
	f2_2 = appendMessage(f2_2, 4, f2_2_4)
	f2 = appendMessage(f2, 2, f2_2)
	var f2_4 []byte
	var f2_4_2 []byte
	f2_4_2 = appendVarintField(f2_4_2, 2, 1)
	f2_4 = appendMessage(f2_4, 2, f2_4_2)
	f2 = appendMessage(f2, 4, f2_4)
	var f2_5 []byte
	f2_5 = appendEmptyMessage(f2_5, 2)
	f2 = appendMessage(f2, 5, f2_5)
	f2 = appendVarintField(f2, 6, 1)
	b = appendMessage(b, 2, f2)

	// Field 3
	var f3 []byte
	var f3_2 []byte
	f3_2 = appendEmptyMessage(f3_2, 3)
	f3_2 = appendEmptyMessage(f3_2, 4)
	f3 = appendMessage(f3, 2, f3_2)
	var f3_3 []byte
	f3_3 = appendEmptyMessage(f3_3, 2)
	var f3_3_3 []byte
	f3_3_3 = appendVarintField(f3_3_3, 2, 1)
	f3_3 = appendMessage(f3_3, 3, f3_3_3)
	f3 = appendMessage(f3, 3, f3_3)
	f3 = appendEmptyMessage(f3, 4)
	var f3_5 []byte
	var f3_5_2 []byte
	f3_5_2 = appendVarintField(f3_5_2, 2, 1)
	f3_5 = appendMessage(f3_5, 2, f3_5_2)
	f3 = appendMessage(f3, 5, f3_5)
	f3 = appendEmptyMessage(f3, 7)
	b = appendMessage(b, 3, f3)

	// Field 4
	var f4 []byte
	var f4_2 []byte
	f4_2 = appendEmptyMessage(f4_2, 2)
	f4 = appendMessage(f4, 2, f4_2)
	b = appendMessage(b, 4, f4)

	// Field 5
	var f5 []byte
	var f5_1 []byte
	var f5_1_2 []byte
	f5_1_2 = appendEmptyMessage(f5_1_2, 3)
	f5_1_2 = appendEmptyMessage(f5_1_2, 4)
	f5_1 = appendMessage(f5_1, 2, f5_1_2)
	var f5_1_3 []byte
	f5_1_3 = appendEmptyMessage(f5_1_3, 2)
	var f5_1_3_3 []byte
	f5_1_3_3 = appendVarintField(f5_1_3_3, 2, 1)
	f5_1_3 = appendMessage(f5_1_3, 3, f5_1_3_3)
	f5_1 = appendMessage(f5_1, 3, f5_1_3)
	f5 = appendMessage(f5, 1, f5_1)
	f5 = appendVarintField(f5, 3, 1)
	b = appendMessage(b, 5, f5)

	return b
}

// buildField9Type builds the complex field9 structure
func buildField9Type() []byte {
	var b []byte
	b = appendEmptyMessage(b, 2)
	var f3 []byte
	f3 = appendEmptyMessage(f3, 1)
	f3 = appendEmptyMessage(f3, 2)
	b = appendMessage(b, 3, f3)

	// Field 4 has deeply nested structure
	var f4 []byte
	var f4_1 []byte
	var f4_1_3 []byte
	var f4_1_3_1 []byte
	var f4_1_3_1_1 []byte
	var f4_1_3_1_1_5 []byte
	f4_1_3_1_1_5 = appendEmptyMessage(f4_1_3_1_1_5, 1)
	f4_1_3_1_1 = appendMessage(f4_1_3_1_1, 5, f4_1_3_1_1_5)
	f4_1_3_1_1 = appendEmptyMessage(f4_1_3_1_1, 6)
	f4_1_3_1_1 = appendEmptyMessage(f4_1_3_1_1, 7)
	f4_1_3_1 = appendMessage(f4_1_3_1, 1, f4_1_3_1_1)
	f4_1_3_1 = appendEmptyMessage(f4_1_3_1, 2)
	var f4_1_3_1_3 []byte
	var f4_1_3_1_3_1 []byte
	var f4_1_3_1_3_1_5 []byte
	f4_1_3_1_3_1_5 = appendEmptyMessage(f4_1_3_1_3_1_5, 1)
	f4_1_3_1_3_1 = appendMessage(f4_1_3_1_3_1, 5, f4_1_3_1_3_1_5)
	f4_1_3_1_3_1 = appendEmptyMessage(f4_1_3_1_3_1, 6)
	f4_1_3_1_3_1 = appendEmptyMessage(f4_1_3_1_3_1, 7)
	f4_1_3_1_3 = appendMessage(f4_1_3_1_3, 1, f4_1_3_1_3_1)
	f4_1_3_1_3 = appendEmptyMessage(f4_1_3_1_3, 2)
	f4_1_3_1 = appendMessage(f4_1_3_1, 3, f4_1_3_1_3)
	f4_1_3 = appendMessage(f4_1_3, 1, f4_1_3_1)
	f4_1 = appendMessage(f4_1, 3, f4_1_3)
	var f4_1_4 []byte
	var f4_1_4_1 []byte
	f4_1_4_1 = appendEmptyMessage(f4_1_4_1, 2)
	f4_1_4 = appendMessage(f4_1_4, 1, f4_1_4_1)
	f4_1 = appendMessage(f4_1, 4, f4_1_4)
	f4 = appendMessage(f4, 1, f4_1)
	b = appendMessage(b, 4, f4)

	return b
}

// buildField1Type builds the entire field1 -> field1 structure
func buildField1Type() []byte {
	var b []byte

	// field1 (media masks)
	b = appendMessage(b, 1, buildField1Inner())

	// field5
	b = appendMessage(b, 5, buildField5Type())

	// field8 empty
	b = appendEmptyMessage(b, 8)

	// field9
	b = appendMessage(b, 9, buildField9Type())

	// field11
	var f11 []byte
	f11 = appendEmptyMessage(f11, 2)
	f11 = appendEmptyMessage(f11, 3)
	var f11_4 []byte
	var f11_4_2 []byte
	f11_4_2 = appendVarintField(f11_4_2, 1, 1)
	f11_4_2 = appendVarintField(f11_4_2, 2, 2)
	f11_4 = appendMessage(f11_4, 2, f11_4_2)
	f11 = appendMessage(f11, 4, f11_4)
	b = appendMessage(b, 11, f11)

	// field12 empty
	b = appendEmptyMessage(b, 12)

	// field14 (same as field11)
	var f14 []byte
	f14 = appendEmptyMessage(f14, 2)
	f14 = appendEmptyMessage(f14, 3)
	var f14_4 []byte
	var f14_4_2 []byte
	f14_4_2 = appendVarintField(f14_4_2, 1, 1)
	f14_4_2 = appendVarintField(f14_4_2, 2, 2)
	f14_4 = appendMessage(f14_4, 2, f14_4_2)
	f14 = appendMessage(f14, 4, f14_4)
	b = appendMessage(b, 14, f14)

	// field15
	var f15 []byte
	f15 = appendEmptyMessage(f15, 1)
	f15 = appendEmptyMessage(f15, 4)
	b = appendMessage(b, 15, f15)

	// field17 (same as field15)
	var f17 []byte
	f17 = appendEmptyMessage(f17, 1)
	f17 = appendEmptyMessage(f17, 4)
	b = appendMessage(b, 17, f17)

	// field19 (same as field11)
	var f19 []byte
	f19 = appendEmptyMessage(f19, 2)
	f19 = appendEmptyMessage(f19, 3)
	var f19_4 []byte
	var f19_4_2 []byte
	f19_4_2 = appendVarintField(f19_4_2, 1, 1)
	f19_4_2 = appendVarintField(f19_4_2, 2, 2)
	f19_4 = appendMessage(f19_4, 2, f19_4_2)
	f19 = appendMessage(f19, 4, f19_4)
	b = appendMessage(b, 19, f19)

	// field21 (only for get_library_state, has field1)
	var f21 []byte
	f21 = appendEmptyMessage(f21, 1)
	b = appendMessage(b, 21, f21)

	// field22 empty
	b = appendEmptyMessage(b, 22)

	// field23 empty
	b = appendEmptyMessage(b, 23)

	return b
}

// buildField2Type builds the album/collection field masks
func buildField2Type() []byte {
	var b []byte

	// field1
	var f1 []byte
	f1 = appendEmptyMessage(f1, 2)
	f1 = appendEmptyMessage(f1, 3)
	f1 = appendEmptyMessage(f1, 4)
	f1 = appendEmptyMessage(f1, 5)
	var f1_6 []byte
	f1_6 = appendEmptyMessage(f1_6, 1)
	f1_6 = appendEmptyMessage(f1_6, 2)
	f1_6 = appendEmptyMessage(f1_6, 3)
	f1_6 = appendEmptyMessage(f1_6, 4)
	f1_6 = appendEmptyMessage(f1_6, 5)
	f1_6 = appendEmptyMessage(f1_6, 7)
	f1 = appendMessage(f1, 6, f1_6)
	f1 = appendEmptyMessage(f1, 7)
	f1 = appendEmptyMessage(f1, 8)
	f1 = appendEmptyMessage(f1, 10)
	f1 = appendEmptyMessage(f1, 12)
	var f1_13 []byte
	f1_13 = appendEmptyMessage(f1_13, 2)
	f1_13 = appendEmptyMessage(f1_13, 3)
	f1 = appendMessage(f1, 13, f1_13)
	var f1_15 []byte
	f1_15 = appendEmptyMessage(f1_15, 1)
	f1 = appendMessage(f1, 15, f1_15)
	f1 = appendEmptyMessage(f1, 18)
	b = appendMessage(b, 1, f1)

	// field4
	var f4 []byte
	f4 = appendEmptyMessage(f4, 1)
	b = appendMessage(b, 4, f4)

	// field9 empty
	b = appendEmptyMessage(b, 9)

	// field11
	var f11 []byte
	var f11_1 []byte
	f11_1 = appendEmptyMessage(f11_1, 1)
	f11_1 = appendEmptyMessage(f11_1, 4)
	f11_1 = appendEmptyMessage(f11_1, 5)
	f11_1 = appendEmptyMessage(f11_1, 6)
	f11_1 = appendEmptyMessage(f11_1, 9)
	f11 = appendMessage(f11, 1, f11_1)
	b = appendMessage(b, 11, f11)

	// field14 complex nested structure - simplified for now
	var f14 []byte
	var f14_1 []byte
	var f14_1_1 []byte
	f14_1_1 = appendEmptyMessage(f14_1_1, 1)
	var f14_1_1_2 []byte
	var f14_1_1_2_2 []byte
	var f14_1_1_2_2_1 []byte
	f14_1_1_2_2_1 = appendEmptyMessage(f14_1_1_2_2_1, 1)
	f14_1_1_2_2 = appendMessage(f14_1_1_2_2, 1, f14_1_1_2_2_1)
	f14_1_1_2_2 = appendEmptyMessage(f14_1_1_2_2, 3)
	f14_1_1_2 = appendMessage(f14_1_1_2, 2, f14_1_1_2_2)
	f14_1_1 = appendMessage(f14_1_1, 2, f14_1_1_2)
	var f14_1_1_3 []byte
	var f14_1_1_3_4 []byte
	var f14_1_1_3_4_1 []byte
	f14_1_1_3_4_1 = appendEmptyMessage(f14_1_1_3_4_1, 1)
	f14_1_1_3_4 = appendMessage(f14_1_1_3_4, 1, f14_1_1_3_4_1)
	f14_1_1_3_4 = appendEmptyMessage(f14_1_1_3_4, 3)
	f14_1_1_3 = appendMessage(f14_1_1_3, 4, f14_1_1_3_4)
	var f14_1_1_3_5 []byte
	var f14_1_1_3_5_1 []byte
	f14_1_1_3_5_1 = appendEmptyMessage(f14_1_1_3_5_1, 1)
	f14_1_1_3_5 = appendMessage(f14_1_1_3_5, 1, f14_1_1_3_5_1)
	f14_1_1_3_5 = appendEmptyMessage(f14_1_1_3_5, 3)
	f14_1_1_3 = appendMessage(f14_1_1_3, 5, f14_1_1_3_5)
	f14_1_1 = appendMessage(f14_1_1, 3, f14_1_1_3)
	f14_1 = appendMessage(f14_1, 1, f14_1_1)
	f14_1 = appendEmptyMessage(f14_1, 2)
	f14 = appendMessage(f14, 1, f14_1)
	b = appendMessage(b, 14, f14)

	// field17 empty
	b = appendEmptyMessage(b, 17)

	// field18
	var f18 []byte
	f18 = appendEmptyMessage(f18, 1)
	var f18_2 []byte
	f18_2 = appendEmptyMessage(f18_2, 1)
	f18 = appendMessage(f18, 2, f18_2)
	b = appendMessage(b, 18, f18)

	// field20
	var f20 []byte
	var f20_2 []byte
	f20_2 = appendEmptyMessage(f20_2, 1)
	f20_2 = appendEmptyMessage(f20_2, 2)
	f20 = appendMessage(f20, 2, f20_2)
	b = appendMessage(b, 20, f20)

	// field22, 23, 24 empty
	b = appendEmptyMessage(b, 22)
	b = appendEmptyMessage(b, 23)
	b = appendEmptyMessage(b, 24)

	return b
}

// buildField3Type builds the settings/system field masks
func buildField3Type() []byte {
	var b []byte

	// field2 empty
	b = appendEmptyMessage(b, 2)

	// field3 with many nested fields
	var f3 []byte
	f3 = appendEmptyMessage(f3, 2)
	f3 = appendEmptyMessage(f3, 3)
	f3 = appendEmptyMessage(f3, 7)
	f3 = appendEmptyMessage(f3, 8)
	var f3_14 []byte
	f3_14 = appendEmptyMessage(f3_14, 1)
	f3 = appendMessage(f3, 14, f3_14)
	f3 = appendEmptyMessage(f3, 16)
	var f3_17 []byte
	f3_17 = appendEmptyMessage(f3_17, 2)
	f3 = appendMessage(f3, 17, f3_17)
	f3 = appendEmptyMessage(f3, 18)
	f3 = appendEmptyMessage(f3, 19)
	f3 = appendEmptyMessage(f3, 20)
	f3 = appendEmptyMessage(f3, 21)
	f3 = appendEmptyMessage(f3, 22)
	f3 = appendEmptyMessage(f3, 23)
	var f3_27 []byte
	f3_27 = appendEmptyMessage(f3_27, 1)
	var f3_27_2 []byte
	f3_27_2 = appendEmptyMessage(f3_27_2, 1)
	f3_27 = appendMessage(f3_27, 2, f3_27_2)
	f3 = appendMessage(f3, 27, f3_27)
	f3 = appendEmptyMessage(f3, 29)
	f3 = appendEmptyMessage(f3, 30)
	f3 = appendEmptyMessage(f3, 31)
	f3 = appendEmptyMessage(f3, 32)
	f3 = appendEmptyMessage(f3, 34)
	f3 = appendEmptyMessage(f3, 37)
	f3 = appendEmptyMessage(f3, 38)
	f3 = appendEmptyMessage(f3, 39)
	f3 = appendEmptyMessage(f3, 41)
	var f3_43 []byte
	f3_43 = appendEmptyMessage(f3_43, 1)
	f3 = appendMessage(f3, 43, f3_43)
	var f3_45 []byte
	var f3_45_1 []byte
	f3_45_1 = appendEmptyMessage(f3_45_1, 1)
	f3_45 = appendMessage(f3_45, 1, f3_45_1)
	f3 = appendMessage(f3, 45, f3_45)
	var f3_46 []byte
	f3_46 = appendEmptyMessage(f3_46, 1)
	f3_46 = appendEmptyMessage(f3_46, 2)
	f3_46 = appendEmptyMessage(f3_46, 3)
	f3 = appendMessage(f3, 46, f3_46)
	f3 = appendEmptyMessage(f3, 47)
	b = appendMessage(b, 3, f3)

	// field4
	var f4 []byte
	f4 = appendEmptyMessage(f4, 2)
	var f4_3 []byte
	f4_3 = appendEmptyMessage(f4_3, 1)
	f4 = appendMessage(f4, 3, f4_3)
	f4 = appendEmptyMessage(f4, 4)
	var f4_5 []byte
	f4_5 = appendEmptyMessage(f4_5, 1)
	f4 = appendMessage(f4, 5, f4_5)
	b = appendMessage(b, 4, f4)

	// field7 empty
	b = appendEmptyMessage(b, 7)

	// field12, 13 empty
	b = appendEmptyMessage(b, 12)
	b = appendEmptyMessage(b, 13)

	// field14 with complex structure
	var f14 []byte
	f14 = appendEmptyMessage(f14, 1)
	var f14_2 []byte
	f14_2 = appendEmptyMessage(f14_2, 1)
	var f14_2_2 []byte
	f14_2_2 = appendEmptyMessage(f14_2_2, 1)
	f14_2 = appendMessage(f14_2, 2, f14_2_2)
	f14_2 = appendEmptyMessage(f14_2, 3)
	var f14_2_4 []byte
	f14_2_4 = appendEmptyMessage(f14_2_4, 1)
	f14_2 = appendMessage(f14_2, 4, f14_2_4)
	f14 = appendMessage(f14, 2, f14_2)
	var f14_3 []byte
	f14_3 = appendEmptyMessage(f14_3, 1)
	var f14_3_2 []byte
	f14_3_2 = appendEmptyMessage(f14_3_2, 1)
	f14_3 = appendMessage(f14_3, 2, f14_3_2)
	f14_3 = appendEmptyMessage(f14_3, 3)
	f14_3 = appendEmptyMessage(f14_3, 4)
	f14 = appendMessage(f14, 3, f14_3)
	b = appendMessage(b, 14, f14)

	// field15 empty
	b = appendEmptyMessage(b, 15)

	// field16
	var f16 []byte
	f16 = appendEmptyMessage(f16, 1)
	b = appendMessage(b, 16, f16)

	// field18 empty
	b = appendEmptyMessage(b, 18)

	// field19
	var f19 []byte
	var f19_4 []byte
	f19_4 = appendEmptyMessage(f19_4, 2)
	f19 = appendMessage(f19, 4, f19_4)
	var f19_6 []byte
	f19_6 = appendEmptyMessage(f19_6, 2)
	f19_6 = appendEmptyMessage(f19_6, 3)
	f19 = appendMessage(f19, 6, f19_6)
	var f19_7 []byte
	f19_7 = appendEmptyMessage(f19_7, 2)
	f19_7 = appendEmptyMessage(f19_7, 3)
	f19 = appendMessage(f19, 7, f19_7)
	f19 = appendEmptyMessage(f19, 8)
	f19 = appendEmptyMessage(f19, 9)
	b = appendMessage(b, 19, f19)

	// field20, 22, 24, 25, 26 empty
	b = appendEmptyMessage(b, 20)
	b = appendEmptyMessage(b, 22)
	b = appendEmptyMessage(b, 24)
	b = appendEmptyMessage(b, 25)
	b = appendEmptyMessage(b, 26)

	return b
}

// buildField9 builds the field9 for main request
func buildField9() []byte {
	var b []byte

	// field1
	var f1 []byte
	var f1_2 []byte
	f1_2 = appendEmptyMessage(f1_2, 1)
	f1_2 = appendEmptyMessage(f1_2, 2)
	f1 = appendMessage(f1, 2, f1_2)
	b = appendMessage(b, 1, f1)

	// field2
	var f2 []byte
	var f2_3 []byte
	f2_3 = appendVarintField(f2_3, 2, 1)
	f2 = appendMessage(f2, 3, f2_3)
	b = appendMessage(b, 2, f2)

	// field3
	var f3 []byte
	f3 = appendEmptyMessage(f3, 2)
	b = appendMessage(b, 3, f3)

	// field4 empty
	b = appendEmptyMessage(b, 4)

	// field7
	var f7 []byte
	f7 = appendEmptyMessage(f7, 1)
	b = appendMessage(b, 7, f7)

	// field8
	var f8 []byte
	f8 = appendVarintField(f8, 1, 2)
	f8 = appendBytesField(f8, 2, []byte{0x01, 0x02, 0x03, 0x05, 0x06, 0x07})
	b = appendMessage(b, 8, f8)

	// field9 empty
	b = appendEmptyMessage(b, 9)

	// field11
	var f11 []byte
	f11 = appendEmptyMessage(f11, 1)
	b = appendMessage(b, 11, f11)

	return b
}

// buildField18Value builds the field18 map value
func buildField18Value() []byte {
	var inner []byte
	inner = appendRepeatedVarint(inner, 4, []int64{2, 1, 6, 8, 10, 15, 18, 13, 17, 19, 14, 20})
	inner = appendVarintField(inner, 5, 6)
	inner = appendVarintField(inner, 6, 2)
	inner = appendVarintField(inner, 7, 1)
	inner = appendVarintField(inner, 8, 2)
	inner = appendVarintField(inner, 11, 3)
	inner = appendVarintField(inner, 12, 1)
	inner = appendVarintField(inner, 13, 3)
	inner = appendVarintField(inner, 15, 1)
	inner = appendVarintField(inner, 16, 1)
	inner = appendVarintField(inner, 17, 1)
	inner = appendVarintField(inner, 18, 2)

	var f1Inner []byte
	f1Inner = appendMessage(f1Inner, 1, inner)

	var f1 []byte
	f1 = appendMessage(f1, 1, f1Inner)

	return f1
}

// buildField19 builds the field19 structure
func buildField19() []byte {
	var b []byte

	// field1
	var f1 []byte
	f1 = appendEmptyMessage(f1, 1)
	f1 = appendEmptyMessage(f1, 2)
	b = appendMessage(b, 1, f1)

	// field2
	var f2 []byte
	f2 = appendRepeatedVarint(f2, 1, []int64{1, 2, 4, 6, 5, 7})
	b = appendMessage(b, 2, f2)

	// field3
	var f3 []byte
	f3 = appendEmptyMessage(f3, 1)
	f3 = appendEmptyMessage(f3, 2)
	b = appendMessage(b, 3, f3)

	// field5
	var f5 []byte
	f5 = appendEmptyMessage(f5, 1)
	f5 = appendEmptyMessage(f5, 2)
	b = appendMessage(b, 5, f5)

	// field6
	var f6 []byte
	f6 = appendEmptyMessage(f6, 1)
	b = appendMessage(b, 6, f6)

	// field7
	var f7 []byte
	f7 = appendEmptyMessage(f7, 1)
	f7 = appendEmptyMessage(f7, 2)
	b = appendMessage(b, 7, f7)

	// field8
	var f8 []byte
	f8 = appendEmptyMessage(f8, 1)
	b = appendMessage(b, 8, f8)

	return b
}

// buildField20 builds the field20 structure (printing promotion sync options)
func buildField20(includeField2 bool) []byte {
	var b []byte
	b = appendVarintField(b, 1, 1)
	if includeField2 {
		b = appendStringField(b, 2, "")
	}

	// field3
	var f3 []byte
	f3 = appendStringField(f3, 1, "type.googleapis.com/photos.printing.client.PrintingPromotionSyncOptions")
	var f3_2 []byte
	var f3_2_1 []byte
	f3_2_1 = appendRepeatedVarint(f3_2_1, 4, []int64{2, 1, 6, 8, 10, 15, 18, 13, 17, 19, 14, 20})
	f3_2_1 = appendVarintField(f3_2_1, 5, 6)
	f3_2_1 = appendVarintField(f3_2_1, 6, 2)
	f3_2_1 = appendVarintField(f3_2_1, 7, 1)
	f3_2_1 = appendVarintField(f3_2_1, 8, 2)
	f3_2_1 = appendVarintField(f3_2_1, 11, 3)
	f3_2_1 = appendVarintField(f3_2_1, 12, 1)
	f3_2_1 = appendVarintField(f3_2_1, 13, 3)
	f3_2_1 = appendVarintField(f3_2_1, 15, 1)
	f3_2_1 = appendVarintField(f3_2_1, 16, 1)
	f3_2_1 = appendVarintField(f3_2_1, 17, 1)
	f3_2_1 = appendVarintField(f3_2_1, 18, 2)
	f3_2 = appendMessage(f3_2, 1, f3_2_1)
	f3 = appendMessage(f3, 2, f3_2)
	b = appendMessage(b, 3, f3)

	return b
}

// buildField21 builds the field21 structure
func buildField21(forGetState bool) []byte {
	var b []byte

	// field2
	var f2 []byte
	var f2_2 []byte
	f2_2 = appendEmptyMessage(f2_2, 4)
	f2 = appendMessage(f2, 2, f2_2)
	f2 = appendEmptyMessage(f2, 4)
	f2 = appendEmptyMessage(f2, 5)
	b = appendMessage(b, 2, f2)

	// field3
	var f3 []byte
	var f3_2 []byte
	f3_2 = appendVarintField(f3_2, 1, 1)
	f3 = appendMessage(f3, 2, f3_2)
	if forGetState {
		var f3_4 []byte
		f3_4 = appendEmptyMessage(f3_4, 2)
		f3 = appendMessage(f3, 4, f3_4)
	}
	b = appendMessage(b, 3, f3)

	// field5
	var f5 []byte
	f5 = appendEmptyMessage(f5, 1)
	b = appendMessage(b, 5, f5)

	// field6
	var f6 []byte
	f6 = appendEmptyMessage(f6, 1)
	var f6_2 []byte
	f6_2 = appendEmptyMessage(f6_2, 1)
	f6 = appendMessage(f6, 2, f6_2)
	b = appendMessage(b, 6, f6)

	// field7
	var f7 []byte
	f7 = appendVarintField(f7, 1, 2)
	if forGetState {
		f7 = appendBytesField(f7, 2, []byte("\x01\x07\x08\x09\x0a\x0d\x0e\x0f\x11\x13\x14\x16\x17-./01:\x06\x18267;>?@A89<GBED"))
	} else {
		f7 = appendBytesField(f7, 2, []byte("\x01\x07\x08\x09\x0a\x0d\x0e\x0f\x11\x13\x14\x16\x17-./01:\x06\x18267;>?@A89<"))
	}
	f7 = appendBytesField(f7, 3, []byte{0x01})
	b = appendMessage(b, 7, f7)

	// field8
	var f8 []byte
	var f8_3 []byte
	var f8_3_1 []byte
	var f8_3_1_1 []byte
	var f8_3_1_1_2 []byte
	f8_3_1_1_2 = appendVarintField(f8_3_1_1_2, 1, 1)
	f8_3_1_1 = appendMessage(f8_3_1_1, 2, f8_3_1_1_2)
	if forGetState {
		var f8_3_1_1_4 []byte
		f8_3_1_1_4 = appendEmptyMessage(f8_3_1_1_4, 2)
		f8_3_1_1 = appendMessage(f8_3_1_1, 4, f8_3_1_1_4)
	}
	f8_3_1 = appendMessage(f8_3_1, 1, f8_3_1_1)
	if forGetState {
		f8_3 = appendEmptyMessage(f8_3, 3)
	}
	f8_3 = appendMessage(f8_3, 1, f8_3_1)
	f8 = appendMessage(f8, 3, f8_3)
	var f8_4 []byte
	f8_4 = appendEmptyMessage(f8_4, 1)
	f8 = appendMessage(f8, 4, f8_4)
	if forGetState {
		var f8_5 []byte
		var f8_5_1 []byte
		var f8_5_1_2 []byte
		f8_5_1_2 = appendVarintField(f8_5_1_2, 1, 1)
		f8_5_1 = appendMessage(f8_5_1, 2, f8_5_1_2)
		var f8_5_1_4 []byte
		f8_5_1_4 = appendEmptyMessage(f8_5_1_4, 2)
		f8_5_1 = appendMessage(f8_5_1, 4, f8_5_1_4)
		f8_5 = appendMessage(f8_5, 1, f8_5_1)
		f8 = appendMessage(f8, 5, f8_5)
	}
	b = appendMessage(b, 8, f8)

	// field9
	var f9 []byte
	f9 = appendEmptyMessage(f9, 1)
	b = appendMessage(b, 9, f9)

	// field10
	var f10 []byte
	var f10_1 []byte
	f10_1 = appendEmptyMessage(f10_1, 1)
	f10 = appendMessage(f10, 1, f10_1)
	f10 = appendEmptyMessage(f10, 3)
	f10 = appendEmptyMessage(f10, 5)
	var f10_6 []byte
	f10_6 = appendEmptyMessage(f10_6, 1)
	f10 = appendMessage(f10, 6, f10_6)
	f10 = appendEmptyMessage(f10, 7)
	f10 = appendEmptyMessage(f10, 9)
	f10 = appendEmptyMessage(f10, 10)
	b = appendMessage(b, 10, f10)

	// field11, 12, 13 empty
	b = appendEmptyMessage(b, 11)
	b = appendEmptyMessage(b, 12)
	b = appendEmptyMessage(b, 13)

	if forGetState {
		// field14 empty
		b = appendEmptyMessage(b, 14)
	}

	// field16
	var f16 []byte
	f16 = appendEmptyMessage(f16, 1)
	b = appendMessage(b, 16, f16)

	return b
}

// buildField25 builds the field25 structure
func buildField25() []byte {
	var b []byte

	// field1
	var f1 []byte
	var f1_1 []byte
	var f1_1_1 []byte
	f1_1_1 = appendEmptyMessage(f1_1_1, 1)
	f1_1 = appendMessage(f1_1, 1, f1_1_1)
	f1 = appendMessage(f1, 1, f1_1)
	b = appendMessage(b, 1, f1)

	// field2 empty
	b = appendEmptyMessage(b, 2)

	return b
}

// buildMainField1 builds the entire field1 structure for the main request
func buildMainField1(stateToken string, pageToken string, forGetState bool, includeField24 bool) []byte {
	var b []byte

	// field1 (media field masks)
	b = appendMessage(b, 1, buildField1Type())

	// field2 (album field masks)
	b = appendMessage(b, 2, buildField2Type())

	// field3 (settings field masks)
	b = appendMessage(b, 3, buildField3Type())

	// field4 (page token) - only for page requests
	if pageToken != "" {
		b = appendStringField(b, 4, pageToken)
	}

	// field6 (state token)
	if stateToken != "" {
		b = appendStringField(b, 6, stateToken)
	}

	// field7
	b = appendVarintField(b, 7, 2)

	// field9
	b = appendMessage(b, 9, buildField9())

	// field11 - repeated varints
	if forGetState {
		b = appendRepeatedVarint(b, 11, []int64{1, 2, 6})
	} else {
		b = appendRepeatedVarint(b, 11, []int64{1, 2})
	}

	// field12
	var f12 []byte
	var f12_2 []byte
	f12_2 = appendEmptyMessage(f12_2, 1)
	f12_2 = appendEmptyMessage(f12_2, 2)
	f12 = appendMessage(f12, 2, f12_2)
	var f12_3 []byte
	f12_3 = appendEmptyMessage(f12_3, 1)
	f12 = appendMessage(f12, 3, f12_3)
	f12 = appendEmptyMessage(f12, 4)
	b = appendMessage(b, 12, f12)

	// field13 empty
	b = appendEmptyMessage(b, 13)

	// field15
	var f15 []byte
	var f15_3 []byte
	f15_3 = appendVarintField(f15_3, 1, 1)
	f15 = appendMessage(f15, 3, f15_3)
	b = appendMessage(b, 15, f15)

	// field18 - map entry with key 169945741
	var f18Entry []byte
	f18Entry = appendVarintField(f18Entry, 1, 169945741)
	f18Entry = appendMessage(f18Entry, 2, buildField18Value())
	b = appendMessage(b, 18, f18Entry)

	// field19
	b = appendMessage(b, 19, buildField19())

	// field20
	b = appendMessage(b, 20, buildField20(forGetState))

	// field21
	b = appendMessage(b, 21, buildField21(forGetState))

	// field22
	var f22 []byte
	if forGetState {
		f22 = appendVarintField(f22, 1, 1)
		f22 = appendStringField(f22, 2, "107818234414673686888")
	} else {
		f22 = appendVarintField(f22, 1, 2)
	}
	b = appendMessage(b, 22, f22)

	if includeField24 {
		b = appendEmptyMessage(b, 24)
	}

	// field25
	b = appendMessage(b, 25, buildField25())

	if forGetState {
		// field26 empty
		b = appendEmptyMessage(b, 26)
	}

	return b
}

// buildField2Outer builds the outer field2 structure
func buildField2Outer() []byte {
	var b []byte

	// field1
	var f1 []byte
	var f1_1 []byte
	var f1_1_1 []byte
	f1_1_1 = appendEmptyMessage(f1_1_1, 1)
	f1_1 = appendMessage(f1_1, 1, f1_1_1)
	f1_1 = appendEmptyMessage(f1_1, 2)
	f1 = appendMessage(f1, 1, f1_1)
	f1 = appendEmptyMessage(f1, 2)
	b = appendMessage(b, 1, f1)

	// field2 empty
	b = appendEmptyMessage(b, 2)

	return b
}

// decodeProtobufToJSON decodes raw protobuf bytes to a JSON-compatible structure
func decodeProtobufToJSON(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}

	result := make(map[string]any)

	for len(data) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(data)
		if n < 0 {
			return nil, fmt.Errorf("failed to consume tag")
		}
		data = data[n:]

		var value any
		var consumed int

		switch wireType {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return nil, fmt.Errorf("failed to consume varint")
			}
			value = v
			consumed = n

		case protowire.Fixed64Type:
			v, n := protowire.ConsumeFixed64(data)
			if n < 0 {
				return nil, fmt.Errorf("failed to consume fixed64")
			}
			value = v
			consumed = n

		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return nil, fmt.Errorf("failed to consume bytes")
			}
			consumed = n

			// Try to decode as nested message
			nested, err := decodeProtobufToJSON(v)
			if err == nil && nested != nil {
				value = nested
			} else {
				// Check if it's valid UTF-8 string
				if isValidUTF8(v) {
					value = string(v)
				} else {
					// Store as base64
					value = base64.StdEncoding.EncodeToString(v)
				}
			}

		case protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(data)
			if n < 0 {
				return nil, fmt.Errorf("failed to consume fixed32")
			}
			value = v
			consumed = n

		case protowire.StartGroupType, protowire.EndGroupType:
			// Skip groups (deprecated)
			return nil, fmt.Errorf("groups not supported")

		default:
			return nil, fmt.Errorf("unknown wire type: %d", wireType)
		}

		data = data[consumed:]

		// Add to result, handling repeated fields
		key := fmt.Sprintf("%d", fieldNum)
		if existing, ok := result[key]; ok {
			// Convert to array if not already
			switch v := existing.(type) {
			case []any:
				result[key] = append(v, value)
			default:
				result[key] = []any{v, value}
			}
		} else {
			result[key] = value
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

// isValidUTF8 checks if bytes form valid UTF-8 and contain printable characters
func isValidUTF8(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	// Check for valid UTF-8
	for i := 0; i < len(data); {
		r, size := decodeRune(data[i:])
		if r == 0xFFFD && size == 1 {
			return false
		}
		// Check if it's a printable character or common whitespace
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
		i += size
	}
	return true
}

// decodeRune decodes a single UTF-8 rune from bytes
func decodeRune(data []byte) (rune, int) {
	if len(data) == 0 {
		return 0xFFFD, 0
	}
	b := data[0]
	if b < 0x80 {
		return rune(b), 1
	}
	if b < 0xC0 {
		return 0xFFFD, 1
	}
	if b < 0xE0 {
		if len(data) < 2 {
			return 0xFFFD, 1
		}
		return rune(b&0x1F)<<6 | rune(data[1]&0x3F), 2
	}
	if b < 0xF0 {
		if len(data) < 3 {
			return 0xFFFD, 1
		}
		return rune(b&0x0F)<<12 | rune(data[1]&0x3F)<<6 | rune(data[2]&0x3F), 3
	}
	if len(data) < 4 {
		return 0xFFFD, 1
	}
	return rune(b&0x07)<<18 | rune(data[1]&0x3F)<<12 | rune(data[2]&0x3F)<<6 | rune(data[3]&0x3F), 4
}

// writeJSONResponse decodes protobuf and writes as JSON to file
func writeJSONResponse(data []byte, outputFile string) error {
	decoded, err := decodeProtobufToJSON(data)
	if err != nil {
		return fmt.Errorf("failed to decode protobuf: %w", err)
	}

	jsonData, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// BuildGetLibraryStateRequest builds the request for GetLibraryState
func BuildGetLibraryStateRequest(stateToken string) []byte {
	var b []byte

	// field1 (main request body)
	b = appendMessage(b, 1, buildMainField1(stateToken, "", true, false))

	// field2 (outer wrapper)
	b = appendMessage(b, 2, buildField2Outer())

	return b
}

// BuildGetLibraryPageInitRequest builds the request for GetLibraryPageInit
func BuildGetLibraryPageInitRequest(pageToken string) []byte {
	var b []byte

	// field1 (main request body)
	b = appendMessage(b, 1, buildMainField1("", pageToken, false, false))

	// field2 (outer wrapper)
	b = appendMessage(b, 2, buildField2Outer())

	return b
}

// BuildGetLibraryPageRequest builds the request for GetLibraryPage
func BuildGetLibraryPageRequest(pageToken string, stateToken string) []byte {
	var b []byte

	// field1 (main request body)
	b = appendMessage(b, 1, buildMainField1(stateToken, pageToken, false, false))

	// field2 (outer wrapper)
	b = appendMessage(b, 2, buildField2Outer())

	return b
}

// GetLibraryState gets the library state
// Writes the response as JSON to the specified file
func (a *Api) GetLibraryState(ctx context.Context, stateToken string, outputFile string) error {
	requestBody := BuildGetLibraryStateRequest(stateToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return fmt.Errorf("GetLibraryState request failed: %w", err)
	}

	if outputFile != "" {
		if err := writeJSONResponse(bodyBytes, outputFile); err != nil {
			return fmt.Errorf("failed to write response to file: %w", err)
		}
		fmt.Printf("GetLibraryState response written to: %s\n", outputFile)
	}

	return nil
}

// GetLibraryPageInit gets the library page during initialization
// Writes the response as JSON to the specified file
func (a *Api) GetLibraryPageInit(ctx context.Context, pageToken string, outputFile string) error {
	requestBody := BuildGetLibraryPageInitRequest(pageToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return fmt.Errorf("GetLibraryPageInit request failed: %w", err)
	}

	if outputFile != "" {
		if err := writeJSONResponse(bodyBytes, outputFile); err != nil {
			return fmt.Errorf("failed to write response to file: %w", err)
		}
		fmt.Printf("GetLibraryPageInit response written to: %s\n", outputFile)
	}

	return nil
}

// GetLibraryPage gets the library page during regular update
// Writes the response as JSON to the specified file
func (a *Api) GetLibraryPage(ctx context.Context, pageToken string, stateToken string, outputFile string) error {
	requestBody := BuildGetLibraryPageRequest(pageToken, stateToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return fmt.Errorf("GetLibraryPage request failed: %w", err)
	}

	if outputFile != "" {
		if err := writeJSONResponse(bodyBytes, outputFile); err != nil {
			return fmt.Errorf("failed to write response to file: %w", err)
		}
		fmt.Printf("GetLibraryPage response written to: %s\n", outputFile)
	}

	return nil
}

// LibraryResponse represents the parsed library state response
type LibraryResponse struct {
	MediaItems []MediaItemInfo
	Albums     []AlbumInfo
	StateToken string
	PageToken  string
}

// MediaItemInfo contains displayable media item information
type MediaItemInfo struct {
	MediaKey          string
	Filename          string
	Caption           string
	CreationTimestamp int64
	FileSize          int64
	Width             int
	Height            int
	IsVideo           bool
	IsInTrash         bool
	TrashedAt         int64
	AlbumMediaKey     string
	DownloadURL       string
	ThumbnailURL      string
	DedupKey          string
}

// AlbumInfo contains displayable album information
type AlbumInfo struct {
	AlbumKey   string
	Name       string
	ItemCount  int
	CoverKey   string
}

// FetchLibraryStateRaw fetches the library state and returns raw JSON bytes
func (a *Api) FetchLibraryStateRaw(ctx context.Context, stateToken string) ([]byte, error) {
	requestBody := BuildGetLibraryStateRequest(stateToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return nil, fmt.Errorf("GetLibraryState request failed: %w", err)
	}

	decoded, err := decodeProtobufToJSON(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode protobuf: %w", err)
	}

	jsonData, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, nil
}

// FetchLibraryState fetches and parses the library state
func (a *Api) FetchLibraryState(ctx context.Context, stateToken string) (*LibraryResponse, error) {
	requestBody := BuildGetLibraryStateRequest(stateToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return nil, fmt.Errorf("GetLibraryState request failed: %w", err)
	}

	return parseLibraryResponse(bodyBytes)
}

// FetchLibraryPageInit fetches and parses library page init response
func (a *Api) FetchLibraryPageInit(ctx context.Context, pageToken string) (*LibraryResponse, error) {
	requestBody := BuildGetLibraryPageInitRequest(pageToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return nil, fmt.Errorf("GetLibraryPageInit request failed: %w", err)
	}

	return parseLibraryResponse(bodyBytes)
}

// FetchLibraryPage fetches and parses library page response
func (a *Api) FetchLibraryPage(ctx context.Context, pageToken, stateToken string) (*LibraryResponse, error) {
	requestBody := BuildGetLibraryPageRequest(pageToken, stateToken)

	bodyBytes, _, err := a.DoRequest(
		ctx,
		libraryStateEndpoint,
		bytes.NewReader(requestBody),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
	if err != nil {
		return nil, fmt.Errorf("GetLibraryPage request failed: %w", err)
	}

	return parseLibraryResponse(bodyBytes)
}

// getInt64 safely extracts int64 from any numeric type
func getInt64(v any) int64 {
	switch n := v.(type) {
	case uint64:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	case uint32:
		return int64(n)
	case int32:
		return int64(n)
	case int:
		return int64(n)
	}
	return 0
}

// getInt safely extracts int from any numeric type
func getInt(v any) int {
	return int(getInt64(v))
}

// parseLibraryResponse parses raw protobuf bytes into LibraryResponse
func parseLibraryResponse(data []byte) (*LibraryResponse, error) {
	decoded, err := decodeProtobufToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode protobuf: %w", err)
	}

	resp := &LibraryResponse{}

	root, ok := decoded.(map[string]any)
	if !ok {
		return resp, nil
	}

	field1, ok := root["1"].(map[string]any)
	if !ok {
		return resp, nil
	}

	// Extract state token (field 1.6)
	if st, ok := field1["6"].(string); ok {
		resp.StateToken = st
	}

	// Extract page token (field 1.5)
	if pt, ok := field1["5"].(string); ok {
		resp.PageToken = pt
	}

	// Parse media items (field 1.2)
	if mediaItems, ok := field1["2"].([]any); ok {
		for _, item := range mediaItems {
			if mi, ok := item.(map[string]any); ok {
				resp.MediaItems = append(resp.MediaItems, parseMediaItem(mi))
			}
		}
	}

	// Parse albums (field 1.3)
	if albums, ok := field1["3"].([]any); ok {
		for _, item := range albums {
			if ai, ok := item.(map[string]any); ok {
				resp.Albums = append(resp.Albums, parseAlbum(ai))
			}
		}
	}

	return resp, nil
}

func parseMediaItem(item map[string]any) MediaItemInfo {
	mi := MediaItemInfo{}

	// Media key (field 1)
	if mk, ok := item["1"].(string); ok {
		mi.MediaKey = mk
	}

	// Parse metadata (field 2)
	if meta, ok := item["2"].(map[string]any); ok {
		// Filename (field 2.4)
		if fn, ok := meta["4"].(string); ok {
			mi.Filename = fn
		}
		// Caption (field 2.3)
		if cap, ok := meta["3"].(string); ok {
			mi.Caption = cap
		}
		// Creation timestamp (field 2.7)
		if ts, exists := meta["7"]; exists {
			mi.CreationTimestamp = getInt64(ts)
		}
		// File size (field 2.10)
		if sz, exists := meta["10"]; exists {
			mi.FileSize = getInt64(sz)
		}
		// Dedup key (field 2.13.1 or 2.21.1)
		if dk, ok := meta["13"].(map[string]any); ok {
			if key, ok := dk["1"].(string); ok {
				mi.DedupKey = key
			}
		}
		// Album media key (field 2.1.1)
		if albumInfo, ok := meta["1"].(map[string]any); ok {
			if amk, ok := albumInfo["1"].(string); ok {
				mi.AlbumMediaKey = amk
			}
		}
		// Trash info (field 2.16)
		if trashInfo, ok := meta["16"].(map[string]any); ok {
			if state, exists := trashInfo["1"]; exists {
				mi.IsInTrash = getInt(state) == 2
			}
			if trashedAt, exists := trashInfo["3"]; exists {
				mi.TrashedAt = getInt64(trashedAt)
			}
		}
	}

	// Parse download info (field 5)
	if dlInfo, ok := item["5"].(map[string]any); ok {
		// Media type (field 5.1): 1=image, 2=video
		if mt, exists := dlInfo["1"]; exists {
			mi.IsVideo = getInt(mt) == 2
		}
		// Image info (field 5.2)
		if imgInfo, ok := dlInfo["2"].(map[string]any); ok {
			// Download URL (field 5.2.6)
			if url, ok := imgInfo["6"].(string); ok {
				mi.DownloadURL = url
			}
			// Thumbnail URL (field 5.2.1.1)
			if urlInfo, ok := imgInfo["1"].(map[string]any); ok {
				if url, ok := urlInfo["1"].(string); ok {
					mi.ThumbnailURL = url
				}
				// Dimensions
				if w, exists := urlInfo["2"]; exists {
					mi.Width = getInt(w)
				}
				if h, exists := urlInfo["3"]; exists {
					mi.Height = getInt(h)
				}
			}
		}
		// Video info (field 5.3)
		if vidInfo, ok := dlInfo["3"].(map[string]any); ok {
			// Download URL (field 5.3.5)
			if url, ok := vidInfo["5"].(string); ok {
				mi.DownloadURL = url
			}
			// Thumbnail URL (field 5.3.2.1)
			if thumbInfo, ok := vidInfo["2"].(map[string]any); ok {
				if url, ok := thumbInfo["1"].(string); ok {
					mi.ThumbnailURL = url
				}
				if w, exists := thumbInfo["2"]; exists {
					mi.Width = getInt(w)
				}
				if h, exists := thumbInfo["3"]; exists {
					mi.Height = getInt(h)
				}
			}
		}
	}

	return mi
}

func parseAlbum(item map[string]any) AlbumInfo {
	ai := AlbumInfo{}

	// Album key (field 1)
	if ak, ok := item["1"].(string); ok {
		ai.AlbumKey = ak
	}

	// Parse metadata (field 2)
	if meta, ok := item["2"].(map[string]any); ok {
		// Name (field 2.5)
		if name, ok := meta["5"].(string); ok {
			ai.Name = name
		}
		// Item count (field 2.7)
		if count, exists := meta["7"]; exists {
			ai.ItemCount = getInt(count)
		}
		// Cover key (field 2.17.1)
		if cover, ok := meta["17"].(map[string]any); ok {
			if ck, ok := cover["1"].(string); ok {
				ai.CoverKey = ck
			}
		}
	}

	return ai
}
