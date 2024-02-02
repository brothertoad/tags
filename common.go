package tags

import (
  "github.com/brothertoad/btu"
)

// Relative path is relative to the musicDir specified in the configuration file.
// Base path is the relative path with the extension removed (but with the trailing
// period retained).
const IdKey = "id"
const RelativePathKey = "relativePath"
const BasePathKey = "basePath"
const TitleKey = "title"
const ArtistKey = "artist"
const AlbumKey = "album"
const TrackNumberKey = "trackNumber"
const DiscNumberKey = "discNumber"
const ArtistSortKey = "artistSort"
const AlbumSortKey = "albumSort"
const DurationKey = "duration"
const MimeKey = "mime"
const ExtensionKey = "extension"
const EncodedExtensionKey = "encodedExtension"
const IsEncodedKey = "isEncoded"
const FlagsKey = "flags"
const Md5Key = "md5"
const SizeAndTimeKey = "sizeAndTime"
const EncodedSourceKey = "encodedSource" // size and time of source of encoding

const EncodeFlag = "e"

type TagMap map[string]string
type TagMapSlice []TagMap

// functions for sorting a slice of TagMaps
func (s TagMapSlice) Len() int { return len(s) }
func (s TagMapSlice) Less(i, j int) bool {
  if s[i][ArtistSortKey] != s[j][ArtistSortKey] { return s[i][ArtistSortKey] < s[j][ArtistSortKey] }
  if s[i][AlbumSortKey] != s[j][AlbumSortKey] { return s[i][AlbumSortKey] < s[j][AlbumSortKey] }
  // convert disc numbers from strings to ints and check those
  di := btu.Atoi(s[i][DiscNumberKey])
  dj := btu.Atoi(s[j][DiscNumberKey])
  if di != dj { return di < dj }
  // convert track numbers from strings to ints and check those
  tni := btu.Atoi(s[i][TrackNumberKey])
  tnj := btu.Atoi(s[j][TrackNumberKey])
  return tni < tnj
}
func (s TagMapSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
