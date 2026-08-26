package moo2exe

import (
	"encoding/binary"
	"testing"
)

func TestParseBoundLEAndReadObject(t *testing.T) {
	const (
		moduleBase = 0x100
		leRel      = 0x80
		le         = moduleBase + leRel
		pageSize   = 0x100
		pageOff    = 0x300
		objTable   = 0xC4
		pageMap    = 0xF4
	)
	data := make([]byte, moduleBase+pageOff+2*pageSize+0x100)
	data[moduleBase] = 'M'
	data[moduleBase+1] = 'Z'
	binary.LittleEndian.PutUint32(data[moduleBase+0x3C:moduleBase+0x40], leRel)
	data[le] = 'L'
	data[le+1] = 'E'
	binary.LittleEndian.PutUint32(data[le+lePageSizeOff:le+lePageSizeOff+4], pageSize)
	binary.LittleEndian.PutUint32(data[le+leObjectTableOff:le+leObjectTableOff+4], objTable)
	binary.LittleEndian.PutUint32(data[le+leObjectCountOff:le+leObjectCountOff+4], 1)
	binary.LittleEndian.PutUint32(data[le+lePageMapOffOff:le+lePageMapOffOff+4], pageMap)
	binary.LittleEndian.PutUint32(data[le+leDataPagesOffOff:le+leDataPagesOffOff+4], pageOff)

	off := le + objTable
	binary.LittleEndian.PutUint32(data[off:off+4], 2*pageSize)
	binary.LittleEndian.PutUint32(data[off+12:off+16], 1)
	binary.LittleEndian.PutUint32(data[off+16:off+20], 2)

	pm := le + pageMap
	copy(data[pm:pm+4], []byte{0, 0, 1, 0})
	copy(data[pm+4:pm+8], []byte{0, 0, 2, 0})
	physical := moduleBase + pageOff
	for i := 0; i < 2*pageSize; i++ {
		data[physical+i] = byte(i)
	}

	exe, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if exe.ModuleBase() != moduleBase || exe.LEOffset() != le {
		t.Fatalf("module/le=%#x/%#x want %#x/%#x", exe.ModuleBase(), exe.LEOffset(), moduleBase, le)
	}
	got, err := exe.ReadObject(1, 0xF8, 24)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range got {
		want := byte(0xF8 + i)
		if value != want {
			t.Fatalf("byte %d=%#x want %#x", i, value, want)
		}
	}
}
