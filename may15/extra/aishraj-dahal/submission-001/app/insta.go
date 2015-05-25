package app

import "encoding/json"

//ImageMetaData is a wrapper around the instagram response
type ImageMetaData struct {
	ThumbnailURL string
	Height       int
	Width        int
	MediaID      string
}

type AuthenticationUser struct {
	ID             string `json:"id"`
	UserName       string `json:"user_name"`
	FullName       string `json:"full_name"`
	ProfilePicture string `json:"profile_picture"`
}

type AuthenticationResponse struct {
	AccessToken string             `json:"access_token"`
	User        AuthenticationUser `json:"user"`
}

type APIResponse struct {
	Pagination PaginationObject `json:"pagination"`
	Meta       MetaData         `json:"meta"`
	Data       []TagResponse    `json:"data"`
	//TODO add error response also
}

type PaginationObject struct {
	NextURL            string `json:"next_url"`
	NextMaxID          string `json:"next_max_id,omitempty"`
	DeprecationWarning string `json:"deprecation_warning,omitempty"`
	NextMaxTagID       string `json:"next_max_tag_id,omitempty"`
	NextMinID          string `json:"next_min_id,omitempty"`
	MinTagID           string `json:"min_tag_id,omitempty"`
}

type MetaData struct {
	ErrorType    string `json:"error_type,omitempty"`
	Code         int32  `json:"code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type TagResponse struct {
	Attribution  json.RawMessage   `json:"attribution,omitempty"`
	Videos       json.RawMessage   `json:"videos,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	MediaType    string            `json:"type,omitempty"`
	Location     json.RawMessage   `json:"location,omitempty"`
	Comments     json.RawMessage   `json:"comments,omitempty"`
	Filter       json.RawMessage   `json:"filter,omitempty"`
	CreatedTime  string            `json:"created_time,omitempty"`
	Link         string            `json:"link,omitempty"`
	Images       DigitialMediaInfo `json:"images,omitempty"`
	Likes        json.RawMessage   `json:"likes,omitempty"`
	UsersInPhoto json.RawMessage   `json:"users_in_photo,omitempty"`
	Caption      json.RawMessage   `json:"caption,omitempty"`
	UserLinked   json.RawMessage   `json:"user_has_liked,omitempty"`
	ID           string            `json:"id"`
	UserInfo     json.RawMessage   `json:"user,omitempty"`
}

type DigitialMediaInfo struct {
	LowResolution      ImageDetails `json:"low_resolution,omitempty"`
	Thumbnail          ImageDetails `json:"thumbnail,omitempty"`
	StandardResolution ImageDetails `json:"standard_resolution,omitempty"`
}

type ImageDetails struct {
	URL    string `json:"url"`
	Height int32  `json:"height"`
	Width  int32  `json:"width"`
}
