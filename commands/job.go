package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

type JobCommandOption struct {
	ProjectProfileOption *internal.ProjectProfileOption `group:"Project, Profile Options"`
	ListOption           *ListJobOption                 `group:"List Options"`
}

func newJobOptionParser(opt *JobCommandOption) *flags.Parser {
	opt.ProjectProfileOption = &internal.ProjectProfileOption{}
	opt.ListOption = newListJobOption()
	parser := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	parser.Usage = "job [options]"
	parser.Usage = `issue - list a job

Synopsis:
  # List job
  lab issue [-n <num>] [--search=<search word>] [-A]`
	return parser
}

type ListJobOption struct {
	Num int  `short:"n" long:"num" value-name:"<num>" default:"20" default-mask:"20" description:"Limit the number of search to output."`
	Log bool `short:"t" long:"log" description:"Get a trace of a specific job of a project."`
	// Scope string `long:"scope" value-name:"<scope>" default:"all" default-mask:"all" description:"Print only given scope. created, pending, running, failed, success, canceled, skipped, manual"`
}

func newListJobOption() *ListJobOption {
	return &ListJobOption{}
}

type JobCommand struct {
	UI              ui.UI
	RemoteCollecter gitutil.Collecter
	ClientFactory   api.APIClientFactory
}

func (c *JobCommand) Synopsis() string {
	return "List job"
}

func (c *JobCommand) Help() string {
	var opt JobCommandOption
	jobCommnadOptionParser := newJobOptionParser(&opt)
	buf := &bytes.Buffer{}
	jobCommnadOptionParser.WriteHelp(buf)
	return buf.String()
}

func (c *JobCommand) Run(args []string) int {
	// Parse flags
	var opt JobCommandOption
	jobCommnadOptionParser := newJobOptionParser(&opt)
	parseArgs, err := jobCommnadOptionParser.ParseArgs(args)
	if err != nil {
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
	client := c.ClientFactory.GetJobClient()

	listOpt := opt.ListOption

	if len(parseArgs) > 0 {
		jid, err := strconv.Atoi(parseArgs[0])
		if err != nil {
			c.UI.Error(fmt.Sprintf("Invalid job id. value: %s, error: %s", parseArgs[0], err))
		}

		if listOpt.Log {
			trace, err := client.GetTraceFile(pInfo.Project, jid)
			if err != nil {
				c.UI.Error(err.Error())
				return ExitCodeError
			}

			b, err := ioutil.ReadAll(trace)
			if err != nil {
				c.UI.Error(err.Error())
				return ExitCodeError
			}

			c.UI.Message(string(b))
			return ExitCodeOK
		}

		job, err := client.GetJob(pInfo.Project, jid)
		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeError
		}
		c.UI.Message(jobDetailOutput(job))
	} else {
		var result string
		jobs, err := client.GetProjectJobs(
			makeProjectJobsOption(listOpt),
			pInfo.Project,
		)
		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeError
		}
		result = columnize.SimpleFormat(projectJobOutput(jobs))
		c.UI.Message(result)
	}

	return ExitCodeOK
}

func makeProjectJobsOption(opt *ListJobOption) *gitlab.ListJobsOptions {
	listOption := &gitlab.ListOptions{
		Page:    1,
		PerPage: opt.Num,
	}
	listJobOption := &gitlab.ListJobsOptions{
		ListOptions: *listOption,
		// Scope:       gitlab.String(opt.Scope),
	}
	return listJobOption
}

func projectJobOutput(jobs []gitlab.Job) []string {
	var outputs []string
	for _, job := range jobs {
		output := strings.Join([]string{
			strconv.Itoa(job.ID),
			job.Status,
			job.Ref,
			job.Commit.ShortID,
			job.User.Username,
			job.Stage,
			job.Name,
		}, "|")
		outputs = append(outputs, output)
	}
	return outputs
}

func jobDetailOutput(job *gitlab.Job) string {
	base := `%d
Ref: %s
Commit: %s
User: %s
Stage: %s
Name: %s
`
	detial := fmt.Sprintf(
		base,
		job.ID,
		job.Ref,
		job.Commit.ShortID,
		job.User.Username,
		job.Stage,
		job.Name,
	)
	return detial
}
