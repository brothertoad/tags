package tags

import (
  "encoding/binary"
)

type bytebuffer struct {
  b [] byte
  n int
}

func bytebufferfromfile(path string) *bytebuffer {
  bb := new(bytebuffer)
  bb.b = readFile(path)
  return bb
}

func bytebufferfromslice(buffer []byte) *bytebuffer {
  bb := new(bytebuffer)
  bb.b = buffer
  return bb
}

func bytebufferfromparent(parent *bytebuffer, size uint32) *bytebuffer {
  bb := new(bytebuffer)
  bb.b = parent.b[parent.n:parent.n+int(size)]
  parent.skip(size)
  return bb
}

func (bb *bytebuffer) peek() byte {
  return bb.b[bb.n]
}

func (bb *bytebuffer) remaining() int {
  return len(bb.b) - bb.n
}

func (bb *bytebuffer) rewind() {
  bb.n = 0
}

func (bb *bytebuffer) readByte() uint32 {
  if (bb.n + 1) > len(bb.b) {
    panic("Attempt to read byte past end of byte buffer")
  }
  u := uint32(bb.b[bb.n])
  bb.n++
  return u
}

func (bb *bytebuffer) read32BE() uint32 {
  if (bb.n + 4) > len(bb.b) {
    panic("Attempt to read 32 BE past end of byte buffer")
  }
  u := binary.BigEndian.Uint32(bb.b[bb.n:bb.n+4])
  bb.n += 4
  return u
}

func (bb *bytebuffer) read16BE() uint16 {
  if (bb.n + 2) > len(bb.b) {
    panic("Attempt to read 16 BE past end of byte buffer")
  }
  u := binary.BigEndian.Uint16(bb.b[bb.n:bb.n+2])
  bb.n += 2
  return u
}

func (bb *bytebuffer) read32LE() uint32 {
  if (bb.n + 4) > len(bb.b) {
    panic("Attempt to read 32 LE past end of byte buffer")
  }
  u := binary.LittleEndian.Uint32(bb.b[bb.n:bb.n+4])
  bb.n += 4
  return u
}

func (bb *bytebuffer) read(size uint32) []byte {
  if (int(size) + bb.n) > len(bb.b) {
    panic("Attempt to read past end of byte buffer")
  }
  b := bb.b[bb.n:bb.n+int(size)]
  bb.n += int(size)
  return b
}

func (bb *bytebuffer) skip(skip uint32) {
  if (int(skip) + bb.n) > len(bb.b) {
    panic("Attempt to skip past end of byte buffer")
  }
  bb.n += int(skip)
}
