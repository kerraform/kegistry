<a href="https://github.com/kerraform">
    <img src="https://avatars.githubusercontent.com/u/82173916?s=200&v=4" alt="Kerraform logo" title="Terraform" align="right" height="80" />
</a>

# Kegistry: Terraform Registry

> Terraform Registry fulfills Terraform provider and module registry protocol

![Test](https://github.com/kerraform/kegistry/workflows/CI/badge.svg)
![Release](https://github.com/kerraform/kegistry/actions/workflows/release.yml)

[![Dependabot Badge][Dependabot Icon]][Dependabot]
[![GoDoc Badge][GoDoc Icon]][GoDoc]
[![Docker Badge][Docker Icon]][Docker]
[![Snyk Badge][Snyk Icon]][Snyk]
[![Fossa Badge][Fossa Icon]][Fossa]

*Note: This is not production ready, it's under development*

## Supported features

These are the list of the supported features.

* [Provider Registry Protocol](https://www.terraform.io/internals/provider-registry-protocol)
* Audit logs
* Storage
    * Local disk
    * Amazon S3 (or S3 compatible object storage)

## Configuration

Theses are environment variable list that you can configure.

| Variable  | Description | Type| Default | 
|:----|:----|:----|:---|
| `PORT`  | Port to listen | `int` | `8888` | 
| `BACKEND_TYPE` | Storage driver to use (supports `local` and `s3`) | `string` | (required) |
| `BACKEND_S3_ACCESS_KEY` | Access key of Amazon S3 | `string` |  - (Required if `BACKEND_TYPE` is `s3`) |
| `BACKEND_S3_BUCKET` | Amazon S3 Bucket name to store the resources | `string` |  - (Required if `BACKEND_TYPE` is `s3`) |
| `BACKEND_S3_SECRET_KEY` | Secret key of Amazon S3 | `string` |  - (Required if `BACKEND_TYPE` is `s3`) |
| `LOG_FORMAT` | Format of the logs (supports `json`, `console`, `color`) | `string` | `json` | 
| `LOG_LEVEL` | Level of the logs (supports `info`, `debug`, `warn`, `error`) | `string` | `info` |

Note that you need to create a GCS bucket before running this server with `gcs` driver otherwise the server will fail to init.

## Author

* [KeisukeYamashita](https://github.com/KeisukeYamashita)

## License

* [Apache License 2.0](./LICENSE)

## References

* [Private Registries, Terraform](https://www.terraform.io/docs/registry/private.html)

<!-- Badge section -->
[Dependabot Icon]: https://img.shields.io/badge/-Dependabot-025E8C?style=flat-square&logo=dependabot&logoColor=white
[Dependabot]: https://github.com/kerraform/kegistry/security/dependabot

[GoDoc Icon]: https://img.shields.io/badge/-Go-00ADD8?style=flat-square&logo=go&logoColor=white
[GoDoc]: xxx

[Docker Icon]: https://img.shields.io/badge/-Docker-2496ED?style=flat-square&logo=docker&logoColor=white
[Docker]: xxx

[Snyk Icon]: https://img.shields.io/badge/-Snyk-4C4A73?style=flat-square&logo=snyk&logoColor=white
[Snyk]: xxx

[Fossa Icon]: https://img.shields.io/badge/-Fossa-289E6D?style=flat-square&logo=fossa&logoColor=white
[Fossa]: xxx
