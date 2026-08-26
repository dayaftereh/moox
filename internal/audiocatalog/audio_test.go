package audiocatalog

import (
	"encoding/binary"
	"testing"
)

func TestParseWAVEFormat(t *testing.T) {
	data := make([]byte, 44)
	copy(data[0:4], "RIFF")
	binary.LittleEndian.PutUint32(data[4:8], 36)
	copy(data[8:12], "WAVE")
	copy(data[12:16], "fmt ")
	binary.LittleEndian.PutUint32(data[16:20], 16)
	binary.LittleEndian.PutUint16(data[20:22], 1)
	binary.LittleEndian.PutUint16(data[22:24], 2)
	binary.LittleEndian.PutUint32(data[24:28], 22050)
	binary.LittleEndian.PutUint32(data[28:32], 44100)
	binary.LittleEndian.PutUint16(data[32:34], 2)
	binary.LittleEndian.PutUint16(data[34:36], 8)
	copy(data[36:40], "data")

	got, ok := parseWAVEFormat(data)
	if !ok {
		t.Fatal("WAVE not recognized")
	}
	if got.FormatTag != 1 || got.Channels != 2 || got.SampleRate != 22050 || got.BitsPerSample != 8 {
		t.Fatalf("format=%+v", got)
	}
}
