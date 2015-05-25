package common

import (
	"database/sql"
	"image"
	"image/color"
	"log"
)

type Tile struct {
	X            int
	Y            int
	Rect         image.Rectangle
	AverageColor color.RGBA
	MatchedImage string
}

type AppContext struct {
	Log          *log.Logger
	SessionStore *SessionManager
	Db           *sql.DB
}

type tinyUser struct {
	DisplayName string
}
