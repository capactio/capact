package action

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

// TODO: For Create and Update for Action CR:
// Validate the list of input TypeInstances with validation webhook,
// to make sure there are no TypeInstances with duplicated names and different IDs

func (s *Service) Create(ctx context.Context, item model.ActionToCreateOrUpdate) error {
	log := s.logWithNameAndNs(item.Action.Name, item.Action.Namespace)

	log.Info("Creating Action")
	err := s.k8sCli.Create(ctx, &item.Action)
	if err != nil {
		errContext := "while creating Action"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
	}

	err = s.createInputParamsSecretIfShould(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Update(ctx context.Context, item model.ActionToCreateOrUpdate) error {
	oldAction, err := s.GetByName(ctx, item.Action.Name)
	if err != nil {
		return err
	}

	newAction := oldAction.DeepCopy()
	newAction.Spec = item.Action.Spec

	// update action with all fields filled from K8s API
	item.Action = *newAction

	err = s.updateAction(ctx, *newAction)
	if err != nil {
		return err
	}

	err = s.deleteInputParamsSecretIfShould(ctx, oldAction)
	if err != nil {
		return err
	}

	err = s.createInputParamsSecretIfShould(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetByName(ctx context.Context, name string) (v1alpha1.Action, error) {
	objKey, err := s.objectKey(ctx, name)
	if err != nil {
		return v1alpha1.Action{}, err
	}

	log := s.logWithNameAndNs(objKey.Name, objKey.Namespace)
	log.Info("Finding Action by name")

	var item v1alpha1.Action
	err = s.k8sCli.Get(ctx, objKey, &item)
	if err != nil {
		errContext := "while getting item"
		switch {
		case apierrors.IsNotFound(err):
			log.Debug(errContext, zap.Error(ErrActionNotFound))
			return v1alpha1.Action{}, errors.Wrap(ErrActionNotFound, errContext)
		default:
			log.Error(errContext, zap.Error(err))
			return v1alpha1.Action{}, errors.Wrap(err, errContext)
		}
	}

	return item, nil
}

func (s *Service) List(ctx context.Context, filter model.ActionFilter) ([]v1alpha1.Action, error) {
	ns, err := namespace.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while reading namespace from context")
	}

	log := s.log.With(zap.String("namespace", ns))
	log.Info("Listing Actions")

	var itemList v1alpha1.ActionList

	err = s.k8sCli.List(ctx, &itemList, &client.ListOptions{Namespace: ns})
	if err != nil {
		errContext := "while listing Actions"
		log.Error(errContext, zap.Error(err))
		return nil, errors.Wrap(err, errContext)
	}

	if filter.Phase == nil {
		return itemList.Items, nil
	}

	// field selectors for CRDs are not supported (apart from name and namespace)
	log.Info("Filtering Actions", zap.String("status.phase", string(*filter.Phase)))

	var filteredItems []v1alpha1.Action
	for _, item := range itemList.Items {
		if item.Status.Phase != *filter.Phase {
			continue
		}

		filteredItems = append(filteredItems, item)
	}

	return filteredItems, nil
}

func (s *Service) DeleteByName(ctx context.Context, name string) error {
	item, err := s.GetByName(ctx, name)
	if err != nil {
		return err
	}

	log := s.logWithNameAndNs(item.Name, item.Namespace)
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
	item, err := s.GetByName(ctx, name)
	if err != nil {
		return err
	}

	log := s.logWithNameAndNs(item.Name, item.Namespace)

	if item.Spec.IsCanceled() {
		log.Info("Action already canceled, so it cannot be run")
		return ErrActionCanceledNotRunnable
	}

	if item.Spec.IsRun() {
		log.Info("Action already run")
		return nil
	}

	item.Spec.Run = ptr.Bool(true)

	err = s.updateAction(ctx, item)
	return err
}

func (s *Service) CancelByName(ctx context.Context, name string) error {
	item, err := s.GetByName(ctx, name)
	if err != nil {
		return err
	}

	log := s.logWithNameAndNs(item.Name, item.Namespace)

	if item.Spec.IsCanceled() {
		log.Info("Action already canceled")
		return nil
	}

	// TODO: Validate it using validation webhook
	if !item.Spec.IsRun() {
		log.Info("Action not run, so it cannot be canceled")
		return ErrActionNotCancelable
	}

	item.Spec.Cancel = ptr.Bool(true)
	item.Spec.Run = ptr.Bool(false)

	err = s.updateAction(ctx, item)
	return err
}

func (s *Service) ContinueAdvancedRendering(ctx context.Context, actionName string, in model.AdvancedModeContinueRenderingInput) error {
	item, err := s.GetByName(ctx, actionName)
	if err != nil {
		return err
	}

	if !item.Spec.IsAdvancedRenderingEnabled() {
		return ErrActionAdvancedRenderingDisabled
	}

	if item.Status.Phase != v1alpha1.AdvancedModeRenderingIterationActionPhase ||
		item.Status.Rendering == nil ||
		item.Status.Rendering.AdvancedRendering == nil ||
		item.Status.Rendering.AdvancedRendering.RenderingIteration == nil {
		return ErrActionAdvancedRenderingIterationNotContinuable
	}

	err = s.validateInputTypeInstancesForRenderingIteration(
		item.Status.Rendering.AdvancedRendering.RenderingIteration.InputTypeInstancesToProvide,
		in.TypeInstances,
	)
	if err != nil {
		return err
	}

	if in.TypeInstances != nil {
		// merge input TypeInstances
		if item.Spec.Input == nil {
			item.Spec.Input = &v1alpha1.ActionInput{}
		}
		item.Spec.Input.TypeInstances = s.mergeTypeInstances(item.Spec.Input.TypeInstances, in.TypeInstances)
	}

	// continue rendering
	item.Spec.AdvancedRendering.RenderingIteration = &v1alpha1.RenderingIteration{
		ApprovedIterationName: item.Status.Rendering.AdvancedRendering.RenderingIteration.CurrentIterationName,
	}

	err = s.updateAction(ctx, item)
	return err
}

func (s *Service) validateInputTypeInstancesForRenderingIteration(optionalTypeInstancesToProvide *[]v1alpha1.InputTypeInstanceToProvide, providedTypeInstances *[]v1alpha1.InputTypeInstance) error {
	if providedTypeInstances == nil {
		return nil
	}

	if optionalTypeInstancesToProvide == nil {
		optionalTypeInstancesToProvide = &[]v1alpha1.InputTypeInstanceToProvide{}
	}

	// prepare a map for provided TypeInstances
	providedTypeInstancesMap := make(map[string]struct{})
	for _, providedTypeInstance := range *providedTypeInstances {
		providedTypeInstancesMap[providedTypeInstance.Name] = struct{}{}
	}

	// prepare a map for optional TypeInstances
	optionalTypeInstancesToProvideMap := make(map[string]struct{})
	for _, optionalTypeInstance := range *optionalTypeInstancesToProvide {
		optionalTypeInstancesToProvideMap[optionalTypeInstance.Name] = struct{}{}
	}

	// check if all provided TypeInstances are in the set of optional TypeInstances to provide
	var invalidTypeInstanceNames []string
	for key := range providedTypeInstancesMap {
		if _, ok := optionalTypeInstancesToProvideMap[key]; !ok {
			invalidTypeInstanceNames = append(invalidTypeInstanceNames, key)
		}
	}

	if len(invalidTypeInstanceNames) > 0 {
		return NewErrInvalidSetOfTypeInstancesForRenderingIteration(invalidTypeInstanceNames)
	}

	return nil
}

func (s *Service) createInputParamsSecretIfShould(ctx context.Context, item model.ActionToCreateOrUpdate) error {
	if item.InputParamsSecret == nil {
		// no secret to create
		return nil
	}

	log := s.logWithNameAndNs(item.Action.Name, item.Action.Namespace)

	owner := item.Action
	secret := item.InputParamsSecret
	secret.SetOwnerReferences([]v1.OwnerReference{
		{
			APIVersion: v1alpha1.GroupVersion.Identifier(),
			Kind:       v1alpha1.ActionKind,
			Name:       owner.Name,
			UID:        owner.UID,
		},
	})

	log.Info("Creating Secret with input params")
	err := s.k8sCli.Create(ctx, secret)
	if err != nil {
		errContext := "while creating Secret for input parameters"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
	}

	return nil
}

func (s *Service) deleteInputParamsSecretIfShould(ctx context.Context, item v1alpha1.Action) error {
	if item.Spec.Input == nil ||
		item.Spec.Input.Parameters == nil {
		// no secret to delete
		return nil
	}

	log := s.logWithNameAndNs(item.Name, item.Namespace)

	secretToDelete := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      item.Spec.Input.Parameters.SecretRef.Name,
			Namespace: item.Namespace,
		},
	}

	log.Info("Deleting secret with input params")
	err := s.k8sCli.Delete(ctx, &secretToDelete)
	if err != nil {
		errContext := "while deleting Secret for input parameters"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
	}

	return nil
}

func (s *Service) updateAction(ctx context.Context, item v1alpha1.Action) error {
	log := s.logWithNameAndNs(item.Name, item.Namespace)
	log.Info("Updating Action")

	err := s.k8sCli.Update(ctx, &item)
	if err != nil {
		errContext := "while updating item"
		log.Error(errContext, zap.Error(err))
		return errors.Wrap(err, errContext)
	}

	return nil
}

func (s *Service) mergeTypeInstances(slice1, slice2 *[]v1alpha1.InputTypeInstance) *[]v1alpha1.InputTypeInstance {
	if slice1 == nil && slice2 == nil {
		return nil
	}

	var merged []v1alpha1.InputTypeInstance
	if slice1 != nil {
		merged = append(merged, *slice1...)
	}
	if slice2 != nil {
		merged = append(merged, *slice2...)
	}

	return &merged
}

func (s *Service) objectKey(ctx context.Context, name string) (client.ObjectKey, error) {
	ns, err := namespace.FromContext(ctx)
	if err != nil {
		return client.ObjectKey{}, errors.Wrap(err, "while reading namespace from context")
	}

	return client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}, nil
}

func (s *Service) logWithNameAndNs(name, namespace string) *zap.Logger {
	return s.log.With(zap.String("name", name), zap.String("namespace", namespace))
}
