package tags

import (
  "log"
  "strings"
)

const magic = 0x664c6143
const id3Magic = 0x49443300
const streaminfotype byte = 0
const commenttype byte = 4

// Most of the info for this code came from this page:
// https://xiph.org/flac/format.html

func FlacTagsFromFile(path string) TagMap {
  song := make(TagMap)
  bb := bbFromFilePrefix(path, 256 * 1024)
  bbMagic := bb.read32BE()
  haveComments := false
  haveDuration := false
  if bbMagic != magic {
    // If the buffer doesn't start with an ID3 block, nothing we can do.
    if (bbMagic & 0xffffff00) != id3Magic {
      log.Printf("flac file %s does not have correct magic number\n", path)
      return song
    }
    bb.n = bb.n - 4 // un-read the magic
    id3size := mp3ParseID3(bb.b, song)
    bb.skip(uint32(id3size + 4))  // the +4 is to skip the magic
  }
  for {
    blocktype, lastone, size := nextmetablock(bb)
    if blocktype == commenttype {
      cbb := bytebufferfromparent(bb, size)
      getFlacComments(cbb, song)
      setMimeAndExtension("audio/flac", "flac", song)
      song[EncodedExtensionKey] = "mp3"
      song[IsEncodedKey] = "false"
      haveComments = true
      if haveComments && haveDuration {
		  break
	  }
    } else if blocktype == streaminfotype {
      sibb := bytebufferfromparent(bb, size)
      getFlacDuration(sibb, song)
      haveDuration = true
      if haveComments && haveDuration {
		  break
	  }
    } else {
      bb.skip(size)
    }
    if lastone {
      break
    }
  }
  return song
}

func getFlacComments(cbb *bytebuffer, m TagMap) {
  vendorsize := cbb.read32LE()
  cbb.skip(vendorsize)
  num := cbb.read32LE()
  for j:= 0; j < int(num); j++ {
    size := cbb.read32LE()
    comment := string(cbb.read(size))
    parts := strings.Split(comment, "=")
    m[parts[0]] = parts[1]
  }
}

func getFlacDuration(bb *bytebuffer, m TagMap) {
  bb.skip(10)
  // We're going to do a shortcut, and assume the upper four bits of the
  // total samples are zero.  This is good to over 750 minutes.
  sampleSize := float64(bb.read32BE() >> 12)
  numSamples := float64(bb.read32BE())
  setDuration(numSamples / sampleSize, m)
}

func nextmetablock(bb *bytebuffer) (byte, bool, uint32) {
  blocktype := bb.peek()
  lastone := blocktype > 127
  if lastone {
    blocktype -= 128
  }
  size := bb.read32BE()
  size &= 0x00ffffff
  return blocktype, lastone, size
}
