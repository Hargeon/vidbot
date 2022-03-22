package consumer

type Response struct {
	ChatID      int64          `json:"chat_id"`
	Err         string         `json:"err"`
	Video       *VideoResponse `json:"video"`
	DownloadUrl string         `json:"download_url"`
}

type VideoResponse struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Bitrate     int64  `json:"bitrate"`
	ResolutionX int    `json:"resolution_x"`
	ResolutionY int    `json:"resolution_y"`
	RatioX      int    `json:"ratio_x"`
	RatioY      int    `json:"ratio_y"`
	ServiceID   string `json:"service_id,omitempty"`
}
