<a href="https://github.com/kerraform">
    <img src="https://avatars.githubusercontent.com/u/82173916?s=200&v=4" alt="Kerraform logo" title="Terraform" align="right" height="80" />
</a>

# Kegistry: Terraform Registry

![Test](https://github.com/kerraform/kegistry/workflows/CI/badge.svg)

> Terraform Registry fulfills Terraform provider and module API

[![Renovate Badge][Renovate Icon]][Renovate]
[![GoDoc Badge][GoDoc Icon]][GoDoc]
[![Docker Badge][Docker Icon]][Docker]

*Note: This is not production ready, it's under development*

## Supported features

These are the list of the supported features.

* Audit logs
* Storage
    * Local disk
    * S3 (or S3 compatible object storage)

## Configuration

Theses are environment variable list that you can configure.

| Variable  | Description | Type| Default | 
|:----:|:----:|:----:|:---:|
| `PORT`  | Port to listen | `int` | `8888` | 
| `BACKEND_TYPE` | Storage driver to use (supports `local` and `s3`) | `string` | (required) |
| `BACKEND_S3_BUCKET` | Bucket to store the blobs and the tags | `string` |  - (Required if `BACKEND_TYPE` is `s3`) |
| `LOG_FORMAT` | Format of the logs (supports `json`, `console`, `color`) | `json` | 
| `LOG_LEVEL` | Level of the logs (supports `info`, `debug`, `warn`, `error`) | `info` |

Note that you need to create a GCS bucket before running this server with `gcs` driver otherwise the server will fail to init.

## Author

* [KeisukeYamashita](https://github.com/KeisukeYamashita)

## License

* [Apache License 2.0](./LICENSE)

## References

* [Private Registries, Terraform](https://www.terraform.io/docs/registry/private.html)

<!-- Badge section -->
[Renovate Icon]: https://img.shields.io/badge/-Renovate-1A1F6C?style=flat-square&logo=renovatebot&logoColor=white
[Renovate]: https://www.whitesourcesoftware.com/free-developer-tools/renovate

[GoDoc Icon]: https://img.shields.io/badge/-Go-00ADD8?style=flat-square&logo=go&logoColor=white
[GoDoc]: xxx

[Docker Icon]: https://img.shields.io/badge/-Docker-2496ED?style=flat-square&logo=docker&logoColor=white
[Docker]: xxx
