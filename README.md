Release notes

Creates release notes based off a branch in k8s-engine

Also contains a helper script `./scripts/k8s-update` for updating images in the k8s-engine repo.
It allows you to select the env, namespace, and image to update.
To use the helper script symlink it into your path (i.e `${HOME}/.local/bin`)
then run `k8s-update` in the k8s-engine repo. 
It does have the prerequisites of `yq` and `gum`
- `brew install gum yq`

## Prerequisites
Generate Github token
- https://github.com/settings/tokens
    - create a classic token
    - select full repo access
    - create token
    - authenticate org via SSO

Generate Jira Tokens 
- https://id.atlassian.com/manage-profile/security/api-tokens

## Usage
- `go install github.com/alex-emery/release-notes@v0.0.6`
- `export JIRA_TOKEN=<jira token>`
- `export JIRA_EMAIL=<jira email>`
- `export GITHUB_TOKEN=<GITHUB_TOKEN># used to create the PR`
- `release-notes pr --path=path/to/k8s-engine --target=yourbranch`
- A PR will automatically be created in `k8s-engine`


