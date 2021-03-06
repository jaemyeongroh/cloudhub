package mocks

import (
	"context"

	cloudhub "github.com/snetsystems/cloudhub/backend"
)

var _ cloudhub.LayoutsStore = &LayoutsStore{}

type LayoutsStore struct {
	AddF    func(ctx context.Context, layout cloudhub.Layout) (cloudhub.Layout, error)
	AllF    func(ctx context.Context) ([]cloudhub.Layout, error)
	DeleteF func(ctx context.Context, layout cloudhub.Layout) error
	GetF    func(ctx context.Context, id string) (cloudhub.Layout, error)
	UpdateF func(ctx context.Context, layout cloudhub.Layout) error
}

func (s *LayoutsStore) Add(ctx context.Context, layout cloudhub.Layout) (cloudhub.Layout, error) {
	return s.AddF(ctx, layout)
}

func (s *LayoutsStore) All(ctx context.Context) ([]cloudhub.Layout, error) {
	return s.AllF(ctx)
}

func (s *LayoutsStore) Delete(ctx context.Context, layout cloudhub.Layout) error {
	return s.DeleteF(ctx, layout)
}

func (s *LayoutsStore) Get(ctx context.Context, id string) (cloudhub.Layout, error) {
	return s.GetF(ctx, id)
}

func (s *LayoutsStore) Update(ctx context.Context, layout cloudhub.Layout) error {
	return s.UpdateF(ctx, layout)
}
