# Engine GraphQL API

The following document describes interaction between client and Engine using GraphQL API.

## Rendering Action in basic mode

The basic mode of rendering Action is when user doesn't provide optional artifacts for nested Actions.

1. User creates Action with `createAction` mutation, providing Implementation or Interface path (e.g. `cap.interface.cms.wordpress.install`), input parameters and artifacts (required and optional for the root Action).
1. Engine saves Action details and sets the Action status to `INITIAL`.
1. Engine detects newly created Action and changes the Action status to `BEING_RENDERED`.
1. Once Engine resolves all nested Implementations, it changes the Action status to `READY_TO_RUN`. From now on, user is able run the rendered Action.


## Rendering Action in advanced mode

The advanced mode of rendering Action is when user can provide optional artifacts for every nested Action.

1. User creates Action with `createAction` mutation, providing Implementation or Interface path (e.g. `cap.interface.cms.wordpress.install`), input parameters and artifacts (required and optional for the root Action).
1. Engine saves Action details and sets the Action status to `INITIAL`.
1. Engine detects newly created Action and changes the Action status to `BEING_RENDERED`.
1. In loop, until Engine resolves all nested Actions:
    1. Engine resolves nested Action. If there are optional artifacts specified that can be provided, Engine changes status of the Action to `ADVANCED_MODE_RENDERING_ITERATION`.
    1. User fetches Action with `action(id)` query and checks optional artifacts which can be provided in the iteration under `Action.renderingAdvancedMode.artifactsForRenderingIteration`.
    1. User continues Action rendering with `continueAdvancedRendering` mutation. In the mutation input, user can specify optional artifacts for a given rendering iteration.
    1. Engine change status of the Action to `BEING_RENDERED`.
1. Once Engine resolves all nested Actions, the status changes to `READY_TO_RUN`. From now on, user is able run the rendered Action.


## Running Action

Once Action is in `READY_TO_RUN` mode, user can run the Action, or, in other words, approve the rendered Action to run.

1. User (approver) runs the Action with `runAction` mutation.
1. Engine changes the Action status to `RUNNING`.
1. Depending on what happens, Engine changes the Action status:
    1. If user cancelled the Action, it changes to `BEING_CANCELLED` and after that to `CANCELLED`. 
    1. If user didn't cancel the action, it changes to `SUCCEEDED` or `FAILED`. 
