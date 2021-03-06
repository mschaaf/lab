package commands

import (
	"bytes"
	"strconv"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/lighttiger2505/lab/commands/internal"
	"github.com/lighttiger2505/lab/internal/api"
	"github.com/lighttiger2505/lab/internal/gitutil"
	"github.com/lighttiger2505/lab/internal/ui"
	"github.com/ryanuber/columnize"
	gitlab "github.com/xanzy/go-gitlab"
)

type UserCommandOption struct {
	ProjectProfileOption *internal.ProjectProfileOption `group:"Project, Profile Options"`
	ListOption           *ListUserOption                `group:"List Options"`
}

func newUserOptionParser(opt *UserCommandOption) *flags.Parser {
	opt.ProjectProfileOption = &internal.ProjectProfileOption{}
	opt.ListOption = newListUserOption()
	parser := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	parser.Usage = `user - list a user

Synopsis:
  # List user
  lab user [-n <num>] [--search=<search word>] [-A]`
	return parser
}

type ListUserOption struct {
	Num        int    `short:"n" long:"num" value-name:"<num>" default:"20" default-mask:"20" description:"Limit the number of search to output."`
	Search     string `short:"s" long:"search" value-name:"<search word>" description:"Search for specific users"`
	AllProject bool   `short:"A" long:"all-project" description:"Print the user of all projects"`
}

func newListUserOption() *ListUserOption {
	return &ListUserOption{}
}

type UserCommand struct {
	UI              ui.UI
	RemoteCollecter gitutil.Collecter
	ClientFactory   api.APIClientFactory
}

func (c *UserCommand) Synopsis() string {
	return "List user"
}

func (c *UserCommand) Help() string {
	var opt UserCommandOption
	userCommnadOptionParser := newUserOptionParser(&opt)
	buf := &bytes.Buffer{}
	userCommnadOptionParser.WriteHelp(buf)
	return buf.String()
}

func (c *UserCommand) Run(args []string) int {
	var opt UserCommandOption
	userCommnadOptionParser := newUserOptionParser(&opt)
	if _, err := userCommnadOptionParser.ParseArgs(args); err != nil {
		c.UI.Error(err.Error())
		return ExitCodeError
	}

	pInfo, err := c.RemoteCollecter.CollectTarget(
		opt.ProjectProfileOption.Project,
		opt.ProjectProfileOption.Profile,
	)
	if err != nil {
		c.UI.Error(err.Error())
		return ExitCodeError
	}

	if err := c.ClientFactory.Init(pInfo.ApiUrl(), pInfo.Token); err != nil {
		c.UI.Error(err.Error())
		return ExitCodeError
	}
	client := c.ClientFactory.GetUserClient()

	listOpt := opt.ListOption
	var result string
	if opt.ListOption.AllProject {
		users, err := client.Users(
			makeUsersOption(listOpt),
		)
		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeError
		}
		result = columnize.SimpleFormat(userOutput(users))
	} else {
		users, err := client.ProjectUsers(
			pInfo.Project,
			makeProjectUsersOption(listOpt),
		)
		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeError
		}
		result = columnize.SimpleFormat(projectUserOutput(users))
	}

	c.UI.Message(result)

	return ExitCodeOK
}

func makeProjectUsersOption(opt *ListUserOption) *gitlab.ListProjectUserOptions {
	listOption := &gitlab.ListOptions{
		Page:    1,
		PerPage: opt.Num,
	}
	listUserOption := &gitlab.ListProjectUserOptions{
		ListOptions: *listOption,
		Search:      gitlab.String(opt.Search),
	}
	return listUserOption
}

func makeUsersOption(opt *ListUserOption) *gitlab.ListUsersOptions {
	listOption := &gitlab.ListOptions{
		Page:    1,
		PerPage: opt.Num,
	}
	listProjectUserOptions := &gitlab.ListUsersOptions{
		ListOptions: *listOption,
		Search:      gitlab.String(opt.Search),
	}
	return listProjectUserOptions
}

func userOutput(users []*gitlab.User) []string {
	var outputs []string
	for _, user := range users {
		output := strings.Join([]string{
			strconv.Itoa(user.ID),
			user.Name,
			user.Username,
		}, "|")
		outputs = append(outputs, output)
	}
	return outputs
}
func projectUserOutput(users []*gitlab.ProjectUser) []string {
	var outputs []string
	for _, user := range users {
		output := strings.Join([]string{
			strconv.Itoa(user.ID),
			user.Name,
			user.Username,
		}, "|")
		outputs = append(outputs, output)
	}
	return outputs
}
