# tfreg (Terraform Registry)
`tfreg` is a simple implementation of Terraform [registry API](https://www.terraform.io/docs/internals/module-registry-protocol.html) which provides modules directly from [Gitlab](https://gitlab.com). You could load private modules directly from version control with the [generic Git repository syntax](https://www.terraform.io/docs/language/modules/sources.html#generic-git-repository) or from other sources, but those sources don't support [version constraints](https://www.terraform.io/docs/language/modules/syntax.html#version).

## Versions
`tfreg` makes available all tags with a `v` prefix (e.g. `v1.1.4`). So to publish a new module version you just have to create a new tag with a `v` prefix (e.g. `git tag -a v1.1.4 -m ""`).

## Repositories
If you specify a Terraform Registry as a module source the source has the following format:
```
module "mytest" {
  source = "HOSTNAME/NAMESPACE/NAME/PROVIDER"
  version = "< 2.0.0"
}
```

With the `-template` option you can specify how the namespace, name and provider are mapped to a gitlab project ID. By default the format is `{{ .Namespace }}/{{ .Name }}` and the provider is not used. This would for example map `example.com/dvob/terraform-dummy-module/base` to the gitlab project https://gitlab.com/dvob/terraform-dummy-module. The provider in this case would be `base` but by default it is not used in the Gitlab project ID.

# Local Test
* Create a DNS record `example.com` which points to `127.0.0.1`
* Create a TLS key pair (`tls.key` and `tls.crt`) for `example.com` (for example with [pcert](https://github.com/dvob/pcert) or OpenSSL)
* Build `go build`
* Run `./tfreg`
* Trust your self signed certificate `export SSL_CERT_FILE=$(pwd)/tls.crt`
* Go to the test dirctory `cd test/` and run `terraform init` and `terraform apply`


If you want to host modules from private Gitlab repositories [create a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token) and set the environment variable `GITLAB_TOKEN` before you start `tfreg`.  Also make sure that you have configured the Git credential helper correctly before you run `terraform init`. For example:
```
git config --global credential.helper '!f() { sleep 1; echo "username=none"; echo "password=${GITLAB_TOKEN}"; }; f'
```
