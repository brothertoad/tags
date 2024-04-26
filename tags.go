package tags

import (
  "log"
  "strings"
)

var keyTranslations = map[string]string {
  "\xa9nam": TitleKey,
  "\xa9ART" : ArtistKey,
  "\xa9alb" : AlbumKey,
  "soar" : ArtistSortKey,
  "soal" : AlbumSortKey,
  "ALBUM" : AlbumKey,
  "ARTIST" : ArtistKey,
  "TITLE" : TitleKey,
  "trkn" : TrackNumberKey,
  "disk" : DiscNumberKey,
  "tracknumber" : TrackNumberKey,
  "TRACKNUMBER" : TrackNumberKey,
  "DISKNUMBER" : DiscNumberKey,
  "TIT2" : TitleKey,
  "TPE1" : ArtistKey,
  "TALB" : AlbumKey,
}

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

func GetStandardTagsFromFile(path string) TagMap {
  tagMap := GetTagsFromFile(path)
  if tagMap == nil || len(tagMap) == 0 {
    return tagMap
  }
  translateKeys(tagMap)
  return tagMap
}

// Replace keys with standard names.
func translateKeys(song TagMap) {
  for k, v := range song {
    if trans, present := keyTranslations[k]; present {
      delete(song, k)
      song[trans] = v
    }
  }
  // If there is a track number, and it has a slash, change it to the TRCK tag.
  if v, present := song[TrackNumberKey]; present {
    if strings.Index(v, "/") >= 0 {
      s := strings.Split(v, "/")
      song[TrackNumberKey] = s[0]
    }
  }
  // Check for the track number.  If it doesn't exist, see if it has the TRCK tag, which
  // is track number / track total and get the track number from that.
  if _, present := song[TrackNumberKey]; !present {
    if tntt, tnttPresent := song["TRCK"]; tnttPresent {
      s := strings.Split(tntt, "/")
      song[TrackNumberKey] = stripLeadingZero(s[0])
    } else {
      log.Printf("Can't get track number for '%s'\n", song[RelativePathKey])
    }
  }
  // Check for the disc number.  If it doesn't exist, see if it has the TPOS tag, which
  // is disc number / disc total and get the disc number from that.  If that doesn't
  // exist, assume disc 1.
  if _, present := song[DiscNumberKey]; !present {
    if dndt, dndtPresent := song["TPOS"]; dndtPresent {
      s := strings.Split(dndt, "/")
      song[DiscNumberKey] = stripLeadingZero(s[0])
    } else {
      song[DiscNumberKey] = "1"
    }
  }
}

func stripLeadingZero(s string) string {
  if len(s) == 1 {
    return s
  }
  if s[0] == '0' {
    return s[1:]
  }
  return s
}
