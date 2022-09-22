// package gmailAPI is used in applications and libraries that access the gmail
// api.
// Most of this code was lightly modified from code found on:
// https://developers.google.com/gmail/api/quickstart/go
package gmailAPI

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	gmail "google.golang.org/api/gmail/v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ConnectToService uses a Context, config and 1 or more scopes to retrieve a
// Token then generate a Client. It uses the client to connect to the gmail api
// service and returns a *gmail.Service.
func ConnectToService(ctx context.Context, scope ...string) *gmail.Service {
	// If you don't have a client_secret.json, go here (as of July 2017):
	// https://auth0.com/docs/connections/social/google
	// https://web.archive.org/web/20170708123613/https://auth0.com/docs/connections/social/google
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	/////////////////////////////////////////////////////////////////////////////
	// SCOPE https://developers.google.com/gmail/api/auth/scopes
	/////////////////////////////////////////////////////////////////////////////
	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	// If modifying this code to access other features of gmail, you may need
	// to change scope.
	config, err := google.ConfigFromJSON(b, scope...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	cacheFile, err := newTokenizer()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(cacheFile, token)
	}
	srv, err := gmail.New(config.Client(ctx, token))
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}
	return srv
}

// newTokenizer returns a new token and generates credential file path and
// returns the generated credential path/filename along with any errors.
func newTokenizer() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, url.QueryEscape("gmail-go-quickstart.json")),
		err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read errors encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// getTokenFromWeb uses Config to request a Token. It returns the retrieved
// Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// saveToken uses a file path to create a file and store the token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
