# Engine GraphQL API

The following document describes interaction between client and Engine using GraphQL API. 
The Engine GraphQL API schema is located in [pkg/engine/api/graphql/schema.graphql](../pkg/engine/api/graphql/schema.graphql) file.

## Examples

To run sample GraphQL queries and mutations against Engine GraphQL API, follow the steps:

1. Open the Voltron Gateway GraphQL Playground. 
   
   For local installation, it is available under [`https://gateway.voltron.local`](https://gateway.voltron.local) URL.

1. Copy and paste the [pkg/engine/api/graphql/examples.graphql](../pkg/engine/api/graphql/examples.graphql) file content to the GraphQL Playground IDE.
1. Click on the "Query Variables" tab.
1. Copy and paste the [pkg/engine/api/graphql/examples.variables.json](../pkg/engine/api/graphql/examples.variables.json) file content to the Query Variables section of the GraphQL Playground IDE.
1. Run any query or mutation from the list.

## Common flows

### Rendering Action in basic mode

The basic mode of rendering Action is when user doesn't provide optional TypeInstances for nested Actions.

1. User creates Action with `createAction` mutation, providing Implementation or Interface path (e.g. `cap.interface.cms.wordpress.install`), input parameters and TypeInstances (required and optional for the root Action).
1. Engine saves Action details and sets the Action status to `INITIAL`.
1. Engine detects newly created Action and changes the Action status to `BEING_RENDERED`.
1. Once Engine resolves all nested Implementations, it changes the Action status to `READY_TO_RUN`. From now on, user is able run the rendered Action.

### Rendering Action in advanced mode

The advanced mode of rendering Action is when user can provide optional TypeInstances for every nested Action.

1. User creates Action with `createAction` mutation, providing Implementation or Interface path (e.g. `cap.interface.cms.wordpress.install`), input parameters and TypeInstances (required and optional for the root Action).
1. Engine saves Action details and sets the Action status to `INITIAL`.
1. Engine detects newly created Action and changes the Action status to `BEING_RENDERED`.
1. In loop, until Engine resolves all nested Actions:
    1. Engine resolves nested Action. If there are optional TypeInstances specified that can be provided, Engine changes status of the Action to `ADVANCED_MODE_RENDERING_ITERATION`.
    1. User fetches Action with `action(id)` query and checks optional TypeInstances which can be provided in the iteration under `Action.renderingAdvancedMode.typeInstancesForRenderingIteration`.
    1. User continues Action rendering with `continueAdvancedRendering` mutation. In the mutation input, user can specify optional TypeInstances for a given rendering iteration.
    1. Engine change status of the Action to `BEING_RENDERED`.
1. Once Engine resolves all nested Actions, the status changes to `READY_TO_RUN`. From now on, user is able run the rendered Action.

### Running Action

Once Action is in `READY_TO_RUN` mode, user can run the Action, or, in other words, approve the rendered Action to run.

1. User (Action approver) runs the Action with `runAction` mutation.
1. Engine changes the Action status to `RUNNING`.
1. Depending on what happens, Engine changes the Action status:
    1. If user canceled the Action, it changes to `BEING_CANCELED` and after that to `CANCELED`. 
    1. If user didn't cancel the action, it changes to `SUCCEEDED` or `FAILED`. 

## Implementation Specific Behavior

This section describes GraphQL API server behaviors which is a result of underlying implementation.
The API consumer should be aware of it, in order to use the API efficiently.

### Kubernetes Engine

Engine GraphQL API is a server which does CRUD operations on Kubernetes resources. As the Engine with actual business logic is implemented using the operator pattern, some fields may not be resolved instantly until controller processes the user request.

For example, User input (`Action.input` field - both parameters and TypeInstances) is returned from GraphQL API once the Engine resolves all details regarding it. It means that the newly created or updated Action (via `createAction` or `updateAction` mutations) has to be processed by controller until the data is available. Although usually it does take less than a second, API consumer should be aware that he/she may not be able to get all details as a mutation result.    
