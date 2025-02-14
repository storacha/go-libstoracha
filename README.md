# go-libstoracha
A unified Go library (monorepo) for Storacha functionality. This repository hosts multiple subpackages that address different aspects of Storacha’s ecosystem—from capability definitions to job queuing and more.

## Subpackages
- [capabilities](./capabilities)
    - **Description:** UCAN capability definitions for the Storacha ecosystem.
- [metadata](./metadata)
    - **Description:** IPNI metadata used by the Storacha Network.
- [jobqueue](./jobqueue)
    - **Description:** A reliable and parallelizable job queue.
- [ipnipublisher](./ipnipublisher)
    - **Description**: A library to create, sign, and publish adverts to a local IPNI chain, then announce them to other indexers.
    - **Usage**: See the [example_test.go](./ipnipublisher/example_test.go) file within this subpackage for a quick start.
- [piece](./piece)
    - **Description**: TODO

# Contributing
All are welcome! Storacha is open-source. Please feel empowered to open a pull request or file an issue if you have questions, suggestions, or bug reports.

# License
Dual-licensed under [Apache 2.0 OR MIT](./LICENSE.md). You may choose either license when using or contributing to this project.

