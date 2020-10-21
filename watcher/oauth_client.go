package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var sharedOauthInstance *OauthClient

type OauthClient struct {
	tokenURL     string
	bitbucketURL string
	appID        string
	secret       string
	oauthConfig  clientcredentials.Config
	token        oauth2.Token
}
type ResponseBody struct {
	Text string `json:"text"`
}

type CommitResponse struct {
	Values []struct {
		Date time.Time `json:"date"`
		Hash string    `json:"hash"`
	} `json:"values"`
}

func GetOauthInstance() *OauthClient {
	if sharedOauthInstance == nil {
		sharedOauthInstance = &OauthClient{}
	}
	return sharedOauthInstance
}

// Get token from bitbucket server
func (o *OauthClient) NewOAuthClientCredentials(tokenURL string, bitbucektURL string, appID string, secret string) (*OauthClient, error) {
	ctx := context.Background()
	conf := &clientcredentials.Config{
		ClientID:     appID,
		ClientSecret: secret,
		TokenURL:     tokenURL,
	}
	tok, err := conf.Token(ctx)
	if err != nil {
		return nil, err
	}
	a := &OauthClient{
		tokenURL:     tokenURL,
		bitbucketURL: bitbucektURL,
		appID:        appID,
		secret:       secret,
		oauthConfig:  *conf,
		token:        *tok,
	}
	return a, nil
}

func (o *OauthClient) GenerateRequestURL(template string, args ...interface{}) string {
	if len(args) == 1 && args[0] == "" {
		return o.bitbucketURL + template
	}
	return o.bitbucketURL + fmt.Sprintf(template, args...)
}

func (o *OauthClient) Get(url string) (*CommitResponse, error) {
	resp, err := o.oauthConfig.Client(oauth2.NoContext).Get(url)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("[oauth]: bad response status code %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	data := new(CommitResponse)
	if err := json.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return data, nil
}
