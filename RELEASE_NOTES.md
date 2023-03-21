## [v0.46.10-pio-3](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.10-pio-3) - 2023-03-21

### Features

* [PR 563](https://github.com/provenance-io/cosmos-sdk/pull/563) Add support for applying send restrictions before doing a bank send

### Improvements

* [PR 565](https://github.com/provenance-io/cosmos-sdk/pull/565) Removes locks around state listening. There's no concurrency risk.

### Bug Fixes

* [PR 564](https://github.com/provenance-io/cosmos-sdk/pull/564) Fix protobufjs parse error by using object form vs. array for `additional_bindings` rpc tag.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10-pio-2...v0.46.10-pio-3
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10..v0.46.10-pio-3
