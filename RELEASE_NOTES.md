# Changelog

# Provenance Specific Releases

## [v0.45.5-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.45.5-pio-1) - 2022-08-01

### API Breaking Changes

* [#215](https://github.com/provenance-io/cosmos-sdk/pull/215) Remove ADR-038 plugin system due to `AppHash` error.
  The current [ADR 038: State Listening](https://github.com/provenance-io/cosmos-sdk/blob/egaxhaj-figure/adr-038-plugin-system/docs/architecture/adr-038-state-listening.md) implementation leads to an `AppHash` mismatch error that causes the node to crash when State Listening is enabled.

### Full Commit History

https://github.com/provenance-io/cosmos-sdk/compare/v0.45.4-pio-4...v0.45.5-pio-1