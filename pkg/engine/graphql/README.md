# Engine GraphQL API

The following document describes interaction between client and Engine using GraphQL API.

## Running Action in basic mode

The basic mode of running Action is when user doesn't provide optional artifacts for nested Actions.

1. User creates Action with `createAction` mutation, providing Implementation or Interface path (e.g. `cap.interface.cms.wordpress.install`), input parameters and artifacts (required and optional for the root Action).
1. Engine saves Action details and sets its state to `INITIAL`.
1. Engine detects new Action and changes status to `BEING_RENDERED`.
1. Once Engine resolves all nested Implementations, the status is changed to `READY_TO_RUN`. From now on, user is able run the rendered Action.
1. User (approver) runs the Action with `runAction` mutation.
1. The Action state is `RUNNING`.
1. The Action state changes.
    1. If user cancelled the Action, it changes to `BEING_CANCELLED` and after that to `CANCELLED`. 
    1. If user didn't cancel the action, it changes to `SUCCEEDED` or `FAILED`. 
