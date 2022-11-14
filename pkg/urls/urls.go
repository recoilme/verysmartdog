package urls

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DomainHostName convert string to url and make request,
// return domainUrl and hostname without www, after redirects
func DomainHostName(link string) (domainUrl, hostname string, err error) {
	if !strings.HasPrefix(link, "http") {
		link = "http://" + link
	}
	domainUrl, err = Head(link, time.Second*1)
	if err != nil {
		log.Println(domainUrl, err)
		return
	}
	uri, err := url.Parse(domainUrl)
	if err != nil {
		return
	}
	return uri.Scheme + "://" + uri.Hostname(), strings.TrimPrefix(uri.Hostname(), "www."), nil
}

// Head make get request with empty body and return redirect url
func Head(link string, timeout time.Duration) (string, error) {
	//log.Println("head", link)
	if timeout == 0 {
		timeout = time.Second * 3
	}
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*3)
	defer cncl()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.Request.URL.String(), nil
	}
	return "", errors.New(fmt.Sprintf("Status code:%d", resp.StatusCode))
}

func IsUrlValid(link string) bool {
	_, err := Head(link, 0)
	if err != nil {
		return false
	}
	return true
}
