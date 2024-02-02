package tags

import (
  "strings"
)

func GetTagsFromFile(path string) TagMap {
  if strings.HasSuffix(path, "flac") {
    return FlacTagsFromFile(path)
  } else if strings.HasSuffix(path, "mp3") {
    return Mp3TagsFromFile(path)
  } else if strings.HasSuffix(path, "m4a") {
    return M4aTagsFromFile(path)
  }
  return make(TagMap)
}
