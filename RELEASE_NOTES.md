# Cosmos SDK v0.44.5 Release Notes

This release adds the updates from cosmos-sdk upstream v0.44.4 and v0.44.5 to the msg fee and rosetta fixes added into the Provenanced fork at v0.45

# Cosmos SDK v0.45.0 Release Notes

- We now charge gas in two new places: on `.Seek()` even if there are no entries, and for the key length (on top of the value length).
- When block gas limit is exceeded, we consume the maximum gas possible (to charge for the performed computation). We also fixed the bug when the last transaction in a block exceeds the block gas limit, it returns an error result, but the tx is actually committed successfully.

Finally, a small improvement in gov, we increased the maximum proposal description size from 5k characters to 10k characters.

### API-Breaking Changes

- The `BankKeeper` interface has a new `HasSupply` method to ensure that input denom actually exists on chain.
- The `CommitMultiStore` interface contains a new `SetIAVLCacheSize` method for a configurable IAVL cache size.
- `AuthKeeper` interface in `x/auth` now includes a function `HasAccount`.
- Moved `TestMnemonic` from `testutil` package to `testdata`.


Finally, when using the `SetOrder*` functions in simapp, e.g. `SetOrderBeginBlocker`, we now require that all modules be present in the function arguments, or else the node panics at startup. We also added a new `SetOrderMigration` function to set the order of running module migrations.

### Improvements

- Speedup improvements (e.g. speedup iterator creation after delete heavy workloads, lower allocations for `Coins.String()`, reduce RAM/CPU usage inside store/cachekv's `Store.Write`) are included in this release.
- Upgrade Rosetta to v0.7.0 .
- Support in-place migration ordering.
- Copied and updated `server.GenerateCoinKey` and `server.GenerateServerCoinKey` functions to the `testutil` package. These functions in `server` package are marked deprecated and will be removed in the next release. In the `testutil.GenerateServerCoinKey` version we  added support for custom mnemonics in in-process testing network.

See our [CHANGELOG](./CHANGELOG.md) for the exhaustive list of all changes, or a full [commit diff](https://github.com/cosmos/cosmos-sdk/compare/v0.44.5...v0.45.0).

# Cosmos SDK v0.45.2 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.45 series:

Highlights:

- Add hooks to allow modules to add things to state-sync. Please see [PR #10961](https://github.com/cosmos/cosmos-sdk/pull/10961) for more information.
- Register [`EIP191`](https://eips.ethereum.org/EIPS/eip-191) as an available `SignMode` for chains to use. Please note that in v0.45.2, the Cosmos SDK does **not** support EIP-191 out of the box. But if your chain wants to integrate EIP-191, it's possible to do so by passing a `TxConfig` with your own sign mode handler which implements EIP-191, using the new provided `authtx.NewTxConfigWithHandler` function.
- Add a new `rollback` CLI command to perform a state rollback by one block. Read more in [PR #11179](https://github.com/cosmos/cosmos-sdk/pull/11179).
- Some new queries were added:
  - x/authz: `GrantsByGrantee` to query grants by grantee,
  - x/bank: `SpendableBalances` to query an account's total (paginated) spendable balances,
  - TxService: `GetBlockWithTxs` to fetch a block along with all its transactions, decoded.
- Some bug fixes, such as:
  - Update the prune `everything` strategy to store the last two heights.
  - Fix data race in store trace component.
  - Fix cgo secp signature verification and update libscep256k1 library.

See the [Cosmos SDK v0.45.2 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/CHANGELOG.md) for the exhaustive list of all changes and check other fixes in the 0.45.x release series.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.1...v0.45.2

# Cosmos SDK v0.45.3 Release Notes

This release introduces a Tendermint dependency update to v0.34.19 which
itself includes two bug fixes related to consensus. See the full changelog from
v0.34.17-v0.34.19 [here](https://github.com/tendermint/tendermint/blob/v0.34.19/CHANGELOG.md#v0.34.19).

In addition, it includes a change to `ScheduleUpgrade` to allow upgrades without
requiring a governance proposal process.

See the [Cosmos SDK v0.45.3 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.45.3/CHANGELOG.md)
for the exhaustive list of all changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.2...v0.45.3
