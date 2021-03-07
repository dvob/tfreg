package main

import (
	"context"
	"io"
	"os"
)

type Module struct {
	Namespace string
	Name      string
	Provider  string
}

type ModuleProvider interface {
	GetVersions(ctx context.Context, m Module) ([]string, error)
	GetReader(ctx context.Context, m Module, version string) (io.Reader, error)
}

type dummyModuleProvider struct{}

func (d *dummyModuleProvider) GetReader(_ context.Context, _ Module, _ string) (io.Reader, error) {
	return os.Open("mymod.tar.gz")
}

func (d *dummyModuleProvider) GetVersions(_ context.Context, m Module) ([]string, error) {
	return []string{"0.1.2"}, nil
}
