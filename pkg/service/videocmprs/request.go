package videocmprs

type Request struct {
	ID int64 `jsonapi:"primary,requests"`

	Bitrate     int64 `jsonapi:"attr,bitrate" json:"bitrate"`
	ResolutionX int   `jsonapi:"attr,resolution_x" json:"resolution_x"`
	ResolutionY int   `jsonapi:"attr,resolution_y" json:"resolution_y"`
	RatioX      int   `jsonapi:"attr,ratio_x" json:"ratio_x"`
	RatioY      int   `jsonapi:"attr,ratio_y" json:"ratio_y"`
}
