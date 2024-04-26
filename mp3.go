package tags

import (
  "bytes"
  "io/ioutil"
  "log"
  "strings"
  "encoding/binary"
  "golang.org/x/text/transform"
  "golang.org/x/text/encoding/unicode"
)

// version 1, layer 1 bit rates
var v1l1BitRates = []float64{ 0, 32000.0, 64000.0, 96000.0,
  128000.0, 160000.0, 192000.0, 224000.0,
  256000.0, 288000.0, 320000.0, 352000.0,
  384000.0, 416000.0, 448000.0 }

// version 1, layer 2 bit rates
var v1l2BitRates = []float64{ 0, 32000.0, 40000.0, 56000.0,
  64000.0, 80000.0, 96000.0, 112000.0,
  128000.0, 160000.0, 192000.0, 224000.0,
  256000.0, 320000.0, 384000.0 }

// version 1, layer 3 bit rates
var v1l3BitRates = []float64{ 0, 32000.0, 40000.0, 48000.0,
  56000.0, 64000.0, 80000.0, 96000.0,
  112000.0, 128000.0, 160000.0, 192000.0,
  224000.0, 256000.0, 320000.0 }

// version 2, layer 1 bit rates
var v2l1BitRates = []float64{ 0, 32000.0, 48000.0, 56000.0,
  64000.0, 80000.0, 96000.0, 112000.0,
  128000.0, 144000.0, 160000.0, 176000.0,
  192000.0, 224000.0, 256000.0 }

// version 2, layers 2 and 3 bit rates
var v2l23BitRates = []float64{ 0, 8000.0, 16000.0, 24000.0,
  32000.0, 40000.0, 48000.0, 56000.0,
  64000.0, 80000.0, 96000.0, 112000.0,
  128000.0, 144000.0, 160000.0 }

var v1SampleRates = []float64{ 44100.0, 48000.0, 32000.0 }
var v2SampleRates = []float64{ 22050.0, 24000.0, 16000.0 }
var v25SampleRates = []float64{ 11025.0, 12000.0, 8000.0 }

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
      // If we don't have at least four bytes, just stop.
      if n > (len(buffer) - 4) {
        break
      }
      // Put validation in separate function - need to check for reserved version or layer
      if validHeader(buffer[n:(n+4)]) {
        frameSize, frameDuration := mp3ParseFrame(path, buffer[n:], n)
        totalFrameBytes += frameSize
        numFrames++
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

func validHeader(b []byte) bool {
  // Convert the first four bytes into a big-endian uint32.
  header := binary.BigEndian.Uint32(b[0:4])
  // Verify this is a frame sync
  if (header >> 21) != 0x07ff {
    return false
  }
  // Verify the MPEG version is not "reserved"
  if (header >> 19) & 0x03 == 0x01 {
    return false
  }
  // Verify the MPEG layer is not "reserved"
  if (header >> 17) & 0x03 == 0x00 {
    return false
  }
  // Verify the bit rate index is not "free" or "bad"
  bri := (header >> 12) & 0x0f
  if bri == 0 || bri == 0x0f {
    return false
  }
  // Verify the sampling rate index is not "reserved"
  if (header >> 10) & 0x03 == 0x03 {
    return false
  }
  return true
}

// Returns the size of this frame in bytes and the duration of the sound
// in this frame.  This link describes the frames:
// http://www.mp3-tech.org/programmer/frame_header.html
// These pages were helpful too:
// https://web.archive.org/web/20070821052201/https://www.id3.org/mp3Frame
// https://stackoverflow.com/questions/6220660/calculating-the-length-of-mp3-frames-in-milliseconds
func mp3ParseFrame(path string, buffer []byte, offset int) (int, float64) {
  // Convert the first four bytes into a big-endian uint32.
  header := binary.BigEndian.Uint32(buffer[0:4])
  versionIndex := (header >> 19) & 0x03
  layerIndex := (header >> 17) & 0x03
  // We only handle MP3 at this time.
  if versionIndex != 3 || layerIndex != 1 {
    // log.Fatalf("Got frame with version %d and layer %d at offset %d\n", version, layer, offset)
  }
  protection := (header & 0x010000) == 0
  bri := (header >> 12) & 0x0f  // bit rate index
  sri := (header >> 10) & 0x03  // sample rate index
  padding := (header >> 9) & 0x01 == 0x01
  // We now have enough info to calculate the size of the frame.
  bitRate, sampleRate := getBitAndSampleRates(path, versionIndex, layerIndex, bri, sri)
  frameSize := int((144.0 * bitRate) / sampleRate)
  if padding {
    frameSize += 1
  }
  if protection {
    frameSize += 2
  }
  return frameSize, 1152.0 / sampleRate
}

// Note that versionIndex and layerIndex are "raw" - i.e., directly from the frame.
// e.g. versionIndex == 3 means MPEG version 1 and layerIndex == 1 means layer III
func getBitAndSampleRates(path string, versionIndex, layerIndex, bri, sri uint32) (float64, float64) {
  // Version index == 3 => MPEG version 1
  if versionIndex == 3 {
    if layerIndex == 1 {
      return v1l3BitRates[bri], v1SampleRates[sri]
    } else if layerIndex == 2 {
      return v1l2BitRates[bri], v1SampleRates[sri]
    } else if layerIndex == 3 {
      return v1l1BitRates[bri], v1SampleRates[sri]
    }
  }
  // Version index == 2 => MPEG version 2
  if versionIndex == 2 {
    if layerIndex == 1 || layerIndex == 2 {
      return v2l23BitRates[bri], v2SampleRates[sri]
    } else if layerIndex == 3 {
      return v2l1BitRates[bri], v2SampleRates[sri]
    }
  }
  // Version index == 0 => MPEG version 2.5
  if versionIndex == 0 {
    if layerIndex == 1 || layerIndex == 2 {
      return v2l23BitRates[bri], v25SampleRates[sri]
    } else if layerIndex == 3 {
      return v2l1BitRates[bri], v25SampleRates[sri]
    }
  }
  // Don't handle anything else at this point.
  log.Fatalf("Unable to determine bit rate for %s from versionIndex %d and layerIndex %d\n", path, versionIndex, layerIndex)
  return 0.0, 0.0 // should never reach this
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
  // Get the size of the tag frame.
  frameSize := mp3GetID3Size(buffer[6:])
  // Start after the header, and read tags until we're through.
  // For now, let's just get the size.
  eob := headerSize + frameSize // eob means end of buffer
  for j := headerSize; j < eob; {
    // Skip zero bytes between tags.  Note: if we find a zero byte, we're probably at the end of the tags.
    if buffer[j] == 0 {
      j++
      continue
    }
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
  return eob
}

// TASK: move this to btu
func stringFromUTF16(b []byte) string {
  bomEncoder := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
  bomReader := transform.NewReader(bytes.NewReader(b), bomEncoder.NewDecoder())
  decoded, err := ioutil.ReadAll(bomReader)
  if err != nil {
    log.Fatalf("Unable to get a string from UTF16\n")
  }
  s := string(decoded)
  return s
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
