package pipeline

import (
	"github.com/lighttiger2505/lab/commands/internal"
	"github.com/lighttiger2505/lab/internal/api"
	"github.com/lighttiger2505/lab/internal/browse"
	"github.com/lighttiger2505/lab/internal/gitutil"
)

type MethodFactory interface {
	CreateMethod(opt Option, pInfo *gitutil.GitLabProjectInfo, iid int, factory api.APIClientFactory) internal.Method
}

type PipelineMethodFacotry struct{}

func (c *PipelineMethodFacotry) CreateMethod(opt Option, pInfo *gitutil.GitLabProjectInfo, iid int, factory api.APIClientFactory) internal.Method {
	if opt.BrowseOption.Browse {
		return &browseMethod{
			opener: &browse.Browser{},
			url:    pInfo.SubpageUrl("pipelines"),
			id:     iid,
		}
	}

	if iid > 0 {
		return &listJobMethod{
			client:  factory.GetPipelineClient(),
			opt:     opt.ListOption,
			project: pInfo.Project,
			id:      iid,
		}
	}

	return &listMethod{
		client:  factory.GetPipelineClient(),
		opt:     opt.ListOption,
		project: pInfo.Project,
	}
}

type MockMethodFactory struct{}

func (c *MockMethodFactory) CreateMethod(opt Option, pInfo *gitutil.GitLabProjectInfo, iid int, factory api.APIClientFactory) internal.Method {
	return &internal.MockMethod{}
}
