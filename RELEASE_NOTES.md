## [v0.46.10-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.10-pio-1) - 2023-02-16

### Improvements

* [#505](https://github.com/provenance-io/cosmos-sdk/pull/505) Revert [#444](https://github.com/provenance-io/cosmos-sdk/pull/505): Revert [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) "Optimize iteration on nested cached KV stores and other operations in general".
* [#505](https://github.com/provenance-io/cosmos-sdk/pull/505) [#517](https://github.com/provenance-io/cosmos-sdk/pull/517) Bring in Cosmos-SDK [v0.46.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.9) and [v0.46.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.10) changes.

### Features

* [#510](https://github.com/provenance-io/cosmos-sdk/pull/510) Add Sanction Tx commands.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.8-pio-3...v0.46.10-pio-1
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10..v0.46.10-pio-1

---

## [v0.46.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.10) - 2022-02-16

This release improves CPU profiling when using the `--cpu-profile` flag, and fixes a possible way to DoS a node.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.9...v0.46.10

### Improvements

* (cli) [#14953](https://github.com/cosmos/cosmos-sdk/pull/14953) Enable profiling block replay during abci handshake with `--cpu-profile`.

## [v0.46.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.9) - 2022-02-07

This release introduces bug fixes and improvements. Notably an extra config in the `app.toml`, `iavl-lazy-loading`, to enable lazy loading of IAVL store.
Changes to be made in the `app.toml` can be found in the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md).

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.8...v0.46.9

### Improvements

* (deps) [#14846](https://github.com/cosmos/cosmos-sdk/pull/14846) Bump btcd.
* (deps) Bump Tendermint version to [v0.34.26](https://github.com/informalsystems/tendermint/releases/tag/v0.34.26).
* (store) [#14189](https://github.com/cosmos/cosmos-sdk/pull/14189) Add config `iavl-lazy-loading` to enable lazy loading of iavl store, to improve start up time of archive nodes, add method `SetLazyLoading` to `CommitMultiStore` interface.
    * A new field has been added to the app.toml. This alllows nodes with larger databases to startup quicker

    ```toml
    # IAVLLazyLoading enable/disable the lazy loading of iavl store.
    # Default is false.
    iavl-lazy-loading = ""
  ```

### Bug Fixes

* (cli) [#14919](https://github.com/cosmos/cosmos-sdk/pull/#14919) Fix never assigned error when write validators.
* (store) [#14798](https://github.com/cosmos/cosmos-sdk/pull/14798) Copy btree to avoid the problem of modify while iteration.
* (cli) [#14799](https://github.com/cosmos/cosmos-sdk/pull/14799) Fix Evidence CLI query flag parsing (backport #13458)

