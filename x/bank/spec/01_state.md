<!--
order: 1
-->

# State

The `x/bank` module manages the following state:

1. The total supply of all balances:
   * `0x00 | byte(denom) -> byte(amount)`
2. Denomination metadata:
   * `0x01 | byte(denom) -> ProtocolBuffer(Metadata)`
3. Account balances:
   * `0x02 | byte(address length) | []byte(address) | []byte(balance.Denom) -> ProtocolBuffer(balance)`
4. Reverse Denomination to Address Index:
   * `0x03 | byte(denom) | 0x00 | []byte(address) -> 0`
5. Information on which denominations are allowed to be sent.
   * `0x04 | byte(denom) -> byte(boolean)`
   * `false = 0x00`
   * `true = 0x01`
6. Addresses that have enabled quarantine:
   * `0x20 | byte(address length) | []byte(address) -> 0`
7. Quarantine auto-responses:
   * `0x21 | byte(to address length) | byte(to address) | byte(from address length) | byte(from address) -> byte(QuarantineAutoResponse)`
   * `QUARANTINE_AUTO_RESPONSE_ACCEPT = 0x01`
   * `QUARANTINE_AUTO_RESPONSE_DECLINE = 0x02`
   * Note: `QUARANTINE_AUTO_RESPONSE_UNSPECIFIED` is not stored. Absence of an entry conveys that value.
8. Quarantined funds:
   * `0x22 | byte(to address length) | byte(to address) | byte(from address length) | byte(from address) -> ProtocolBuffer(QuarantineRecord)`

