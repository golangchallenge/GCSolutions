// Package imgur interacts with images and apis on imgur.com.
// It implements oauth authentication, as well as api calls required to fetch galleries and images.
package imgur

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var client_id, client_secret string
var aes_key, hmac_key []byte

func init() {
	if client_id = os.Getenv("IMGUR_ID"); client_id == "" {
		log.Fatal("IMGUR_ID environment variable required.")
	}
	if client_secret = os.Getenv("IMGUR_SECRET"); client_secret == "" {
		log.Fatal("IMGUR_SECRET environment variable required")
	}
	var encrypt_key string
	if encrypt_key = os.Getenv("IMGUR_COOKIE_KEY"); encrypt_key == "" {
		encrypt_key = client_secret
		log.Println("WARNING: IMGUR_COOKIE_KEY environment variable not set. Cookies may be insecure.")
	}
	// get 256 bit hash of key to make seperate 128 bit keys for aes and hmac
	key32 := sha256.Sum256([]byte(encrypt_key))
	aes_key, hmac_key = key32[:16], key32[16:]
}

// The url to redirect the user to in order to authorize this application.
func ImgurLoginUrl() string {
	return fmt.Sprintf("https://api.imgur.com/oauth2/authorize?client_id=%s&response_type=code", client_id)
}

// This callback should be wired up to the callback url imgur has for your application.
// After completing its work, the user will be redirected back to "/".
// If an error occured, the redirect url will have ?oauthErr = 1 set.
// Otherwise a token cookie should be set.
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	success := true
	defer func() {
		//always redirect to /, passing error message if applicable.
		url := "/"
		if !success {
			url += "?oauthErr=1"
		}
		http.Redirect(w, r, url, 302)
	}()
	if e := r.FormValue("error"); e != "" {
		log.Printf("error returned from imgur: %s.\n", e)
		success = false
		return
	}
	code := r.FormValue("code")
	if code == "" {
		log.Printf("No code from imgur.\n")
		success = false
		return
	}
	token, err := parseToken(http.PostForm("https://api.imgur.com/oauth2/token",
		url.Values{"code": {code}, "client_id": {client_id}, "grant_type": {"authorization_code"}, "client_secret": {client_secret}}))
	if err != nil {
		success = false
		return
	}
	token.storeInCookie(w)
}

// Struct to hold an access token from imgur
type ImgurAccessToken struct {
	AccessToken     string    `json:"access_token"`
	ExpiresIn       int       `json:"expires_in"`
	RefreshToken    string    `json:"refresh_token"`
	AccountId       int64     `json:"account_id"`
	AccountUsername string    `json:"account_username"`
	ExpirationTime  time.Time `json:"expiration"`
}

const cookieName = "ImgAccTok"

// TokenForRequest reads the imgur token cookie, decodes it, and builds an ImgurAccessToken.
// If no valid token is found, result will be nil. If expired token is found, it will attempt to
// refresh it and store it via the current request.
func TokenForRequest(w http.ResponseWriter, r *http.Request) *ImgurAccessToken {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	tok := decodeCookie(cookie.Value)
	if tok == nil {
		ClearImgurCookie(w)
		return nil
	}

	// if expires in next 5 minutes or sooner, refresh token
	if tok.ExpirationTime.Before(time.Now().Add(5 * time.Minute)) {
		tok = tok.renew()
		if tok == nil {
			ClearImgurCookie(w)
		} else {
			tok.storeInCookie(w)
		}
	}
	return tok
}

func ClearImgurCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Path: "/", MaxAge: -1})
}

func (t *ImgurAccessToken) renew() *ImgurAccessToken {
	tok, err := parseToken(http.PostForm("https://api.imgur.com/oauth2/token",
		url.Values{"refresh_token": {t.RefreshToken}, "client_id": {client_id}, "grant_type": {"refresh_token"}, "client_secret": {client_secret}}))
	if err != nil {
		return nil
	}

	return tok
}

func parseToken(resp *http.Response, err error) (*ImgurAccessToken, error) {
	if err != nil {
		log.Println("Error posting to token endpoint", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Println("Bad auth response from imgur: ", resp.StatusCode)
		return nil, fmt.Errorf("Non-200")
	}
	decoder := json.NewDecoder(resp.Body)
	token := &ImgurAccessToken{}
	err = decoder.Decode(token)
	if err != nil {
		log.Println("Decoding error", err)
		return nil, fmt.Errorf("json-error")
	}
	token.ExpirationTime = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return token, nil
}

func (t *ImgurAccessToken) storeInCookie(w http.ResponseWriter) {
	data := t.encodeForCookie()
	if data == "" {
		return
	}
	cookie := &http.Cookie{}
	cookie.HttpOnly = true
	cookie.Name = cookieName
	cookie.Path = "/"
	cookie.Secure = true
	cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
	cookie.Value = data
	http.SetCookie(w, cookie)
}

// We don't want to tell anybody the access token, lest they make bad oauth requests on our behalf.
// I'd also rather not store it serverside for this challenge either :)
// We will use this function to encode the token for putting in cookie.
//
// Token -> json -> aes(cfb) + hmac -> base64
func (t *ImgurAccessToken) encodeForCookie() string {
	plaintext, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	block, err := aes.NewCipher(aes_key)
	if err != nil {
		return ""
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return ""
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	mac := hmac.New(sha256.New, hmac_key)
	_, err = mac.Write(ciphertext)
	if err != nil {
		return ""
	}
	expectedMAC := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(append(expectedMAC, ciphertext...))
}

// Decodes cookie data into a token struct.
// This is the exact opposite of encodeForCookie.
// If any error is encountered, token will be null and cookie should be cleared.
func decodeCookie(cookieData string) *ImgurAccessToken {
	ciphertext, err := base64.StdEncoding.DecodeString(cookieData)
	if err != nil {
		return nil
	}
	if len(ciphertext) < sha256.Size+aes.BlockSize { //at least room for iv and hmac
		return nil
	}
	//first validate hmac
	msgMac := ciphertext[:sha256.Size]
	ciphertext = ciphertext[sha256.Size:]
	mac := hmac.New(sha256.New, hmac_key)
	_, err = mac.Write(ciphertext)
	if err != nil {
		return nil
	}
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(msgMac, expectedMAC) {
		return nil
	}
	// pull out iv and decrypt
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	block, err := aes.NewCipher(aes_key)
	if err != nil {
		return nil
	}
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	token := &ImgurAccessToken{}
	if err = json.Unmarshal(ciphertext, token); err != nil {
		return nil
	}
	return token
}
