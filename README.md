Release notes

Creates release notes based off a branch in k8s-engine

Also contains a helper script `./scripts/k8s-update` for updating images in the k8s-engine repo.
It allows you to select the env, namespace, and image to update.
To use the helper script symlink it into your path (i.e `${HOME}/.local/bin`)
then run `k8s-update` in the k8s-engine repo. 
It does have the prerequisites of `yq` and `gum`
- `brew install gum yq`

## Usage
- copy .env.sample to .env and fill in the email and jira token
- `go run main.go pr --private-key=/path/to/ssh/privatekey --path=path/to/k8s-engine --target=yourbranch`
- hopefully get some release notes in `release-notes.md`
