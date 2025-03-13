# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.2-alpha]

- Added `--debug/-D` flag for service

## [v0.2.1-alpha]

- Added `/healthz` healthcheck endpoint
- Added middleware logging:

```
{
    "level":"info",
    "method":"POST",
    "path":"/",
    "status":201,
    "message":{
        "notif_size": "299",
        "hl7_size": "10",
        "result": "ORM processed successfully",
        "elapsed": 357.931263
    }
}
```

- Added `gosec` in CI
- Changed hl7 package to use byte buffer instead of map for performance
- Modified release workflow

## [v0.1.0-alpha]

- Added CHANGELOG.md ([Keep a Changelog](https://keepachangelog.com/en/1.0.0/)).
- Added README.md, MIT license.
- Removed `/vendor` and updated `go.mod`.
- Added CI & Lint workflows.

[Unreleased]: https://github.com/s-hammon/volta/compare/v0.1.0-alpha...HEAD
[v0.1.0-alpha]: https://github.com/s-hammon/volta/releases/tag/v0.1.0-alpha
[v0.2.1-alpha]: https://github.com/s-hammon/volta/releases/tag/v0.2.1-alpha
[v0.2.2-alpha]: https://github.com/s-hammon/volta/releases/tag/v0.2.2-alpha
