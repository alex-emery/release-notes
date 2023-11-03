Release notes

Creates release notes based off a branch in k8s-engine
## Usage
- copy .env.sample to .env and fill in the email and jira token
- `go run main.go pr --private-key=/path/to/ssh/privatekey --source=main --target=yourbranch`
- hopefully get some release notes in `release-notes.md`
