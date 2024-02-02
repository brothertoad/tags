package tags

import (
  "os"
  "math"
  "fmt"
)

func check(e error) {
  if e!= nil {
    panic(e)
  }
}

func readFile(path string) []byte {
  b, err := os.ReadFile(path)
  check(err)
  return b
}

func setDuration(duration float64, m TagMap) {
  // Round to nearest integer, make it a string, convert to mm:ss.
  totalSeconds := int(math.Round(duration))
  minutes := totalSeconds / 60
  seconds := totalSeconds % 60
  m[DurationKey] = fmt.Sprintf("%d:%02d", minutes, seconds)
}

func setMimeAndExtension(mime string, extension string, m TagMap) {
  m[MimeKey] = mime
  m[ExtensionKey] = extension
}
