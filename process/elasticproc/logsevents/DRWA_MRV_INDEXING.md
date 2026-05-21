# DRWA/MRV Indexing Wiring

The live pipeline wires DRWA and MRV event materialization through the same logs-and-events path as standard ESDT/NFT data:

1. `logsAndEventsProcessor.ExtractDataFromLogs` receives block logs plus block hash, round, shard, and timestamp metadata.
2. Standard processors run first. DRWA and MRV processors run last and only accept events emitted by configured authorized smart-contract addresses.
3. Accepted records are accumulated in `data.PreparedLogsResults`:
   `DrwaDenials`, `DrwaIdentities`, `DrwaHolderCompliances`, `DrwaAttestations`, `DrwaTokenPolicies`, `DrwaControlEvents`, and `MrvAnchoredProofs`.
4. `elasticProcessor.prepareAndSaveTransactionsData` serializes those fields into the matching dedicated Elasticsearch indexes.
5. Reverts remove DRWA/MRV records by `blockHash` + `shardID`; finalized-block notifications mark the same records as finalized.

Empty `authorized-emitters` config fails closed. Generic logs/events are still indexed, but DRWA/MRV materialized records are not produced until operators configure the contract emitters.
