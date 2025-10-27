## aws-mfa-sso
Runs Aws SSO login with a no-browser option and uses a headless implementation to do the workflow.

[![Go Reference](https://pkg.go.dev/badge/github.com/rhyas/aws-mfa-sso.svg)](https://pkg.go.dev/github.com/rhyas/aws-mfa-sso) [![Go Report Card](https://goreportcard.com/badge/github.com/rhyas/aws-mfa-sso)](https://goreportcard.com/report/github.com/rhyas/aws-mfa-sso) 

### Background

Opening a browser for a CLI tool is silly.

### Install

```bash
go install github.com/rhyas/aws-mfa-sso@latest
```

### Usage:

``` bash
./aws-mfa-sso --profile <my-profile> (Or AWS_PROFILE env)
```

### Credential Helper
This can also be used as a credential helper for multiple accounts. Simply setup your main sso account in .aws/config like this:
```
[profile sso]
sso_start_url = https://<identifier>.awsapps.com/start
sso_region = <region>
sso_account_id = <aws-account>
sso_role_name = <role-name>
```
Then setup the downstream account profile:
```
[profile my-account]
credential_process = /path/to/aws-mfa-sso credential-process -r "arn:aws:iam::<my-account-id>:role/<role-to-assume>"
source_profile = sso
region = <region>
```

### Limitations:
- Requires a Virtual MFA token to generate token responses.

### License:
MIT