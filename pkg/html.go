package pkg

import (
	"errors"
	"io"

	"golang.org/x/net/html"
)

// source https://siongui.github.io/2016/05/10/go-get-html-title-via-net-html/

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node) (string, bool) {
	if isTitleElement(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func GetHtmlTitle(r io.Reader) (string, bool, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", false, err
	}

	t, ok := traverse(doc)

	if !ok {
		return t, ok, errors.New("Title not found!")
	}

	return t, ok, nil
}
