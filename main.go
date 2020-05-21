package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type Data struct {
	AppLaunchURI string
	CompleteURI  string
}

func main() {
	launchTmpl := template.Must(template.ParseFiles("launch.html"))
	clientId := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	authURL := os.Getenv("OAUTH_AUTH_URL")
	tokenURL := os.Getenv("OAUTH_TOKEN_URL")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	appLaunchURI := os.Getenv("OAUTH_APP_LAUNCH_URI")
	completeURI := os.Getenv("OAUTH_COMPLETE_URI")
	endpoint := oauth2.Endpoint{
		AuthURL:  authURL,
		TokenURL: tokenURL,
	}
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		config := oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint:     endpoint,
			RedirectURL:  redirectURI,
		}
		q := r.URL.Query()
		state := q.Get("state")
		if v := q["scopes"]; len(v) > 0 {
			config.Scopes = v
		}
		url := config.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	r.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		config := oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint:     endpoint,
			RedirectURL:  redirectURI,
		}
		q := r.URL.Query()
		state := q.Get("state")
		code := q.Get("code")
		token, err := config.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tokenData, _ := json.Marshal(token)
		tokenQuery := url.Values{}
		tokenQuery.Set("token", base64.URLEncoding.EncodeToString(tokenData))
		tokenQuery.Set("state", state)
		data := Data{
			AppLaunchURI: appLaunchURI + "?" + tokenQuery.Encode(),
			CompleteURI:  completeURI,
		}
		launchTmpl.Execute(w, data)
	})

	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
