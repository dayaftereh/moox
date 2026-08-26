package textcatalog

import (
	"encoding/binary"
	"testing"
)

func TestAnalyzeFixedRecordStringTableAndArray(t *testing.T) {
	fixed := make([]byte, 104)
	binary.LittleEndian.PutUint16(fixed[0:2], 1)
	binary.LittleEndian.PutUint16(fixed[2:4], 100)
	copy(fixed[4:], []byte("Hello world"))
	rec, ok := analyze("TEST.LBX", 0, fixed)
	if !ok || rec.Kind != "fixed_record_v1" || rec.BodyOffset != 4 || rec.BodySize != 100 || len(rec.Runs) != 1 {
		t.Fatalf("fixed=%+v ok=%v", rec, ok)
	}

	array := make([]byte, 44)
	binary.LittleEndian.PutUint16(array[0:2], 2)
	binary.LittleEndian.PutUint16(array[2:4], 20)
	copy(array[4:24], []byte("Alpha"))
	copy(array[24:44], []byte("Beta"))
	rec, ok = analyze("TEST.LBX", 1, array)
	if !ok || rec.Kind != "fixed_array_v1" || rec.ArrayCount != 2 || rec.RecordSize != 20 || len(rec.ArrayRecords) != 2 || countRuns(rec) != 2 {
		t.Fatalf("array=%+v ok=%v", rec, ok)
	}
	if rec.ArrayRecords[1].Offset != 24 {
		t.Fatalf("second record offset=%d", rec.ArrayRecords[1].Offset)
	}

	table := []byte("Alpha\x00Beta\x00Gamma\x00Delta\x00")
	rec, ok = analyze("TEST.LBX", 2, table)
	if !ok || rec.Kind != "string_table_candidate" || len(rec.Runs) != 4 {
		t.Fatalf("table=%+v ok=%v", rec, ok)
	}
}
