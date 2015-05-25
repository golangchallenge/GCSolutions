package model

import (
	"mime"
	"os/exec"
	"path"
	"strings"
)

// Image model.
type Image struct {
	Path     string `json:"path,omitempty"`
	MIMEType string `json:"mimetype,omitempty"`
}

// NewImage creates a new Image pointer.
func NewImage(imgPath string) *Image {
	mimeType := mime.TypeByExtension(path.Ext(imgPath))
	return &Image{
		Path:     imgPath,
		MIMEType: mimeType,
	}
}

// User model.
type User struct {
	ID        string `json:"id,omitempty"`
	Anonymous bool   `json:"anonymous,omitempty"`
	Token     string `json:"-"`
	Target    *Image `json:"-"`
	Mosaic    *Image `json:"-"`
}

// NewUser creates a new User pointer, currently not used.
func NewUser(token string) *User {
	return &User{
		ID:        generateUUID(),
		Token:     token,
		Anonymous: false,
	}
}

// NewAnonymousUser creates a new anonymous User pointer.
func NewAnonymousUser(token string) *User {
	return &User{
		ID:        generateUUID(),
		Token:     token,
		Anonymous: true,
	}
}

// generateUUID generates a unique user ID.
func generateUUID() string {
	out, _ := exec.Command("uuidgen").Output()
	uuid := strings.TrimSuffix(string(out), "\n")
	return uuid
}
