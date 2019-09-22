package trackinfo

// Model mirrors the structure of track_info_config.json. Used to edit
// track_info_config's contents
type Model struct {
	IsTrackInfoNew bool `json:"isTrackInfoNew"`
}
