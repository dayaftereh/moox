package textcatalog

import (
	"encoding/binary"
	"testing"
)

func TestAnalyzeFixedRecordAndStringTable(t *testing.T) {
	fixed := make([]byte, 104)
	binary.LittleEndian.PutUint16(fixed[0:2], 1)
	binary.LittleEndian.PutUint16(fixed[2:4], 100)
	copy(fixed[4:], []byte("Hello world"))
	rec, ok := analyze("TEST.LBX", 0, fixed)
	if !ok || rec.Kind != "fixed_record_v1" || rec.BodyOffset != 4 || rec.BodySize != 100 || len(rec.Runs) != 1 {
		t.Fatalf("fixed=%+v ok=%v", rec, ok)
	}

	table := []byte("Alpha\x00Beta\x00Gamma\x00Delta\x00")
	rec, ok = analyze("TEST.LBX", 1, table)
	if !ok || rec.Kind != "string_table_candidate" || len(rec.Runs) != 4 {
		t.Fatalf("table=%+v ok=%v", rec, ok)
	}
}
