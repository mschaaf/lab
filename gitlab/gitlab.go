package gitlab

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lighttiger2505/lab/config"
	"github.com/lighttiger2505/lab/git"
	"github.com/lighttiger2505/lab/ui"
	"github.com/xanzy/go-gitlab"
)

type RemoteFilter interface {
	Collect() error
	Filter(ui.Ui, *config.Config) (*git.RemoteInfo, error)
}

type GitlabRemoteFilter struct {
	GitRemotes []git.RemoteInfo
}

func (g *GitlabRemoteFilter) Collect() error {
	// Get remote urls
	gitRemotes, err := git.GitRemotes()
	if err != nil {
		return err
	}
	g.GitRemotes = gitRemotes
	return nil
}

func (g *GitlabRemoteFilter) Filter(ui ui.Ui, conf *config.Config) (*git.RemoteInfo, error) {
	// Filtering only gitlab remote info
	gitlabRemotes := filterHasGitlabDomain(g.GitRemotes)

	// Filter gitlab remote url only
	var gitlabRemote *git.RemoteInfo
	if len(gitlabRemotes) == 1 {
		gitlabRemote = &gitlabRemotes[0]
	} else if len(gitlabRemotes) > 1 {
		var err error
		gitlabRemote, err = selectUseRemote(ui, gitlabRemotes, conf)
		if err != nil {
			return nil, fmt.Errorf("Failed select multi remote repository. %v", err.Error())
		}
	} else {
		// Current directory is not git repository
		return nil, nil
	}
	return gitlabRemote, nil
}

func GitlabRemote(ui ui.Ui, conf *config.Config) (*git.RemoteInfo, error) {
	// Get remote urls
	gitRemotes, err := git.GitRemotes()
	if err != nil {
		return nil, err
	}
	// Filtering only gitlab remote info
	gitlabRemotes := filterHasGitlabDomain(gitRemotes)

	// Filter gitlab remote url only
	var gitlabRemote *git.RemoteInfo
	if len(gitlabRemotes) == 1 {
		gitlabRemote = &gitlabRemotes[0]
	} else if len(gitlabRemotes) > 1 {
		var err error
		gitlabRemote, err = selectUseRemote(ui, gitlabRemotes, conf)
		if err != nil {
			return nil, fmt.Errorf("Failed select multi remote repository. %v", err.Error())
		}
	} else {
		// Current directory is not git repository
		return nil, nil
	}
	return gitlabRemote, nil
}

func filterHasGitlabDomain(remoteInfos []git.RemoteInfo) []git.RemoteInfo {
	var gitlabRemotes []git.RemoteInfo
	for _, remoteInfo := range remoteInfos {
		if strings.HasPrefix(remoteInfo.Domain, "gitlab") {
			gitlabRemotes = append(gitlabRemotes, remoteInfo)
		}
	}
	return gitlabRemotes
}

func selectUseRemote(ui ui.Ui, gitlabRemotes []git.RemoteInfo, conf *config.Config) (*git.RemoteInfo, error) {
	// Search for remote repositorie whose selection is prioritized in the config
	var gitlabRemote *git.RemoteInfo
	gitlabRemote = hasPriorityRemote(gitlabRemotes, conf.PreferredDomains)
	if gitlabRemote == nil {
		// Get remote repository selected by user input
		var err error
		gitlabRemote, err = inputUseRemote(ui, gitlabRemotes)
		if err != nil {
			return nil, fmt.Errorf("Failed choise gitlab remote. %v", err.Error())
		}

		// Add selected remote repository to config
		conf.AddRepository(gitlabRemote.Domain)
		if err := conf.Write(); err != nil {
			return nil, fmt.Errorf("Failed update config of repository priority. %v", err.Error())
		}
	}
	return gitlabRemote, nil
}

func hasPriorityRemote(remoteInfos []git.RemoteInfo, preferredDomains []string) *git.RemoteInfo {
	for _, preferredDomain := range preferredDomains {
		for _, remoteInfo := range remoteInfos {
			if preferredDomain == remoteInfo.Domain {
				return &remoteInfo
			}
		}
	}
	return nil
}

func inputUseRemote(ui ui.Ui, remoteInfos []git.RemoteInfo) (*git.RemoteInfo, error) {
	// Receive number of the domain of the remote repository to be searched from stdin
	ui.Message("That repository existing multi gitlab remote repository.")
	for i, remoteInfo := range remoteInfos {
		ui.Message(fmt.Sprintf("%d) %s", i+1, remoteInfo.Domain))
	}
	text, err := ui.Ask("Please choice target domain :")
	if err != nil {
		return nil, fmt.Errorf("Failed target domain input. %v", err.Error())
	}

	// Check valid number
	choiceNumber, err := strconv.Atoi(text)
	if err != nil {
		return nil, fmt.Errorf("Failed parse number. %v", err.Error())
	}
	if choiceNumber < 1 || choiceNumber > len(remoteInfos) {
		return nil, fmt.Errorf("Invalid numver. %d", choiceNumber)
	}

	gitLabRemote := &remoteInfos[choiceNumber-1]
	return gitLabRemote, nil
}

func NewGitlabClient(ui ui.Ui, gitlabRemote *git.RemoteInfo, conf *config.Config) (*gitlab.Client, error) {
	token, err := getPrivateToken(ui, gitlabRemote.Domain, conf)
	if err != nil {
		return nil, fmt.Errorf("Failed getting private token. %s", err.Error())
	}

	// Create client
	client := gitlab.NewClient(nil, token)
	if err := client.SetBaseURL(gitlabRemote.ApiUrl()); err != nil {
		return nil, fmt.Errorf("Invalid api url. %s", err.Error())
	}
	return client, nil
}

func getPrivateToken(ui ui.Ui, domain string, conf *config.Config) (string, error) {
	token := ""
	for _, mapItem := range conf.Tokens {
		if mapItem.Key.(string) == domain {
			token = mapItem.Value.(string)
		}
	}

	if token == "" {
		token, err := ui.Ask("Please input GitLab private token :")
		if err != nil {
			return "", fmt.Errorf("Failed input private token. %s", err.Error())
		}

		conf.AddToken(domain, token)
		if err := conf.Write(); err != nil {
			return "", fmt.Errorf("Failed update config of private token. %s", err.Error())
		}
	}
	return token, nil
}

func ParceRepositoryFullName(webURL string) string {
	sp := strings.Split(webURL, "/")
	return strings.Join([]string{sp[3], sp[4]}, "/")
}
