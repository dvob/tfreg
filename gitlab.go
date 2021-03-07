package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"text/template"

	"github.com/xanzy/go-gitlab"
)

type gitlabModuleProvider struct {
	client          *gitlab.Client
	mappingTemplate *template.Template
}

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

func (g *gitlabModuleProvider) getProjectID(m Module) (string, error) {
	buf := &bytes.Buffer{}
	err := g.mappingTemplate.Execute(buf, &m)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

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

func (g *gitlabModuleProvider) GetReader(_ context.Context, m Module, version string) (io.Reader, error) {
	projectID, err := g.getProjectID(m)
	if err != nil {
		return nil, err
	}

	tag, _, err := g.client.Tags.GetTag(projectID, "v"+version)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()
	go func() {
		sha := tag.Commit.ID
		format := "tar"

		opts := &gitlab.ArchiveOptions{
			Format: &format,
			SHA:    &sha,
		}
		_, err := g.client.Repositories.StreamArchive(projectID, w, opts)
		if err != nil {
			w.CloseWithError(err)
		}
		w.Close()
	}()

	return gzipReader(tarStripDirReader(r, 1)), nil
}

func gzipReader(r io.Reader) io.Reader {
	newReader, w := io.Pipe()
	gzipWriter := gzip.NewWriter(w)
	go func() {
		_, err := io.Copy(gzipWriter, r)
		if err != nil {
			w.CloseWithError(err)
		}
		gzipWriter.Close()
		w.Close()
	}()
	return newReader
}

func tarStripDirReader(origTar io.Reader, n uint) io.Reader {
	r, w := io.Pipe()
	tr := tar.NewReader(origTar)
	newTar := tar.NewWriter(w)
	go func() {
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				w.CloseWithError(err)
				return
			}

			if hdr.Typeflag == tar.TypeXGlobalHeader {
				_, err := io.Copy(io.Discard, tr)
				if err != nil {
					w.CloseWithError(fmt.Errorf("can't strip %d elements from '%s'", n, hdr.Name))
					return
				}
				continue
			}

			// TODO: not sure if i should use os.PathSeperator here
			parts := strings.Split(hdr.Name, "/")
			if len(parts) <= int(n) {
				w.CloseWithError(fmt.Errorf("can't strip %d elements from '%s'", n, hdr.Name))
				return
			}

			hdr.Name = path.Join(parts[n:]...)

			newTar.WriteHeader(hdr)
			if _, err := io.Copy(newTar, tr); err != nil {
				w.CloseWithError(err)
				return
			}
		}
		newTar.Close()
		w.Close()
	}()
	return r
}
