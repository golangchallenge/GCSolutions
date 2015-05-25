package main

type AllFlickrGroups struct {
	Groups FlickrGroups `json:"groups"`
	Stat   string       `json:"stat"`
}

type FlickrGroups struct {
	Page    int           `json:"page"`
	Pages   int           `json:"pages"`
	Perpage int           `json:"perpage"`
	Total   string        `json:"total"`
	Group   []FlickrGroup `json:"group"`
}

type FlickrGroup struct {
	NSID         string `json:"nsid"`
	Name         string `json:"name"`
	Eighteenplus int    `json:"eighteenplus"`
	Iconserver   string `json:"iconserver"`
	Iconfarm     int    `json:"iconfarm"`
	Members      string `json:"members"`
	Pool_count   string `json:"pool_count"`
	Topic_count  string `json:"topic_count"`
	Privacy      string `json:"privacy"`
}

type AllFlickrPool struct {
	Groups FlickrPoolPhotos `json:"photos"`
	Stat   string           `json:"stat"`
}

type FlickrPoolPhotos struct {
	Page       int              `json:"page"`
	Pages      int              `json:"pages"`
	PerPage    int              `json:"perpage"`
	Total      string           `json:"total"`
	PoolPhotos []FlickPoolPhoto `json:"photo"`
}

type FlickPoolPhoto struct {
	Id        string `json:"id"`
	Owner     string `json:"owner"`
	Secret    string `json:"secret"`
	Server    string `json:"server"`
	Farm      int    `json:"farm"`
	Title     string `json:"title"`
	IsPublic  int    `json:"ispublic"`
	IsFriend  int    `json:"isfriend"`
	IsFamily  int    `json:"isfamily"`
	OwnerName string `json:"ownername"`
	DateAdded string `json:"dateadded"`
}
