# Changelog

# Provenance Specific Releases

## [v0.45.5-pio-2](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.45.5-pio-2) - 2022-08-10

# Features

* Add support for event data injection into end block handlers (https://github.com/provenance-io/provenance/issues/626)
  Currently the Cosmos SDK End block methods are only provided with the block height as a parameter, this feature would inject the events from block processing to support more advanced reactive features.

### Full Commit History

https://github.com/provenance-io/cosmos-sdk/compare/v0.45.5-pio-1...v0.45.5-pio-2