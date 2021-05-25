# Contributing to Capact

This document describes the process of contribution to this project. Any type of contribution is welcome!

## Table of contents

<!-- toc -->

- [Contribution process](#contribution-process)
  * [Report an issue](#report-an-issue)
  * [Create a pull request](#create-a-pull-request)
- [Support Channels](#support-channels)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

<!-- tocstop -->

## Contribution process

We use GitHub to host code, track issues and feature requests, and accept pull requests.

### Report an issue

To report an issue, follow the steps:

1. Search [open and closed issues](https://github.com/Project-Voltron/voltron/issues?q=is%3Aissue) to see if your issue is not a duplicate. 
1. Navigate to [New issue](https://github.com/Project-Voltron/go-voltron/issues/new) page on the Capact GitHub repository.
1. Select an issue template, which is the most accurate for the issue type you report.
1. Describe the issue clearly according to the selected issue template.

### Create a pull request

To create a new pull request, follow the steps:

1. Make sure an issue related to the change [is reported](#report-an-issue).
1. Fork the repository and configure the fork on your local machine. To learn how to do it, read the [Prepare the fork](./docs/git-workflow.md#prepare-the-fork) section in the **Git workflow** guide.
1. Create a branch from the `main` repository branch. To learn how to do it, follow the [Contribute](./docs/git-workflow.md#contribute) section in the **Git workflow** guide.
1. Do the proposed changes.
   
    - Learn how to [develop to the project](./docs/development.md).
    - Adhere to our [development guidelines](./docs/development-guidelines.md).
    - Make sure the changes are tested locally.
    
    > **NOTE:** Remember that large pull requests with multiple files changed are very difficult to review. Discuss the planned changes upfront in the related issue and consider splitting one large pull request into smaller ones.

1. Commit and push the changes.
   
    To learn how to do it, follow the [Contribute](./docs/git-workflow.md#contribute) section in the **Git workflow** guide. 

1. Create a new pull request on the [Compare changes](https://github.com/Project-Voltron/voltron/compare) page.
    
    Make sure the option [Allow edits from maintainers](https://docs.github.com/en/github/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork) is selected.

### Get your pull request approved

Once you create a pull request:

1. Make sure all automated pull request tests pass. 
   
   To read more, see the [Capact CI and CD](./docs/ci.md) document.

1. Wait for the [repository maintainers](./CODEOWNERS) to review and approve the pull request. 
   
    One or more reviewers are assigned automatically. These reviewers will do a thorough code review, looking for correctness, bugs, opportunities for improvement, documentation and comments, and style. Respond to the review comments and commit changes to the same branch on your fork.

You can contact the reviewers by mentioning GitHub username (`@username`) in the pull request comments.  

## Support Channels

Currently, the only official way to get support is through [GitHub issues](https://github.com/Project-Voltron/go-voltron/issues).
In the future, we will also create a dedicated Slack channel for making it easy to ask a question.

## Code of Conduct

We adopted a Code of Conduct and we expect project participants to adhere to it. To understand what actions will and will not be tolerated, read the [`CODE_OF_CONDUCT.md`](./CODE_OF_CONDUCT.md) document.

## License

By contributing to Capact, you agree that your contributions will be licensed under [the same license that covers the project](./LICENSE).