package action

import (
	"context"

	"go.uber.org/zap"

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
	log    *zap.Logger
	k8sCli client.Client
}

func NewService(log *zap.Logger, actionCli client.Client) *Service {
	return &Service{
		log:    log.With(zap.String("module", "service")),
		k8sCli: actionCli,
	}
}

func (s *Service) Create(ctx context.Context, item model.ActionToCreateOrUpdate) error {
	log := s.log.With(zap.Any("ActionToCreate", item))

	log.Info("Creating Action")
	err := s.k8sCli.Create(ctx, &item.Action)
	if err != nil {
		errContext := "while creating Action"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
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

		log.Info("Creating Secret",
			zap.Any("secret", secret),
		)
		err = s.k8sCli.Create(ctx, secret)
		if err != nil {
			errContext := "while creating Secret for input parameters"
			log.Error(errContext, zap.Error(err))
			return errors.Wrap(err, errContext)
		}
	}

	return nil
}

func (s *Service) updateAction(ctx context.Context, item v1alpha1.Action) error {
	log := s.log.With(zap.Any("Action", item))
	log.Info("Updating Action")

	err := s.k8sCli.Update(ctx, &item)
	if err != nil {
		errContext := "while updating item"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
	}

	return nil
}

func (s *Service) FindByName(ctx context.Context, name string) (v1alpha1.Action, error) {
	objKey, err := s.objectKey(ctx, name)
	if err != nil {
		return v1alpha1.Action{}, err
	}

	log := s.log.With(zap.Any("objectKey", objKey))
	log.Info("Finding Action by name")

	var item v1alpha1.Action
	err = s.k8sCli.Get(ctx, objKey, &item)
	if err != nil {
		errContext := "while getting item"
		errToReturn := err
		if apierrors.IsNotFound(err) {
			errToReturn = ErrActionNotFound
			log.Info("Item not found")
		} else {
			log.Error(errContext, zap.Error(errToReturn))
		}

		return v1alpha1.Action{}, errors.Wrap(errToReturn, errContext)
	}

	return item, nil
}

func (s *Service) List(ctx context.Context) ([]v1alpha1.Action, error) {
	ns, err := namespace.ReadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while reading namespace from context")
	}

	log := s.log.With(zap.String("namespace", ns))
	log.Info("Listing Actions")

	var itemList v1alpha1.ActionList
	err = s.k8sCli.List(ctx, &itemList, &client.ListOptions{
		Namespace: ns,
	})
	if err != nil {
		errContext := "while listing Actions"
		log.Error(errContext, zap.Error(err))
		return nil, errors.Wrap(err, errContext)
	}

	return itemList.Items, nil
}

func (s *Service) DeleteByName(ctx context.Context, name string) error {
	item, err := s.FindByName(ctx, name)
	if err != nil {
		return err
	}

	log := s.log.With(zap.Any("Action", item))
	log.Info("Deleting Action by name")

	err = s.k8sCli.Delete(ctx, &item)
	if err != nil {
		errContext := "while deleting item"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
	}

	return nil
}

func (s *Service) RunByName(ctx context.Context, name string) error {
	item, err := s.FindByName(ctx, name)
	if err != nil {
		return err
	}

	log := s.log.With(zap.Any("action", item))

	if item.Spec.Run != nil && *item.Spec.Run {
		log.Info("Action already run")
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

	log := s.log.With(zap.Any("action", item))

	if item.Spec.Cancel != nil && *item.Spec.Cancel {
		log.Info("Action already cancelled")
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
