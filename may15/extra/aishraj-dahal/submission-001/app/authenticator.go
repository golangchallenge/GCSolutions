// +build !appengine

package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aishraj/gopherlisa/common"
)

func AuthroizeHandler(context *common.AppContext, w http.ResponseWriter, r *http.Request) (revVal int, err error) {
	u, err := url.Parse(common.OauthBaseURI)
	u.Path = "/oauth/authorize/"
	if err != nil {
		context.Log.Println("Unable to parse the URI. Error is: ", err)
		return http.StatusBadRequest, err
	}
	query := u.Query()
	query.Set("client_id", common.InstagramClientID)
	query.Set("redirect_uri", common.RedirectURI)
	query.Set("response_type", "code")

	u.RawQuery = query.Encode()
	context.Log.Println("Query is: ", u)
	http.Redirect(w, r, fmt.Sprintf("%v", u), http.StatusSeeOther) //TODO change this so that redirect happens in the calling method.
	return http.StatusOK, nil                                      //TODO change this
}

func GetAuthToken(applicationContext *common.AppContext, w http.ResponseWriter, req *http.Request, code string) (token AuthenticationResponse, err error) {

	applicationContext.Log.Printf("Performing Post trigggered with the code value %v \n", code)

	uri, err := url.ParseRequestURI(common.OauthBaseURI)

	if err != nil {
		return token, err
	}

	uri.Path = "/oauth/access_token/"
	data := url.Values{}
	data.Set("client_id", common.InstagramClientID)
	data.Add("client_secret", common.InstagramSecret)
	data.Add("grant_type", "authorization_code")
	data.Add("redirect_uri", common.RedirectURI)
	data.Add("code", code)

	urlStr := fmt.Sprintf("%v", uri)

	applicationContext.Log.Print("posting to the url: ", urlStr)

	client := &http.Client{}

	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode())) // <-- URL-encoded payload
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, err := client.Do(r)

	if err != nil {
		applicationContext.Log.Println("Unable to send the post request with the code")
		return token, err

	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			applicationContext.Log.Println("Could not parse the error resposne received from instagram.")
			return token, errors.New("Did not get a success while posting on instagram and Could not parse the error resposne received from instagram.")
		}
		applicationContext.Log.Println("Did not get 200 for the post authn request. Response was: ", string(contents))
		return token, errors.New("Did not get a success while posting on instagram")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		applicationContext.Log.Println("Unable to parse HTTP response to body")
		return token, errors.New("Did not get a success while posting on instagram")
	}

	var authToken AuthenticationResponse

	err = json.Unmarshal(body, &authToken)
	if err != nil {
		applicationContext.Log.Printf("%T\n%s\n%#v\n", err, err, err)
		switch v := err.(type) {
		case *json.SyntaxError:
			applicationContext.Log.Println(string(body[v.Offset-40 : v.Offset]))
		}
		applicationContext.Log.Println("Eror while unmarshalling data.")
		return token, errors.New("Error unmarshalling data")
	}

	applicationContext.Log.Printf("Yippie!! your authentication token is %v \n", authToken.AccessToken)
	applicationContext.Log.Println("Got the data: ", authToken)
	return authToken, nil
}
