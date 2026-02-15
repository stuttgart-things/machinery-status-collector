# [0.1.0](https://github.com/stuttgart-things/machinery-status-collector/compare/v0.0.0...v0.1.0) (2026-02-15)


### Bug Fixes

* add GoReleaser to release workflow and reset versioning ([83fefae](https://github.com/stuttgart-things/machinery-status-collector/commit/83fefae9f0c0b4c22353183d035d257c67f6a1c7))
* change default server port from 8080 to 8095 ([7244eca](https://github.com/stuttgart-things/machinery-status-collector/commit/7244eca8a82b3767d2ff8ac36169630aff1ca5ec))
* replace disallowed semantic-release action with npx ([91a0099](https://github.com/stuttgart-things/machinery-status-collector/commit/91a00992adb8346e7f4f1d9761a25cd6bd2acb56))
* report all missing env vars at once on server start ([5167726](https://github.com/stuttgart-things/machinery-status-collector/commit/5167726dc21f8721919801953311dbe6be41ed32))
* run semantic-release only from release/next branch ([8ba4bfa](https://github.com/stuttgart-things/machinery-status-collector/commit/8ba4bfa9fb7578f9e579b68101a38abdab210c4d)), closes [#35](https://github.com/stuttgart-things/machinery-status-collector/issues/35)


### Features

* add build and release configuration ([b840d96](https://github.com/stuttgart-things/machinery-status-collector/commit/b840d9632a0bf7fe75d6a80252d743f141a98244)), closes [#15](https://github.com/stuttgart-things/machinery-status-collector/issues/15)
* add central API server with HTTP handlers and middleware ([08d51fb](https://github.com/stuttgart-things/machinery-status-collector/commit/08d51fb4e8130bd90b6cb4ef4a6c34d86e673db6))
* add GitHub REST API client for batch PR creation ([45e9564](https://github.com/stuttgart-things/machinery-status-collector/commit/45e9564a901e03ddb36fc7aa7db2a43a8e187b7f)), closes [#11](https://github.com/stuttgart-things/machinery-status-collector/issues/11) [#10](https://github.com/stuttgart-things/machinery-status-collector/issues/10)
* add informer Cobra subcommand for cluster agent ([c2299ab](https://github.com/stuttgart-things/machinery-status-collector/commit/c2299ab7280fd4d73f398cef9d7e1c5bc99f3a35)), closes [#14](https://github.com/stuttgart-things/machinery-status-collector/issues/14)
* add Kubernetes dynamic informer for Crossplane claim status ([d704c2a](https://github.com/stuttgart-things/machinery-status-collector/commit/d704c2a4714fa8a37982fe392fa6418bbc0be3de)), closes [#13](https://github.com/stuttgart-things/machinery-status-collector/issues/13)
* add OpenAPI spec and project documentation ([53f0e2c](https://github.com/stuttgart-things/machinery-status-collector/commit/53f0e2ccae33411f3a1e781704abbbdd63c53ecb)), closes [#16](https://github.com/stuttgart-things/machinery-status-collector/issues/16)
* add reconciler for periodic batch PR creation ([aca435a](https://github.com/stuttgart-things/machinery-status-collector/commit/aca435a5c7510723d5d729497c2767f04d25d3ec)), closes [#11](https://github.com/stuttgart-things/machinery-status-collector/issues/11)
* add registry types and YAML parsing for claim entries ([603f6ef](https://github.com/stuttgart-things/machinery-status-collector/commit/603f6efc3a4f9f6afbab7881247ea115b12c9526)), closes [#7](https://github.com/stuttgart-things/machinery-status-collector/issues/7)
* add server Cobra subcommand ([4b93a19](https://github.com/stuttgart-things/machinery-status-collector/commit/4b93a19113e5f26a11895a41b58a1fbdcfae9d32)), closes [#12](https://github.com/stuttgart-things/machinery-status-collector/issues/12)
* add thread-safe in-memory status store for cluster agents ([1bfddd1](https://github.com/stuttgart-things/machinery-status-collector/commit/1bfddd162090c0771799b874aa25b78e898fbbde)), closes [#8](https://github.com/stuttgart-things/machinery-status-collector/issues/8)
* Cobra CLI scaffold with version and logo commands ([5c42595](https://github.com/stuttgart-things/machinery-status-collector/commit/5c425955d6b896d92afd207c6da615c5680f38c7)), closes [#6](https://github.com/stuttgart-things/machinery-status-collector/issues/6)
* docs/server-usage ([204c29e](https://github.com/stuttgart-things/machinery-status-collector/commit/204c29e089a3b6b0fd630d6c6663a389627aa0a2))
* update ASCII logo to MACHINERY STATUS COLLECTOR ([7a0e943](https://github.com/stuttgart-things/machinery-status-collector/commit/7a0e943798abdaa508856aa6bea348924924480d)), closes [#33](https://github.com/stuttgart-things/machinery-status-collector/issues/33)

# Changelog

All notable changes to this project will be documented in this file.

This project uses [semantic-release](https://github.com/semantic-release/semantic-release) to manage releases.

## [Unreleased]
- Initial scaffold created via Backstage template
