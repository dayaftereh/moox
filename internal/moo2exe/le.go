package moo2exe

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
)

const (
	leHeaderMinSize   = 0x84
	leObjectSize      = 24
	leObjectTableOff  = 0x40
	leObjectCountOff  = 0x44
	lePageMapOffOff   = 0x48
	lePageSizeOff     = 0x28
	leDataPagesOffOff = 0x80
)

type Executable struct {
	data       []byte
	sha256     string
	moduleBase int
	leOffset   int
	pageSize   int
	pageMapOff int
	dataPages  int
	objects    []object
}

type object struct {
	VirtualSize int
	PageMapIdx  int
	PageCount   int
}

func Open(path string) (*Executable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*Executable, error) {
	moduleBase, leOffset, err := findBoundLE(data)
	if err != nil {
		return nil, err
	}
	if leOffset+leHeaderMinSize > len(data) {
		return nil, fmt.Errorf("truncated LE header")
	}
	pageSize := int(binary.LittleEndian.Uint32(data[leOffset+lePageSizeOff : leOffset+lePageSizeOff+4]))
	if pageSize <= 0 {
		return nil, fmt.Errorf("invalid LE page size %d", pageSize)
	}
	objectTableRel := int(binary.LittleEndian.Uint32(data[leOffset+leObjectTableOff : leOffset+leObjectTableOff+4]))
	objectCount := int(binary.LittleEndian.Uint32(data[leOffset+leObjectCountOff : leOffset+leObjectCountOff+4]))
	pageMapRel := int(binary.LittleEndian.Uint32(data[leOffset+lePageMapOffOff : leOffset+lePageMapOffOff+4]))
	dataPages := int(binary.LittleEndian.Uint32(data[leOffset+leDataPagesOffOff : leOffset+leDataPagesOffOff+4]))
	if objectCount <= 0 || objectCount > 64 {
		return nil, fmt.Errorf("invalid LE object count %d", objectCount)
	}
	objectTable := leOffset + objectTableRel
	if objectTable < 0 || objectTable+objectCount*leObjectSize > len(data) {
		return nil, fmt.Errorf("LE object table outside file")
	}
	pageMapOff := leOffset + pageMapRel
	if pageMapOff < 0 || pageMapOff >= len(data) {
		return nil, fmt.Errorf("LE page map outside file")
	}

	objects := make([]object, objectCount)
	for i := 0; i < objectCount; i++ {
		off := objectTable + i*leObjectSize
		objects[i] = object{
			VirtualSize: int(binary.LittleEndian.Uint32(data[off : off+4])),
			PageMapIdx:  int(binary.LittleEndian.Uint32(data[off+12 : off+16])),
			PageCount:   int(binary.LittleEndian.Uint32(data[off+16 : off+20])),
		}
		if objects[i].PageMapIdx <= 0 || objects[i].PageCount <= 0 {
			return nil, fmt.Errorf("LE object %d has invalid page mapping", i+1)
		}
	}

	sum := sha256.Sum256(data)
	return &Executable{
		data:       data,
		sha256:     hex.EncodeToString(sum[:]),
		moduleBase: moduleBase,
		leOffset:   leOffset,
		pageSize:   pageSize,
		pageMapOff: pageMapOff,
		dataPages:  dataPages,
		objects:    objects,
	}, nil
}

func (e *Executable) SHA256() string  { return e.sha256 }
func (e *Executable) ModuleBase() int { return e.moduleBase }
func (e *Executable) LEOffset() int   { return e.leOffset }

func (e *Executable) ReadObject(objectNumber, objectOffset, size int) ([]byte, error) {
	if objectNumber < 1 || objectNumber > len(e.objects) {
		return nil, fmt.Errorf("object %d outside [1,%d]", objectNumber, len(e.objects))
	}
	if objectOffset < 0 || size < 0 {
		return nil, fmt.Errorf("negative object offset/size")
	}
	obj := e.objects[objectNumber-1]
	if objectOffset+size > obj.VirtualSize {
		return nil, fmt.Errorf("object %d range 0x%X..0x%X exceeds virtual size 0x%X", objectNumber, objectOffset, objectOffset+size, obj.VirtualSize)
	}
	out := make([]byte, size)
	remaining := size
	srcOff := objectOffset
	dstOff := 0
	for remaining > 0 {
		pageWithin := srcOff / e.pageSize
		pageRemainder := srcOff % e.pageSize
		if pageWithin >= obj.PageCount {
			return nil, fmt.Errorf("object %d page %d outside page count %d", objectNumber, pageWithin, obj.PageCount)
		}
		mapIndex := obj.PageMapIdx + pageWithin
		mapOff := e.pageMapOff + (mapIndex-1)*4
		if mapOff < 0 || mapOff+4 > len(e.data) {
			return nil, fmt.Errorf("page-map entry %d outside file", mapIndex)
		}
		entry := e.data[mapOff : mapOff+4]
		pageNum := int(entry[0])<<16 | int(entry[1])<<8 | int(entry[2])
		flags := entry[3]
		if pageNum <= 0 {
			return nil, fmt.Errorf("page-map entry %d has invalid page number %d", mapIndex, pageNum)
		}
		if flags != 0 {
			return nil, fmt.Errorf("page-map entry %d has unsupported LE page flags 0x%02X", mapIndex, flags)
		}
		physical := e.moduleBase + e.dataPages + (pageNum-1)*e.pageSize + pageRemainder
		chunk := e.pageSize - pageRemainder
		if chunk > remaining {
			chunk = remaining
		}
		if physical < 0 || physical+chunk > len(e.data) {
			return nil, fmt.Errorf("mapped object %d page %d outside file", objectNumber, pageWithin)
		}
		copy(out[dstOff:dstOff+chunk], e.data[physical:physical+chunk])
		remaining -= chunk
		srcOff += chunk
		dstOff += chunk
	}
	return out, nil
}

func findBoundLE(data []byte) (moduleBase, leOffset int, err error) {
	for base := 0; base+0x40 <= len(data); base++ {
		if data[base] != 'M' || data[base+1] != 'Z' {
			continue
		}
		rel := int(binary.LittleEndian.Uint32(data[base+0x3C : base+0x40]))
		if rel <= 0 || rel > len(data)-base-4 {
			continue
		}
		candidate := base + rel
		if candidate+2 <= len(data) && data[candidate] == 'L' && data[candidate+1] == 'E' {
			return base, candidate, nil
		}
	}
	return 0, 0, fmt.Errorf("bound MZ/LE module not found")
}
