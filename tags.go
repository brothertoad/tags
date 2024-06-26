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
  "DISCNUMBER" : DiscNumberKey,
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
  // Check for the track number.  If it exists, clean it up.  If not, see if
  // the TRCK tag exists, which is track number / track total and get the track number from that.
  if v, present := song[TrackNumberKey]; present {
      song[TrackNumberKey] = cleanUpNumber(v)
    } else {
      if tntt, tnttPresent := song["TRCK"]; tnttPresent {
        song[TrackNumberKey] = cleanUpNumber(tntt)
      } else {
      log.Printf("Can't get track number for '%s'\n", song[RelativePathKey])
    }
  }
  // Check for the disc number.  If it exists, clean it up.  If not, see if it has the
  // TPOS tag, which is disc number / disc total and get the disc number from that.
  // If that doesn't exist, assume disc 1.
  if v, present := song[DiscNumberKey]; present {
    song[DiscNumberKey] = cleanUpNumber(v)
  } else {
    if dndt, dndtPresent := song["TPOS"]; dndtPresent {
      song[DiscNumberKey] = cleanUpNumber(dndt)
    } else {
      song[DiscNumberKey] = "1"
    }
  }
}

// Clean up track and disc numbers by removing a slash (and anything following the
// slash) and also removing a leading zero if there is one.
func cleanUpNumber(v string) string {
  // If the string contains a slash, just get the part in front of the slash.
  if strings.Index(v, "/") >= 0 {
    s := strings.Split(v, "/")
    v = s[0]
  }
  return stripLeadingZero(v)
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
