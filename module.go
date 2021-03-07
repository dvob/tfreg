package main

import (
	"context"
)

type Module struct {
	Namespace string
	Name      string
	Provider  string
}

type ModuleProvider interface {
	GetVersions(ctx context.Context, m Module) ([]string, error)
	GetSource(ctx context.Context, m Module, version string) (string, error)
}
