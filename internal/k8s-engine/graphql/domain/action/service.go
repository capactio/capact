package action

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const actionResourceKind = "Action"

type Service struct {
	k8sCli client.Client
}

func NewService(actionCli client.Client) *Service {
	return &Service{
		k8sCli: actionCli,
	}
}

func (s *Service) Create(ctx context.Context, item model.ActionToCreateOrUpdate) error {
	err := s.k8sCli.Create(ctx, &item.Action)
	if err != nil {
		return errors.Wrap(err, "while creating item")
	}

	if item.InputParamsSecret != nil {
		owner := item.Action
		secret := item.InputParamsSecret
		secret.SetOwnerReferences([]v1.OwnerReference{
			{
				APIVersion: v1alpha1.GroupVersion.Identifier(),
				Kind:       actionResourceKind,
				Name:       owner.Name,
				UID:        owner.UID,
			},
		})

		err = s.k8sCli.Create(ctx, secret)
		if err != nil {
			return errors.Wrap(err, "while creating secret for input parameters")
		}
	}

	return nil
}

func (s *Service) updateAction(ctx context.Context, item v1alpha1.Action) error {
	err := s.k8sCli.Update(ctx, &item)
	if err != nil {
		return errors.Wrap(err, "while updating item")
	}

	return nil
}

func (s *Service) FindByName(ctx context.Context, name string) (v1alpha1.Action, error) {
	objKey, err := s.objectKey(ctx, name)
	if err != nil {
		return v1alpha1.Action{}, err
	}

	var item v1alpha1.Action
	err = s.k8sCli.Get(ctx, objKey, &item)
	if err != nil {
		errToReturn := err
		if apierrors.IsNotFound(err) {
			errToReturn = ErrActionNotFound
		}

		return v1alpha1.Action{}, errors.Wrap(errToReturn, "while getting item")
	}

	return item, nil
}

func (s *Service) List(ctx context.Context) ([]v1alpha1.Action, error) {
	ns, err := namespace.ReadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while reading namespace from context")
	}

	var itemList v1alpha1.ActionList
	err = s.k8sCli.List(ctx, &itemList, &client.ListOptions{
		Namespace: ns,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while listing items")
	}

	return itemList.Items, nil
}

func (s *Service) DeleteByName(ctx context.Context, name string) error {
	item, err := s.FindByName(ctx, name)
	if err != nil {
		return err
	}

	err = s.k8sCli.Delete(ctx, &item)
	if err != nil {
		return errors.Wrap(err, "while deleting item")
	}

	return nil
}

func (s *Service) RunByName(ctx context.Context, name string) error {
	item, err := s.FindByName(ctx, name)
	if err != nil {
		return err
	}

	if item.Spec.Run != nil && *item.Spec.Run {
		return nil
	}

	item.Spec.Run = ptr.Bool(true)

	err = s.updateAction(ctx, item)
	return err
}

func (s *Service) CancelByName(ctx context.Context, name string) error {
	item, err := s.FindByName(ctx, name)
	if err != nil {
		return err
	}

	if item.Spec.Cancel != nil && *item.Spec.Cancel {
		return nil
	}

	item.Spec.Cancel = ptr.Bool(true)

	err = s.updateAction(ctx, item)
	return err
}

func (s *Service) objectKey(ctx context.Context, name string) (client.ObjectKey, error) {
	ns, err := namespace.ReadFromContext(ctx)
	if err != nil {
		return client.ObjectKey{}, errors.Wrap(err, "while reading namespace from context")
	}

	return client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}, nil
}
