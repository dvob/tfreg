package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/xanzy/go-gitlab"
)

type gitlabModuleProvider struct {
	client          *gitlab.Client
	mappingTemplate *template.Template
}

// newGitlabModuleProvider creates a new module provider for Gitlab. If all
// repositories you want to provide are public you can leave the token empty.
func newGitlabModuleProvider(token, tmplStr string) (*gitlabModuleProvider, error) {
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("test").Parse(tmplStr)
	if err != nil {
		return nil, err
	}

	return &gitlabModuleProvider{
		client:          client,
		mappingTemplate: tmpl,
	}, nil
}

// GetVersions returns all tags from a project which starts with v (e.g v0.0.1).
func (g *gitlabModuleProvider) GetVersions(_ context.Context, m Module) ([]string, error) {
	projectID, err := g.getProjectID(m)
	if err != nil {
		return nil, err
	}

	tags, _, err := g.client.Tags.ListTags(projectID, &gitlab.ListTagsOptions{})
	if err != nil {
		return nil, err
	}

	versions := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag.Name, "v") {
			versions = append(versions, tag.Name[1:])
		}
	}
	return versions, nil
}

// GetSource returns the source of a certain version.
// See https://www.terraform.io/docs/language/modules/sources.html#generic-git-repository for details.
func (g *gitlabModuleProvider) GetSource(_ context.Context, m Module, version string) (string, error) {
	projectID, err := g.getProjectID(m)
	if err != nil {
		return "", err
	}
	source := fmt.Sprintf("git::https://gitlab.com/%s.git?ref=v%s", projectID, version)
	return source, nil
}

// getProjectID maps a module to a Gitlab project ID.
func (g *gitlabModuleProvider) getProjectID(m Module) (string, error) {
	buf := &bytes.Buffer{}
	err := g.mappingTemplate.Execute(buf, &m)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
