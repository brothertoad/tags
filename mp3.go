package tags

import (
  "log"
  "strings"
  "encoding/binary"
  "unicode/utf16"
)

// version 1, layer 3 bit rates and sample rates
var v1l3BitRates = []float64{ 0, 32000.0, 40000.0, 48000.0,
  56000.0, 64000.0, 80000.0, 96000.0,
  112000.0, 128000.0, 160000.0, 192000.0,
  224000.0, 256000.0, 320000.0 }

// version 2, layer 3 bit rates and sample rates
var v2l3BitRates = []float64{ 0, 8000.0, 16000.0, 24000.0,
  32000.0, 40000.0, 48000.0, 56000.0,
  64000.0, 80000.0, 96000.0, 112000.0,
  128000.0, 144000.0, 160000.0 }

var v1l3SampleRates = []float64{ 44100.0, 48000.0, 32000.0 }

func Mp3TagsFromFile(path string) TagMap {
  buffer := readFile(path)
  m := make(TagMap)
  // Look at each byte.  If the byte is 0xff, check to see if the upper three bits
  // of the next byte are set.  If so, it is the start of a frame.  If not, check
  // to see if the byte is 0x49, which represents the letter 'I'.  If so, check
  // to see if it is followed by "D3".  If so, it is the start of the ID3 block.
  // If neither of these is true, then just ignore the byte and move on.
  numFrames := 0
  totalFrameBytes := 0
  duration := 0.0
  var increment int
  for n := 0; n < len(buffer); n += increment {
    increment = 1
    b := buffer[n]
    if b == 0xff {
      // If we're at the last byte, just stop.
      if n == (len(buffer) - 1) {
        break
      }
      if (buffer[n+1] & 0xe0) == 0xe0 {
        numFrames++
        frameSize, frameDuration := mp3ParseFrame(path, buffer[n:], n)
        totalFrameBytes += frameSize
        increment = frameSize
        duration = duration + frameDuration
      }
    } else if b == 0x49 {
      if buffer[n+1] == 0x44 && buffer[n+2] == 0x33 {
        increment = mp3ParseID3(buffer[n:], m)
      }
    }
  }
  setDuration(duration, m)
  setMimeAndExtension("audio/mp3", "mp3", m)
  m[EncodedExtensionKey] = "mp3"
  m[IsEncodedKey] = "true"
  return m
}

// Returns the size of this frame in bytes and the duration of the sound
// in this frame.  This link describes the frames:
// http://www.mp3-tech.org/programmer/frame_header.html
// These pages were helpful too:
// https://web.archive.org/web/20070821052201/https://www.id3.org/mp3Frame
// https://stackoverflow.com/questions/6220660/calculating-the-length-of-mp3-frames-in-milliseconds
func mp3ParseFrame(path string, buffer []byte, offset int) (int, float64) {
  version := (buffer[1] >> 3) & 0x03
  layer := (buffer[1] >> 1) & 0x03
  // We only handle MP3 at this time.
  if version != 3 || layer != 1 {
    // log.Fatalf("Got frame with version %d and layer %d at offset %d\n", version, layer, offset)
  }
  protection := (buffer[1] & 0x01) == 0
  bri := buffer[2] >> 4  // bit rate index
  sri := (buffer[2] >> 2) & 0x03  // sample rate index
  padding := (buffer[2] >> 1) & 0x01 == 0x01
  // We now have enough info to calculate the size of the frame.
  bitRate := getBitRate(path, version, layer, bri) // v1l3BitRates[bri]
  sampleRate := v1l3SampleRates[sri]
  frameSize := int((144.0 * bitRate) / sampleRate)
  if padding {
    frameSize += 1
  }
  if protection {
    frameSize += 2
  }
  return frameSize, 1152.0 / sampleRate
}

// Note that the version and layer are "raw" - i.e., directly from the frame.
// e.g. version == 3 means MPEG version 1
// layer == 1 means layer III
func getBitRate(path string, version, layer, bri byte) float64 {
  // The usual for MP3's.
  if version == 3 && layer == 1 {
    return v1l3BitRates[bri]
  }
  if version == 2 && layer == 1 {
    return v2l3BitRates[bri]
  }
  // Don't handle anything else at this point.
  log.Fatalf("Unable to determine bit rate for %s from version %d and layer %d\n", path, version, layer)
  return 0.0
}

// MP3 ID3 blocks are described here:
// https://id3.org/id3v2.3.0
// The encodings are listed here:
// https://stackoverflow.com/questions/9857727/text-encoding-in-id3v2-3-tags
func mp3ParseID3(buffer []byte, m TagMap) int {
  headerSize := 10
  // Check for extended header
  if buffer[5] & 0x40 == 0x40 {
    headerSize += 4 + int(buffer[13])
  }
  // Start after the header, and read tags until we're through.
  // For now, let's just get the size.
  for j := headerSize; j < len(buffer); {
    key := string(buffer[j:j+4])
    size := int(binary.BigEndian.Uint32(buffer[j+4:j+8]))
    if strings.HasPrefix(key, "T") {
      // Ignore encoding for now.
      encoding := buffer[j+10]
      var value string
      // 0 is ASCII, 3 is UTF-8
      if encoding == 0 || encoding == 3 {
        value = string(buffer[j+11:j+size+10])
      } else if encoding == 1 {
        // String is UTF-16, with BOM
        value = stringFromUTF16(buffer[j+11:j+size+10])
      }
      // Some tags have a zero byte at the end to make the string an even length.
      // We need to remove it.
      m[key] = strings.TrimSuffix(value, "\000")
    }
    j += size + 10
  }
  return mp3GetID3Size(buffer[6:]) + headerSize
}

func stringFromUTF16(b []byte) string {
  b16 := make([]uint16, len(b)/2)
  for j := 0; j < len(b)/2; j++ {
    b16[j] = uint16(b[2*j+1]) << 8 + uint16(b[2*j])
  }
  return string(utf16.Decode(b16))
}

func mp3GetID3Size(b []byte) int {
  // Read four bytes, use the lower 7 bits of each one to form a 28-bit size.
  var total int = 0
  for j := 0; j < 4; j++ {
    total <<= 7
    total += int(b[j]) & 0x7f
  }
  return total
}
