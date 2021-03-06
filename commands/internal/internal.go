package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/lighttiger2505/lab/internal/browse"
)

type Method interface {
	Process() (string, error)
}

type BrowseMethod struct {
	Opener browse.URLOpener
	Opt    *BrowseOption
	URL    string
	ID     int
}

func (m *BrowseMethod) Process() (string, error) {
	url := m.URL
	if m.ID > 0 {
		url = strings.Join([]string{url, strconv.Itoa(m.ID)}, "/")
	}

	if m.Opt.Browse {
		if err := m.Opener.Open(url); err != nil {
			return "", err
		}
	}

	if m.Opt.Copy {
		if err := clipboard.WriteAll(url); err != nil {
			return "", fmt.Errorf(fmt.Sprintf("Error copying %s to clipboard:\n%s\n", url, err))
		}
	}

	if m.Opt.URL {
		return url, nil
	}

	// Return empty value
	return "", nil
}

type MockMethod struct{}

func (m *MockMethod) Process() (string, error) {
	return "result", nil
}

func SweepMarkdownComment(text string) string {
	r := regexp.MustCompile("<!--[\\s\\S]*?-->[\\n]*")
	return r.ReplaceAllString(text, "")
}

func ParceRepositoryFullName(webURL string) string {
	splitURL := strings.Split(webURL, "/")[3:]

	subPageWords := []string{
		"issues",
		"merge_requests",
	}
	var subPageIndex int
	for i, word := range splitURL {
		for _, subPageWord := range subPageWords {
			if word == subPageWord {
				subPageIndex = i
			}
		}
	}

	return strings.Join(splitURL[:subPageIndex], "/")
}
