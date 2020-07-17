package docs

const Swagger = `
basePath: /
definitions:
  Address:
    description: bech32 encoded address
    example: cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27
    type: string
  BaseReq:
    properties:
      account_number:
        example: "0"
        type: string
      chain_id:
        example: Cosmos-Hub
        type: string
      fees:
        items:
          $ref: '#/definitions/Coin'
        type: array
      from:
        description: Sender address or Keybase name to generate a transaction
        example: cosmos1g9ahr6xhht5rmqven628nklxluzyv8z9jqjcmc
        type: string
      gas:
        example: "200000"
        type: string
      gas_adjustment:
        example: "1.2"
        type: string
      memo:
        example: "Sent via Cosmos Voyager \U0001F680"
        type: string
      sequence:
        example: "1"
        type: string
      simulate:
        description: Estimate gas for a transaction (cannot be used in conjunction with generate_only)
        example: false
        type: boolean
    type: object
  Block:
    properties:
      evidence:
        items:
          type: string
        type: array
      header:
        $ref: '#/definitions/BlockHeader'
      last_commit:
        properties:
          block_id:
            $ref: '#/definitions/BlockID'
          precommits:
            items:
              properties:
                block_id:
                  $ref: '#/definitions/BlockID'
                height:
                  example: "0"
                  type: string
                round:
                  example: "0"
                  type: string
                signature:
                  example: 7uTC74QlknqYWEwg7Vn6M8Om7FuZ0EO4bjvuj6rwH1mTUJrRuMMZvAAqT9VjNgP0RA/TDp6u/92AqrZfXJSpBQ==
                  type: string
                timestamp:
                  example: "2017-12-30T05:53:09.287+01:00"
                  type: string
                type:
                  example: 2
                  type: number
                validator_address:
                  type: string
                validator_index:
                  example: "0"
                  type: string
              type: object
            type: array
        type: object
      txs:
        items:
          type: string
        type: array
    type: object
  BlockHeader:
    properties:
      app_hash:
        $ref: '#/definitions/Hash'
      chain_id:
        example: cosmoshub-2
        type: string
      consensus_hash:
        $ref: '#/definitions/Hash'
      data_hash:
        $ref: '#/definitions/Hash'
      evidence_hash:
        $ref: '#/definitions/Hash'
      height:
        example: 1
        type: number
      last_block_id:
        $ref: '#/definitions/BlockID'
      last_commit_hash:
        $ref: '#/definitions/Hash'
      last_results_hash:
        $ref: '#/definitions/Hash'
      next_validators_hash:
        $ref: '#/definitions/Hash'
      num_txs:
        example: 0
        type: number
      proposer_address:
        $ref: '#/definitions/Address'
      time:
        example: "2017-12-30T05:53:09.287+01:00"
        type: string
      total_txs:
        example: 35
        type: number
      validators_hash:
        $ref: '#/definitions/Hash'
      version:
        properties:
          app:
            example: 0
            type: string
          block:
            example: 10
            type: string
        type: object
    type: object
  BlockID:
    properties:
      hash:
        $ref: '#/definitions/Hash'
      parts:
        properties:
          hash:
            $ref: '#/definitions/Hash'
          total:
            example: 0
            type: number
        type: object
    type: object
  BlockQuery:
    properties:
      block:
        $ref: '#/definitions/Block'
      block_meta:
        properties:
          block_id:
            $ref: '#/definitions/BlockID'
          header:
            $ref: '#/definitions/BlockHeader'
        type: object
    type: object
  BroadcastTxCommitResult:
    properties:
      check_tx:
        $ref: '#/definitions/CheckTxResult'
      deliver_tx:
        $ref: '#/definitions/DeliverTxResult'
      hash:
        $ref: '#/definitions/Hash'
      height:
        type: integer
    type: object
  CheckTxResult:
    example:
      code: 0
      data: data
      gas_used: 5000
      gas_wanted: 10000
      info: info
      log: log
      tags:
      - ""
      - ""
    properties:
      code:
        type: integer
      data:
        type: string
      gas_used:
        type: integer
      gas_wanted:
        type: integer
      info:
        type: string
      log:
        type: string
      tags:
        items:
          $ref: '#/definitions/KVPair'
        type: array
    type: object
  Coin:
    properties:
      amount:
        example: "50"
        type: string
      denom:
        example: stake
        type: string
    type: object
  Delegation:
    properties:
      balance:
        $ref: '#/definitions/Coin'
      delegator_address:
        type: string
      shares:
        type: string
      validator_address:
        type: string
    type: object
  DelegationDelegatorReward:
    properties:
      reward:
        items:
          $ref: '#/definitions/Coin'
        type: array
      validator_address:
        $ref: '#/definitions/ValidatorAddress'
    type: object
  DelegatorTotalRewards:
    properties:
      rewards:
        items:
          $ref: '#/definitions/DelegationDelegatorReward'
        type: array
      total:
        items:
          $ref: '#/definitions/Coin'
        type: array
    type: object
  DeliverTxResult:
    example:
      code: 5
      data: data
      gas_used: 5000
      gas_wanted: 10000
      info: info
      log: log
      tags:
      - ""
      - ""
    properties:
      code:
        type: integer
      data:
        type: string
      gas_used:
        type: integer
      gas_wanted:
        type: integer
      info:
        type: string
      log:
        type: string
      tags:
        items:
          $ref: '#/definitions/KVPair'
        type: array
    type: object
  Deposit:
    properties:
      amount:
        items:
          $ref: '#/definitions/Coin'
        type: array
      depositor:
        $ref: '#/definitions/Address'
      proposal_id:
        type: string
    type: object
  Hash:
    example: EE5F3404034C524501629B56E0DDC38FAD651F04
    type: string
  KVPair:
    properties:
      key:
        type: string
      value:
        type: string
    type: object
  Msg:
    type: string
  PaginatedQueryTxs:
    properties:
      count:
        example: 1
        type: number
      limit:
        example: 30
        type: number
      page_number:
        example: 1
        type: number
      page_total:
        example: 1
        type: number
      total_count:
        example: 1
        type: number
      txs:
        items:
          $ref: '#/definitions/TxQuery'
        type: array
    type: object
  ParamChange:
    properties:
      key:
        example: MaxValidators
        type: string
      subkey:
        example: ""
        type: string
      subspace:
        example: staking
        type: string
      value:
        type: object
    type: object
  Proposer:
    properties:
      proposal_id:
        type: string
      proposer:
        type: string
    type: object
  PublicKey:
    properties:
      type:
        type: string
      value:
        type: string
    type: object
  Redelegation:
    properties:
      delegator_address:
        type: string
      entries:
        items:
          $ref: '#/definitions/Redelegation'
        type: array
      validator_dst_address:
        type: string
      validator_src_address:
        type: string
    type: object
  RedelegationEntry:
    properties:
      balance:
        type: string
      completion_time:
        type: integer
      creation_height:
        type: integer
      initial_balance:
        type: string
      shares_dst:
        type: string
    type: object
  SigningInfo:
    properties:
      index_offset:
        type: string
      jailed_until:
        type: string
      missed_blocks_counter:
        type: string
      start_height:
        type: string
    type: object
  StdTx:
    properties:
      fee:
        properties:
          amount:
            items:
              $ref: '#/definitions/Coin'
            type: array
          gas:
            type: string
        type: object
      memo:
        type: string
      msg:
        items:
          $ref: '#/definitions/Msg'
        type: array
      signature:
        properties:
          account_number:
            example: "0"
            type: string
          pub_key:
            properties:
              type:
                example: tendermint/PubKeySecp256k1
                type: string
              value:
                example: Avz04VhtKJh8ACCVzlI8aTosGy0ikFXKIVHQ3jKMrosH
                type: string
            type: object
          sequence:
            example: "0"
            type: string
          signature:
            example: MEUCIQD02fsDPra8MtbRsyB1w7bqTM55Wu138zQbFcWx4+CFyAIge5WNPfKIuvzBZ69MyqHsqD8S1IwiEp+iUb6VSdtlpgY=
            type: string
        type: object
    type: object
  Supply:
    properties:
      total:
        items:
          $ref: '#/definitions/Coin'
        type: array
    type: object
  TallyResult:
    properties:
      abstain:
        example: "0.0000000000"
        type: string
      "false":
        example: "0.0000000000"
        type: string
      no_with_veto:
        example: "0.0000000000"
        type: string
      "true":
        example: "0.0000000000"
        type: string
    type: object
  TendermintValidator:
    properties:
      address:
        $ref: '#/definitions/ValidatorAddress'
      proposer_priority:
        example: "1000"
        type: string
      pub_key:
        example: cosmosvalconspub1zcjduepq0vu2zgkgk49efa0nqwzndanq5m4c7pa3u4apz4g2r9gspqg6g9cs3k9cuf
        type: string
      voting_power:
        example: "1000"
        type: string
    type: object
  TextProposal:
    properties:
      description:
        type: string
      final_tally_result:
        $ref: '#/definitions/TallyResult'
      proposal_id:
        type: integer
      proposal_status:
        type: string
      proposal_type:
        type: string
      submit_time:
        type: string
      title:
        type: string
      total_deposit:
        items:
          $ref: '#/definitions/Coin'
        type: array
      voting_start_time:
        type: string
    type: object
  TxQuery:
    properties:
      hash:
        example: D085138D913993919295FF4B0A9107F1F2CDE0D37A87CE0644E217CBF3B49656
        type: string
      height:
        example: 368
        type: number
      result:
        properties:
          gas_used:
            example: "26354"
            type: string
          gas_wanted:
            example: "200000"
            type: string
          log:
            type: string
          tags:
            items:
              $ref: '#/definitions/KVPair'
            type: array
        type: object
      tx:
        $ref: '#/definitions/StdTx'
    type: object
  UnbondingDelegation:
    properties:
      balance:
        type: string
      creation_height:
        type: integer
      delegator_address:
        type: string
      initial_balance:
        type: string
      min_time:
        type: integer
      validator_address:
        type: string
    type: object
  UnbondingDelegationPair:
    properties:
      delegator_address:
        type: string
      entries:
        items:
          $ref: '#/definitions/UnbondingEntries'
        type: array
      validator_address:
        type: string
    type: object
  UnbondingEntries:
    properties:
      balance:
        type: string
      creation_height:
        type: string
      initial_balance:
        type: string
      min_time:
        type: string
    type: object
  Validator:
    properties:
      bond_height:
        example: "0"
        type: string
      bond_intra_tx_counter:
        example: 0
        type: integer
      commission:
        properties:
          max_change_rate:
            example: "0"
            type: string
          max_rate:
            example: "0"
            type: string
          rate:
            example: "0"
            type: string
          update_time:
            example: "1970-01-01T00:00:00Z"
            type: string
        type: object
      consensus_pubkey:
        example: cosmosvalconspub1zcjduepq0vu2zgkgk49efa0nqwzndanq5m4c7pa3u4apz4g2r9gspqg6g9cs3k9cuf
        type: string
      delegator_shares:
        type: string
      description:
        properties:
          details:
            type: string
          identity:
            type: string
          moniker:
            type: string
          security_contact:
            type: string
          website:
            type: string
        type: object
      jailed:
        type: boolean
      operator_address:
        $ref: '#/definitions/ValidatorAddress'
      status:
        type: integer
      tokens:
        type: string
      unbonding_height:
        example: "0"
        type: string
      unbonding_time:
        example: "1970-01-01T00:00:00Z"
        type: string
    type: object
  ValidatorAddress:
    description: bech32 encoded address
    example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
    type: string
  ValidatorDistInfo:
    properties:
      operator_address:
        $ref: '#/definitions/ValidatorAddress'
      self_bond_rewards:
        items:
          $ref: '#/definitions/Coin'
        type: array
      val_commission:
        items:
          $ref: '#/definitions/Coin'
        type: array
    type: object
  Vote:
    properties:
      option:
        type: string
      proposal_id:
        type: string
      voter:
        type: string
    type: object
  ccstorage.Currency:
    $ref: '#/definitions/types.Currency'
  client.MoveFile:
    properties:
      code:
        type: string
    type: object
  markets.MarketExtended:
    $ref: '#/definitions/types.MarketExtended'
  rest.BaseReq:
    properties:
      account_number:
        type: integer
      chain_id:
        type: string
      fees:
        $ref: '#/definitions/types.Coins'
        type: object
      from:
        type: string
      gas:
        type: string
      gas_adjustment:
        type: string
      gas_prices:
        $ref: '#/definitions/types.DecCoins'
        type: object
      memo:
        type: string
      sequence:
        type: integer
      simulate:
        type: boolean
    type: object
  rest.CCRespGetCurrency:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/ccstorage.Currency'
        type: object
    type: object
  rest.CCRespGetIssue:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Issue'
        type: object
    type: object
  rest.CCRespGetWithdraw:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Withdraw'
        type: object
    type: object
  rest.CCRespGetWithdraws:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Withdraws'
        type: object
    type: object
  rest.ErrorResponse:
    properties:
      code:
        type: integer
      error:
        type: string
    type: object
  rest.MSRespGetCall:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.CallResp'
        type: object
    type: object
  rest.MSRespGetCalls:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.CallsResp'
        type: object
    type: object
  rest.MarketsRespGetMarket:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Market'
        type: object
    type: object
  rest.MarketsRespGetMarkets:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Markets'
        type: object
    type: object
  rest.OracleRespGetAssets:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Assets'
        type: object
    type: object
  rest.OracleRespGetPrice:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.CurrentPrice'
        type: object
    type: object
  rest.OracleRespGetRawPrices:
    properties:
      height:
        type: integer
      result:
        items:
          $ref: '#/definitions/types.PostedPrice'
        type: array
    type: object
  rest.OrdersRespGetOrder:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Order'
        type: object
    type: object
  rest.OrdersRespGetOrders:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.Orders'
        type: object
    type: object
  rest.OrdersRespPostOrder:
    properties:
      type:
        type: string
      value:
        properties:
          fee:
            $ref: '#/definitions/types.StdFee'
            type: object
          memo:
            type: string
          msg:
            $ref: '#/definitions/rest.PostOrderMsg'
            type: object
          signatures:
            items:
              $ref: '#/definitions/types.StdSignature'
            type: array
        type: object
    type: object
  rest.OrdersRespRevokeOrder:
    properties:
      type:
        type: string
      value:
        properties:
          fee:
            $ref: '#/definitions/types.StdFee'
            type: object
          memo:
            type: string
          msg:
            $ref: '#/definitions/rest.RevokeOrderMsg'
            type: object
          signatures:
            items:
              $ref: '#/definitions/types.StdSignature'
            type: array
        type: object
    type: object
  rest.PoaRespGetValidators:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.ValidatorsConfirmationsResp'
        type: object
    type: object
  rest.PostOrderMsg:
    properties:
      type:
        type: string
      value:
        $ref: '#/definitions/types.MsgPostOrder'
        type: object
    type: object
  rest.PostOrderReq:
    properties:
      asset_code:
        description: 'Market assetCode in the following format: {base_denomination_symbol}_{quote_denomination_symbol}'
        example: btc_dfi
        type: string
      base_req:
        $ref: '#/definitions/rest.BaseReq'
        type: object
      direction:
        description: Order type (ask/bid)
        example: ask
        type: string
      price:
        description: QuoteAsset price with decimals (1.0 DFI with 18 decimals -> 1000000000000000000)
        example: "100"
        type: string
      quantity:
        description: BaseAsset quantity with decimals (1.0 BTC with 8 decimals -> 100000000)
        example: "10"
        type: string
      ttl_in_sec:
        description: Order TTL [s]
        example: "3"
        type: string
    type: object
  rest.RevokeOrderMsg:
    properties:
      type:
        type: string
      value:
        $ref: '#/definitions/types.MsgRevokeOrder'
        type: object
    type: object
  rest.RevokeOrderReq:
    properties:
      base_req:
        $ref: '#/definitions/rest.BaseReq'
        type: object
      order_id:
        example: "100"
        type: string
    type: object
  rest.VmData:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.QueryValueResp'
        type: object
    type: object
  rest.VmRespCompile:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/client.MoveFile'
        type: object
    type: object
  rest.VmTxStatus:
    properties:
      height:
        type: integer
      result:
        $ref: '#/definitions/types.TxVMStatus'
        type: object
    type: object
  rest.compileReq:
    properties:
      address:
        description: Code address
        example: wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h
        type: string
      code:
        description: Script code
        type: string
    type: object
  rest.postPriceReq:
    properties:
      asset_code:
        description: AssetCode
        example: dfi
        type: string
      base_req:
        $ref: '#/definitions/rest.BaseReq'
        type: object
      price:
        description: BigInt
        example: "100"
        type: string
      received_at:
        description: Timestamp Price createdAt
        example: "2020-03-27T13:45:15.293426Z"
        format: RFC 3339
        type: string
    type: object
  types.AccAddress:
    items:
      type: integer
    type: array
  types.Asset:
    properties:
      active:
        description: Not used ATM
        type: boolean
      asset_code:
        description: Asset code
        example: btc_dfi
        type: string
      oracles:
        $ref: '#/definitions/types.Oracles'
        description: List of registered RawPrice sources
        type: object
    type: object
  types.Assets:
    items:
      $ref: '#/definitions/types.Asset'
    type: array
  types.CallsResp:
    items:
      $ref: '#/definitions/types.CallResp'
    type: array
  types.Coin:
    properties:
      amount:
        $ref: '#/definitions/types.Int'
        description: |-
          To allow the use of unsigned integers (see: #1273) a larger refactor will
          need to be made. So we use signed integers for now with safety measures in
          place preventing negative values being used.
        type: object
      denom:
        type: string
    type: object
  types.Coins:
    items:
      $ref: '#/definitions/types.Coin'
    type: array
  types.Currency:
    properties:
      decimals:
        description: Number of currency decimals
        example: 0
        type: integer
      denom:
        description: Currency denom (symbol)
        example: dfi
        type: string
      supply:
        description: Total amount of currency coins in Bank
        example: "100"
        type: string
    type: object
  types.CurrentPrice:
    properties:
      asset_code:
        description: Asset code
        example: dfi
        type: string
      price:
        description: Price
        example: "1000"
        type: string
      received_at:
        description: UNIX Timestamp price createdAt [sec]
        example: "2020-03-27T13:45:15.293426Z"
        format: RFC 3339
        type: string
    type: object
  types.Dec:
    type: object
  types.DecCoin:
    properties:
      amount:
        $ref: '#/definitions/types.Dec'
        type: object
      denom:
        type: string
    type: object
  types.DecCoins:
    items:
      $ref: '#/definitions/types.DecCoin'
    type: array
  types.ID:
    $ref: '#/definitions/sdk.Uint'
  types.Int:
    type: object
  types.Issue:
    properties:
      coin:
        description: Issuing coin
        example: 100dfi
        type: string
      payee:
        description: Target account for increasing coin balance
        example: wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h
        format: bech32
        type: string
    type: object
  types.Market:
    properties:
      base_asset_denom:
        description: Base asset denomination (for ex. btc)
        example: btc
        type: string
      id:
        description: Market unique ID
        example: "0"
        type: string
      quote_asset_denom:
        description: Quote asset denomination (for ex. dfi)
        example: dfi
        type: string
    type: object
  types.MarketExtended:
    properties:
      baseCurrency:
        $ref: '#/definitions/ccstorage.Currency'
        description: Base asset currency (for ex. btc)
        type: object
      id:
        description: Market unique ID
        example: "0"
        type: string
      quoteCurrency:
        $ref: '#/definitions/ccstorage.Currency'
        description: Quote asset currency (for ex. dfi)
        type: object
    type: object
  types.Markets:
    items:
      $ref: '#/definitions/types.Market'
    type: array
  types.MsgPostOrder:
    properties:
      asset_code:
        type: string
      direction:
        type: string
      owner:
        $ref: '#/definitions/types.AccAddress'
        type: object
      price:
        $ref: '#/definitions/types.Uint'
        type: object
      quantity:
        $ref: '#/definitions/types.Uint'
        type: object
      ttl_in_sec:
        type: integer
    type: object
  types.MsgRevokeOrder:
    properties:
      order_id:
        $ref: '#/definitions/types.ID'
        type: object
      owner:
        $ref: '#/definitions/types.AccAddress'
        type: object
    type: object
  types.Oracle:
    properties:
      address:
        $ref: '#/definitions/types.AccAddress'
        description: Address
        type: object
    type: object
  types.Oracles:
    items:
      $ref: '#/definitions/types.Oracle'
    type: array
  types.Order:
    properties:
      created_at:
        description: Created timestamp
        example: "2020-03-27T13:45:15.293426Z"
        format: RFC 3339
        type: string
      direction:
        description: Order type (bid/ask)
        example: bid
        type: string
      id:
        description: Order unique ID
        example: "0"
        type: string
      market:
        $ref: '#/definitions/markets.MarketExtended'
        description: Market order belong to
        type: object
      owner:
        description: Order owner account address
        example: wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h
        format: bech32
        type: string
      price:
        description: Order target price (in quote asset denom)
        example: "100"
        type: string
      quantity:
        description: Order target quantity
        example: "50"
        type: string
      ttl_dur:
        description: TimeToLive order auto-cancel period
        example: 60
        type: integer
      updated_at:
        description: Updated timestamp
        example: "2020-03-27T13:45:15.293426Z"
        format: RFC 3339
        type: string
    type: object
  types.Orders:
    items:
      $ref: '#/definitions/types.Order'
    type: array
  types.PostedPrice:
    properties:
      asset_code:
        description: Asset code
        example: dfi
        type: string
      oracle_address:
        description: Source oracle address
        example: wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h
        type: string
      price:
        description: Price
        example: "1000"
        type: string
      received_at:
        description: UNIX Timestamp price receivedAt [sec]
        example: "2020-03-27T13:45:15.293426Z"
        format: RFC 3339
        type: string
    type: object
  types.QueryValueResp:
    properties:
      value:
        format: HEX string
        type: string
    type: object
  types.StdFee:
    properties:
      amount:
        $ref: '#/definitions/types.Coins'
        type: object
      gas:
        type: integer
    type: object
  types.StdSignature:
    properties:
      signature:
        items:
          type: integer
        type: array
    type: object
  types.TxVMStatus:
    properties:
      hash:
        type: string
      vm_status:
        $ref: '#/definitions/types.VMStatuses'
        type: object
    type: object
  types.Uint:
    type: object
  types.VMStatus:
    properties:
      major_code:
        description: Major code.
        type: string
      message:
        description: Message.
        type: string
      status:
        description: 'Status of error: error/discard.'
        type: string
      str_code:
        description: Detailed exaplantion of code.
        type: string
      sub_code:
        description: Sub code.
        type: string
    type: object
  types.VMStatuses:
    items:
      $ref: '#/definitions/types.VMStatus'
    type: array
  types.Validators:
    items:
      $ref: '#/definitions/types.Validator'
    type: array
  types.ValidatorsConfirmationsResp:
    properties:
      confirmations:
        description: Minimum number of confirmations needed to approve Call
        example: 3
        type: integer
      validators:
        $ref: '#/definitions/types.Validators'
        description: Registered validators list
        type: object
    type: object
  types.Withdraw:
    properties:
      coin:
        description: Target currency Coin
        example: 100dfi
        type: string
      id:
        description: Withdraw unique ID
        example: "0"
        type: string
      pegzone_chain_id:
        description: 'Second blockchain: ID'
        example: testnet
        type: string
      pegzone_spender:
        description: 'Second blockchain: spender account'
        example: wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h
        format: bech32
        type: string
      spender:
        description: Target account for reducing coin balance
        example: wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h
        format: bech32
        type: string
      timestamp:
        description: Tx UNIX time [s]
        example: 1585295757
        format: seconds
        type: integer
      tx_hash:
        description: Tx hash
        example: fd82ce32835dfd7042808eaf6ff09cece952b9da20460fa462420a93607fa96f
        type: string
    type: object
  types.Withdraws:
    items:
      $ref: '#/definitions/types.Withdraw'
    type: array
host: stargate.cosmos.network
info:
  contact: {}
  description: A REST interface for state queries, transaction generation and broadcasting.
  license: {}
  title: Gaia-Lite for Cosmos
  version: "3.0"
paths:
  /auth/accounts/{address}:
    get:
      parameters:
      - description: Account address
        in: path
        name: address
        required: true
        type: string
        x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
      produces:
      - application/json
      responses:
        "200":
          description: Account information on the blockchain
          schema:
            properties:
              type:
                type: string
              value:
                properties:
                  account_number:
                    type: string
                  address:
                    type: string
                  coins:
                    items:
                      $ref: '#/definitions/Coin'
                    type: array
                  public_key:
                    $ref: '#/definitions/PublicKey'
                  sequence:
                    type: string
                type: object
            type: object
        "500":
          description: Server internel error
      summary: Get the account information on blockchain
      tags:
      - Auth
  /bank/accounts/{address}/transfers:
    post:
      consumes:
      - application/json
      parameters:
      - description: Account address in bech32 format
        in: path
        name: address
        required: true
        type: string
        x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
      - description: The sender and tx information
        in: body
        name: account
        required: true
        schema:
          properties:
            amount:
              items:
                $ref: '#/definitions/Coin'
              type: array
            base_req:
              $ref: '#/definitions/BaseReq'
          type: object
      produces:
      - application/json
      responses:
        "202":
          description: Tx was succesfully generated
          schema:
            $ref: '#/definitions/StdTx'
        "400":
          description: Invalid request
        "500":
          description: Server internal error
      summary: Send coins from one account to another
      tags:
      - Bank
  /bank/balances/{address}:
    get:
      parameters:
      - description: Account address in bech32 format
        in: path
        name: address
        required: true
        type: string
        x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
      produces:
      - application/json
      responses:
        "200":
          description: Account balances
          schema:
            items:
              $ref: '#/definitions/Coin'
            type: array
        "500":
          description: Server internal error
      summary: Get the account balances
      tags:
      - Bank
  /blocks/{height}:
    get:
      parameters:
      - description: Block height
        in: path
        name: height
        required: true
        type: number
        x-example: 1
      produces:
      - application/json
      responses:
        "200":
          description: The block at a specific height
          schema:
            $ref: '#/definitions/BlockQuery'
        "400":
          description: Invalid height
        "404":
          description: Request block height doesn't
        "500":
          description: Server internal error
      summary: Get a block at a certain height
      tags:
      - Tendermint RPC
  /blocks/latest:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: The latest block
          schema:
            $ref: '#/definitions/BlockQuery'
        "500":
          description: Server internal error
      summary: Get the latest block
      tags:
      - Tendermint RPC
  /currencies/currency/{denom}:
    get:
      consumes:
      - application/json
      description: Get currency by denom
      operationId: currenciesGetCurrency
      parameters:
      - description: currency denomination symbol
        in: path
        name: denom
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.CCRespGetCurrency'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get currency
      tags:
      - Currencies
  /currencies/issue/{issueID}:
    get:
      consumes:
      - application/json
      description: Get currency issue by issueID
      operationId: currenciesGetIssue
      parameters:
      - description: issueID
        in: path
        name: issueID
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.CCRespGetIssue'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get currency issue
      tags:
      - Currencies
  /currencies/withdraw/{withdrawID}:
    get:
      consumes:
      - application/json
      description: Get currency withdraw by withdrawID
      operationId: currenciesGetWithdraw
      parameters:
      - description: withdrawID
        in: path
        name: withdrawID
        required: true
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.CCRespGetWithdraw'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get currency withdraw
      tags:
      - Currencies
  /currencies/withdraws:
    get:
      consumes:
      - application/json
      description: Get array of Withdraw objects with pagination
      operationId: currenciesGetWithdraws
      parameters:
      - description: 'page number (first page: 1)'
        in: query
        name: page
        type: integer
      - description: 'items per page (default: 100)'
        in: query
        name: limit
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.CCRespGetWithdraws'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get currency withdraws
      tags:
      - Currencies
  /distribution/community_pool:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Coin'
            type: array
        "500":
          description: Internal Server Error
      summary: Community pool parameters
      tags:
      - Distribution
  /distribution/delegators/{delegatorAddr}/rewards:
    get:
      description: Get the sum of all the rewards earned by delegations by a single delegator
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/DelegatorTotalRewards'
        "400":
          description: Invalid delegator address
        "500":
          description: Internal Server Error
      summary: Get the total rewards balance from all delegations
      tags:
      - Distribution
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos167w96tdvmazakdwkw2u57227eduula2cy572lf
    post:
      consumes:
      - application/json
      description: Withdraw all the delegator's delegation rewards
      parameters:
      - in: body
        name: Withdraw request body
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid delegator address
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Withdraw all the delegator's delegation rewards
      tags:
      - Distribution
  /distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}:
    get:
      description: Query a single delegation reward by a delegator
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Coin'
            type: array
        "400":
          description: Invalid delegator address
        "500":
          description: Internal Server Error
      summary: Query a delegation reward
      tags:
      - Distribution
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
    post:
      consumes:
      - application/json
      description: Withdraw a delegator's delegation reward from a single validator
      parameters:
      - in: body
        name: Withdraw request body
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid delegator address or delegation body
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Withdraw a delegation reward
      tags:
      - Distribution
  /distribution/delegators/{delegatorAddr}/withdraw_address:
    get:
      description: Get the delegations' rewards withdrawal address. This is the address in which the user will receive the reward funds
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Address'
        "400":
          description: Invalid delegator address
        "500":
          description: Internal Server Error
      summary: Get the rewards withdrawal address
      tags:
      - Distribution
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos167w96tdvmazakdwkw2u57227eduula2cy572lf
    post:
      consumes:
      - application/json
      description: Replace the delegations' rewards withdrawal address for a new one.
      parameters:
      - in: body
        name: Withdraw request body
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            withdraw_address:
              $ref: '#/definitions/Address'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid delegator or withdraw address
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Replace the rewards withdrawal address
      tags:
      - Distribution
  /distribution/parameters:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              base_proposer_reward:
                type: string
              bonus_proposer_reward:
                type: string
              community_tax:
                type: string
        "500":
          description: Internal Server Error
      summary: Fee distribution parameters
      tags:
      - Distribution
  /distribution/validators/{validatorAddr}:
    get:
      description: Query the distribution information of a single validator
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/ValidatorDistInfo'
        "400":
          description: Invalid validator address
        "500":
          description: Internal Server Error
      summary: Validator distribution information
      tags:
      - Distribution
    parameters:
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /distribution/validators/{validatorAddr}/outstanding_rewards:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Coin'
            type: array
        "500":
          description: Internal Server Error
      summary: Fee distribution outstanding rewards of a single validator
      tags:
      - Distribution
    parameters:
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /distribution/validators/{validatorAddr}/rewards:
    get:
      description: Query the commission and self-delegation rewards of validator.
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Coin'
            type: array
        "400":
          description: Invalid validator address
        "500":
          description: Internal Server Error
      summary: Commission and self-delegation rewards of a single validator
      tags:
      - Distribution
    parameters:
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
    post:
      consumes:
      - application/json
      description: Withdraw the validator's self-delegation and commissions rewards
      parameters:
      - in: body
        name: Withdraw request body
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid validator address
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Withdraw the validator's rewards
      tags:
      - Distribution
  /gov/parameters/deposit:
    get:
      description: Query governance deposit parameters. The max_deposit_period units are in nanoseconds.
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              max_deposit_period:
                example: "86400000000000"
                type: string
              min_deposit:
                items:
                  $ref: '#/definitions/Coin'
                type: array
            type: object
        "400":
          description: <other_path> is not a valid query request path
        "404":
          description: Found no deposit parameters
        "500":
          description: Internal Server Error
      summary: Query governance deposit parameters
      tags:
      - Governance
  /gov/parameters/tallying:
    get:
      description: Query governance tally parameters
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              governance_penalty:
                example: "0.0100000000"
                type: string
              threshold:
                example: "0.5000000000"
                type: string
              veto:
                example: "0.3340000000"
                type: string
        "400":
          description: <other_path> is not a valid query request path
        "404":
          description: Found no tally parameters
        "500":
          description: Internal Server Error
      summary: Query governance tally parameters
      tags:
      - Governance
  /gov/parameters/voting:
    get:
      description: Query governance voting parameters. The voting_period units are in nanoseconds.
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              voting_period:
                example: "86400000000000"
                type: string
        "400":
          description: <other_path> is not a valid query request path
        "404":
          description: Found no voting parameters
        "500":
          description: Internal Server Error
      summary: Query governance voting parameters
      tags:
      - Governance
  /gov/proposals:
    get:
      description: Query proposals information with parameters
      parameters:
      - description: voter address
        in: query
        name: voter
        required: false
        type: string
      - description: depositor address
        in: query
        name: depositor
        required: false
        type: string
      - description: proposal status, valid values can be '"deposit_period"', '"voting_period"', '"passed"', '"rejected"'
        in: query
        name: status
        required: false
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/TextProposal'
            type: array
        "400":
          description: Invalid query parameters
        "500":
          description: Internal Server Error
      summary: Query proposals
      tags:
      - Governance
    post:
      consumes:
      - application/json
      description: Send transaction to submit a proposal
      parameters:
      - description: valid value of '"proposal_type"' can be '"text"', '"parameter_change"', '"software_upgrade"'
        in: body
        name: post_proposal_body
        required: true
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            description:
              type: string
            initial_deposit:
              items:
                $ref: '#/definitions/Coin'
              type: array
            proposal_type:
              example: text
              type: string
            proposer:
              $ref: '#/definitions/Address'
            title:
              type: string
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: Tx was succesfully generated
          schema:
            $ref: '#/definitions/StdTx'
        "400":
          description: Invalid proposal body
        "500":
          description: Internal Server Error
      summary: Submit a proposal
      tags:
      - Governance
  /gov/proposals/{proposalId}:
    get:
      description: Query a proposal by id
      parameters:
      - in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/TextProposal'
        "400":
          description: Invalid proposal id
        "500":
          description: Internal Server Error
      summary: Query a proposal
      tags:
      - Governance
  /gov/proposals/{proposalId}/deposits:
    get:
      description: Query deposits by proposalId
      parameters:
      - in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Deposit'
            type: array
        "400":
          description: Invalid proposal id
        "500":
          description: Internal Server Error
      summary: Query deposits
      tags:
      - Governance
    post:
      consumes:
      - application/json
      description: Send transaction to deposit tokens to a proposal
      parameters:
      - description: proposal id
        in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      - description: ""
        in: body
        name: post_deposit_body
        required: true
        schema:
          properties:
            amount:
              items:
                $ref: '#/definitions/Coin'
              type: array
            base_req:
              $ref: '#/definitions/BaseReq'
            depositor:
              $ref: '#/definitions/Address'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid proposal id or deposit body
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Deposit tokens to a proposal
      tags:
      - Governance
  /gov/proposals/{proposalId}/deposits/{depositor}:
    get:
      description: Query deposit by proposalId and depositor address
      parameters:
      - description: proposal id
        in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      - description: Bech32 depositor address
        in: path
        name: depositor
        required: true
        type: string
        x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Deposit'
        "400":
          description: Invalid proposal id or depositor address
        "404":
          description: Found no deposit
        "500":
          description: Internal Server Error
      summary: Query deposit
      tags:
      - Governance
  /gov/proposals/{proposalId}/proposer:
    get:
      description: Query for the proposer for a proposal
      parameters:
      - in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Proposer'
        "400":
          description: Invalid proposal ID
        "500":
          description: Internal Server Error
      summary: Query proposer
      tags:
      - Governance
  /gov/proposals/{proposalId}/tally:
    get:
      description: Gets a proposal's tally result at the current time. If the proposal is pending deposits (i.e status 'DepositPeriod') it returns an empty tally result.
      parameters:
      - description: proposal id
        in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/TallyResult'
        "400":
          description: Invalid proposal id
        "500":
          description: Internal Server Error
      summary: Get a proposal's tally result at the current time
      tags:
      - Governance
  /gov/proposals/{proposalId}/votes:
    get:
      description: Query voters information by proposalId
      parameters:
      - description: proposal id
        in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Vote'
            type: array
        "400":
          description: Invalid proposal id
        "500":
          description: Internal Server Error
      summary: Query voters
      tags:
      - Governance
    post:
      consumes:
      - application/json
      description: Send transaction to vote a proposal
      parameters:
      - description: proposal id
        in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      - description: valid value of '"option"' field can be '"yes"', '"no"', '"no_with_veto"' and '"abstain"'
        in: body
        name: post_vote_body
        required: true
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            option:
              example: "yes"
              type: string
            voter:
              $ref: '#/definitions/Address'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid proposal id or vote body
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Vote a proposal
      tags:
      - Governance
  /gov/proposals/{proposalId}/votes/{voter}:
    get:
      description: Query vote information by proposal Id and voter address
      parameters:
      - description: proposal id
        in: path
        name: proposalId
        required: true
        type: string
        x-example: "2"
      - description: Bech32 voter address
        in: path
        name: voter
        required: true
        type: string
        x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Vote'
        "400":
          description: Invalid proposal id or voter address
        "404":
          description: Found no vote
        "500":
          description: Internal Server Error
      summary: Query vote
      tags:
      - Governance
  /gov/proposals/param_change:
    post:
      consumes:
      - application/json
      description: Generate a parameter change proposal transaction
      parameters:
      - description: The parameter change proposal body that contains all parameter changes
        in: body
        name: post_proposal_body
        required: true
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            changes:
              items:
                $ref: '#/definitions/ParamChange'
              type: array
            deposit:
              items:
                $ref: '#/definitions/Coin'
              type: array
            description:
              type: string
              x-example: Update max validators
            proposer:
              $ref: '#/definitions/Address'
            title:
              type: string
              x-example: Param Change
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: The transaction was succesfully generated
          schema:
            $ref: '#/definitions/StdTx'
        "400":
          description: Invalid proposal body
        "500":
          description: Internal Server Error
      summary: Generate a parameter change proposal transaction
      tags:
      - Governance
  /markets:
    get:
      consumes:
      - application/json
      description: Get array of Market objects with pagination and filters
      operationId: marketsGetMarketsWithParams
      parameters:
      - description: 'page number (first page: 1)'
        in: query
        name: page
        type: integer
      - description: 'items per page (default: 100)'
        in: query
        name: limit
        type: integer
      - description: BaseAsset denom filter
        in: query
        name: baseAssetDenom
        type: string
      - description: QuoteAsset denom filter
        in: query
        name: quoteAssetDenom
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.MarketsRespGetMarkets'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get markets
      tags:
      - Markets
  /markets/{marketID}:
    get:
      consumes:
      - application/json
      description: Get Market object by marketID
      operationId: marketsGetMarket
      parameters:
      - description: marketID
        in: path
        name: marketID
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.MarketsRespGetMarket'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get market
      tags:
      - Markets
  /minting/annual-provisions:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
        "500":
          description: Internal Server Error
      summary: Current minting annual provisions value
      tags:
      - Mint
  /minting/inflation:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
        "500":
          description: Internal Server Error
      summary: Current minting inflation value
      tags:
      - Mint
  /minting/parameters:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              blocks_per_year:
                type: string
              goal_bonded:
                type: string
              inflation_max:
                type: string
              inflation_min:
                type: string
              inflation_rate_change:
                type: string
              mint_denom:
                type: string
        "500":
          description: Internal Server Error
      summary: Minting module parameters
      tags:
      - Mint
  /multisig/call/{callID}:
    get:
      consumes:
      - application/json
      description: Get call object by it's ID
      operationId: multisigGetCall
      parameters:
      - description: call ID
        in: path
        name: callID
        required: true
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.MSRespGetCall'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get call
      tags:
      - Multisig
  /multisig/calls:
    get:
      consumes:
      - application/json
      description: Get active call objects
      operationId: multisigGetCalls
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.MSRespGetCalls'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get active calls
      tags:
      - Multisig
  /multisig/unique/{uniqueID}:
    get:
      consumes:
      - application/json
      description: Get call object by it's uniqueID
      operationId: multisigGetUniqueCall
      parameters:
      - description: call uniqueID
        in: path
        name: uniqueID
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.MSRespGetCall'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get call
      tags:
      - Multisig
  /node_info:
    get:
      description: Information about the connected node
      produces:
      - application/json
      responses:
        "200":
          description: Node status
          schema:
            properties:
              application_version:
                properties:
                  build_tags:
                    type: string
                  client_name:
                    type: string
                  commit:
                    type: string
                  go:
                    type: string
                  name:
                    type: string
                  server_name:
                    type: string
                  version:
                    type: string
              node_info:
                properties:
                  channels:
                    type: string
                  id:
                    type: string
                  listen_addr:
                    example: 192.168.56.1:26656
                    type: string
                  moniker:
                    example: validator-name
                    type: string
                  network:
                    example: gaia-2
                    type: string
                  other:
                    description: more information on versions
                    properties:
                      rpc_address:
                        example: tcp://0.0.0.0:26657
                        type: string
                      tx_index:
                        example: true
                        type: string
                    type: object
                  protocol_version:
                    properties:
                      app:
                        example: 0
                        type: string
                      block:
                        example: 10
                        type: string
                      p2p:
                        example: 7
                        type: string
                  version:
                    description: Tendermint version
                    example: 0.15.0
                    type: string
            type: object
        "500":
          description: Failed to query node status
      summary: The properties of the connected node
      tags:
      - Tendermint RPC
  /oracle/assets:
    get:
      consumes:
      - application/json
      description: Get asset objects
      operationId: oracleGetAssets
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OracleRespGetAssets'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get assets
      tags:
      - Oracle
  /oracle/currentprice/{assetCode}:
    get:
      consumes:
      - application/json
      description: Get current Price by assetCode
      operationId: oracleGetRawPrices
      parameters:
      - description: asset code
        in: path
        name: assetCode
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OracleRespGetPrice'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get current Price
      tags:
      - Oracle
  /oracle/rawprices:
    put:
      consumes:
      - application/json
      description: Send asset rawPrice signed Tx
      operationId: oraclePostPrice
      parameters:
      - description: PostPrice request with signed transaction
        in: body
        name: postRequest
        required: true
        schema:
          $ref: '#/definitions/rest.postPriceReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OracleRespGetAssets'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Post asset rawPrice
      tags:
      - Oracle
  /oracle/rawprices/{assetCode}/{blockHeight}:
    get:
      consumes:
      - application/json
      description: Get rawPrice objects by assetCode and blockHeight
      operationId: oracleGetRawPrices
      parameters:
      - description: asset code
        in: path
        name: assetCode
        required: true
        type: string
      - description: block height rawPrices relates to
        in: path
        name: blockHeight
        required: true
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OracleRespGetRawPrices'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get rawPrices
      tags:
      - Oracle
  /orders:
    get:
      consumes:
      - application/json
      description: Get array of Order objects with pagination and filters
      operationId: ordersGetOrdersWithParams
      parameters:
      - description: 'page number (first page: 1)'
        in: query
        name: page
        type: integer
      - description: 'items per page (default: 100)'
        in: query
        name: limit
        type: integer
      - description: owner filter
        in: query
        name: owner
        type: string
      - description: direction filter (bid/ask)
        in: query
        name: direction
        type: string
      - description: marketID filter (bid/ask)
        in: query
        name: marketID
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OrdersRespGetOrders'
        "400":
          description: Returned if the request doesn't have valid query/path params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get orders
      tags:
      - Orders
  /orders/{orderID}:
    get:
      consumes:
      - application/json
      description: Get Order object by orderID
      operationId: ordersGetOrder
      parameters:
      - description: orderID
        in: path
        name: orderID
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OrdersRespGetOrder'
        "400":
          description: Returned if the request doesn't have valid query/path params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get order
      tags:
      - Orders
  /orders/post:
    put:
      consumes:
      - application/json
      description: Post new order
      operationId: ordersPostOrder
      parameters:
      - description: PostOrder request with signed transaction
        in: body
        name: postRequest
        required: true
        schema:
          $ref: '#/definitions/rest.PostOrderReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OrdersRespPostOrder'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Post new order
      tags:
      - Orders
  /orders/revoke:
    put:
      consumes:
      - application/json
      description: Revoke order
      operationId: ordersRevokeOrder
      parameters:
      - description: RevokeOrder request with signed transaction
        in: body
        name: postRequest
        required: true
        schema:
          $ref: '#/definitions/rest.RevokeOrderReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.OrdersRespRevokeOrder'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Revoke order
      tags:
      - Orders
  /poa/validators:
    get:
      consumes:
      - application/json
      description: Get validator objects and required confirmations count
      operationId: poaValidators
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.PoaRespGetValidators'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get validators
      tags:
      - PoA
  /slashing/parameters:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              double_sign_unbond_duration:
                type: string
              downtime_unbond_duration:
                type: string
              max_evidence_age:
                type: string
              min_signed_per_window:
                type: string
              signed_blocks_window:
                type: string
              slash_fraction_double_sign:
                type: string
              slash_fraction_downtime:
                type: string
            type: object
        "500":
          description: Internal Server Error
      summary: Get the current slashing parameters
      tags:
      - Slashing
  /slashing/signing_infos:
    get:
      description: Get sign info of all validators
      parameters:
      - description: Page number
        in: query
        name: page
        required: true
        type: integer
        x-example: 1
      - description: Maximum number of items per page
        in: query
        name: limit
        required: true
        type: integer
        x-example: 5
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/SigningInfo'
            type: array
        "400":
          description: Invalid validator public key for one of the validators
        "500":
          description: Internal Server Error
      summary: Get sign info of given all validators
      tags:
      - Slashing
  /slashing/validators/{validatorAddr}/unjail:
    post:
      consumes:
      - application/json
      description: Send transaction to unjail a jailed validator
      parameters:
      - description: Bech32 validator address
        in: path
        name: validatorAddr
        required: true
        type: string
        x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
      - description: ""
        in: body
        name: UnjailBody
        required: true
        schema:
          properties:
            base_req:
              $ref: '#/definitions/StdTx'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: Tx was succesfully generated
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid validator address or base_req
        "500":
          description: Internal Server Error
      summary: Unjail a jailed validator
      tags:
      - Slashing
  /staking/delegators/{delegatorAddr}/delegations:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Delegation'
            type: array
        "400":
          description: Invalid delegator address
        "500":
          description: Internal Server Error
      summary: Get all delegations from a delegator
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    post:
      consumes:
      - application/json
      parameters:
      - description: The password of the account to remove from the KMS
        in: body
        name: delegation
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            delegation:
              $ref: '#/definitions/Coin'
            delegator_address:
              $ref: '#/definitions/Address'
            validator_address:
              $ref: '#/definitions/ValidatorAddress'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid delegator address or delegation request body
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Submit delegation
      tags:
      - Staking
  /staking/delegators/{delegatorAddr}/delegations/{validatorAddr}:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Delegation'
        "400":
          description: Invalid delegator address or validator address
        "500":
          description: Internal Server Error
      summary: Query the current delegation between a delegator and a validator
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /staking/delegators/{delegatorAddr}/redelegations:
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    post:
      consumes:
      - application/json
      parameters:
      - description: The sender and tx information
        in: body
        name: delegation
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            delegator_address:
              $ref: '#/definitions/Address'
            shares:
              example: "100"
              type: string
            validator_dst_address:
              $ref: '#/definitions/ValidatorAddress'
            validator_src_addressess:
              $ref: '#/definitions/ValidatorAddress'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: Tx was succesfully generated
          schema:
            $ref: '#/definitions/StdTx'
        "400":
          description: Invalid delegator address or redelegation request body
        "500":
          description: Internal Server Error
      summary: Submit a redelegation
      tags:
      - Staking
  /staking/delegators/{delegatorAddr}/unbonding_delegations:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/UnbondingDelegation'
            type: array
        "400":
          description: Invalid delegator address
        "500":
          description: Internal Server Error
      summary: Get all unbonding delegations from a delegator
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    post:
      consumes:
      - application/json
      parameters:
      - description: The password of the account to remove from the KMS
        in: body
        name: delegation
        schema:
          properties:
            base_req:
              $ref: '#/definitions/BaseReq'
            delegator_address:
              $ref: '#/definitions/Address'
            shares:
              example: "100"
              type: string
            validator_address:
              $ref: '#/definitions/ValidatorAddress'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "400":
          description: Invalid delegator address or unbonding delegation request body
        "401":
          description: Key password is wrong
        "500":
          description: Internal Server Error
      summary: Submit an unbonding delegation
      tags:
      - Staking
  /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/UnbondingDelegationPair'
        "400":
          description: Invalid delegator address or validator address
        "500":
          description: Internal Server Error
      summary: Query all unbonding delegations between a delegator and a validator
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /staking/delegators/{delegatorAddr}/validators:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Validator'
            type: array
        "400":
          description: Invalid delegator address
        "500":
          description: Internal Server Error
      summary: Query all validators that a delegator is bonded to
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
  /staking/delegators/{delegatorAddr}/validators/{validatorAddr}:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Validator'
        "400":
          description: Invalid delegator address or validator address
        "500":
          description: Internal Server Error
      summary: Query a validator that a delegator is bonded to
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: path
      name: delegatorAddr
      required: true
      type: string
      x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
    - description: Bech32 ValAddress of Delegator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /staking/parameters:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              bond_denom:
                type: string
              goal_bonded:
                type: string
              inflation_max:
                type: string
              inflation_min:
                type: string
              inflation_rate_change:
                type: string
              max_validators:
                type: integer
              unbonding_time:
                type: string
            type: object
        "500":
          description: Internal Server Error
      summary: Get the current staking parameter values
      tags:
      - Staking
  /staking/pool:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            properties:
              bonded_tokens:
                type: string
              date_last_commission_reset:
                type: string
              inflation:
                type: string
              inflation_last_time:
                type: string
              loose_tokens:
                type: string
              prev_bonded_shares:
                type: string
            type: object
        "500":
          description: Internal Server Error
      summary: Get the current state of the staking pool
      tags:
      - Staking
  /staking/redelegations:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Redelegation'
            type: array
        "500":
          description: Internal Server Error
      summary: Get all redelegations (filter by query params)
      tags:
      - Staking
    parameters:
    - description: Bech32 AccAddress of Delegator
      in: query
      name: delegator
      required: false
      type: string
    - description: Bech32 ValAddress of SrcValidator
      in: query
      name: validator_from
      required: false
      type: string
    - description: Bech32 ValAddress of DstValidator
      in: query
      name: validator_to
      required: false
      type: string
  /staking/validators:
    get:
      parameters:
      - description: The validator bond status. Must be either 'bonded', 'unbonded', or 'unbonding'.
        in: query
        name: status
        type: string
        x-example: bonded
      - description: The page number.
        in: query
        name: page
        type: integer
        x-example: 1
      - description: The maximum number of items per page.
        in: query
        name: limit
        type: integer
        x-example: 1
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Validator'
            type: array
        "500":
          description: Internal Server Error
      summary: Get all validator candidates. By default it returns only the bonded validators.
      tags:
      - Staking
  /staking/validators/{validatorAddr}:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Validator'
        "400":
          description: Invalid validator address
        "500":
          description: Internal Server Error
      summary: Query the information from a single validator
      tags:
      - Staking
    parameters:
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /staking/validators/{validatorAddr}/delegations:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/Delegation'
            type: array
        "400":
          description: Invalid validator address
        "500":
          description: Internal Server Error
      summary: Get all delegations from a validator
      tags:
      - Staking
    parameters:
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /staking/validators/{validatorAddr}/unbonding_delegations:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/UnbondingDelegation'
            type: array
        "400":
          description: Invalid validator address
        "500":
          description: Internal Server Error
      summary: Get all unbonding delegations from a validator
      tags:
      - Staking
    parameters:
    - description: Bech32 OperatorAddress of validator
      in: path
      name: validatorAddr
      required: true
      type: string
      x-example: cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l
  /supply/total:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/Supply'
        "500":
          description: Internal Server Error
      summary: Total supply of coins in the chain
      tags:
      - Supply
  /supply/total/{denomination}:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
        "400":
          description: Invalid coin denomination
        "500":
          description: Internal Server Error
      summary: Total supply of a single coin denomination
      tags:
      - Supply
    parameters:
    - description: Coin denomination
      in: path
      name: denomination
      required: true
      type: string
      x-example: uatom
  /syncing:
    get:
      description: Get if the node is currently syning with other nodes
      produces:
      - application/json
      responses:
        "200":
          description: Node syncing status
          schema:
            properties:
              syncing:
                type: boolean
            type: object
        "500":
          description: Server internal error
      summary: Syncing state of node
      tags:
      - Tendermint RPC
  /txs:
    get:
      description: Search transactions by events.
      parameters:
      - description: 'transaction events such as ''message.action=send'' which results in the following endpoint: ''GET /txs?message.action=send''. note that each module documents its own events. look for xx_events.md in the corresponding cosmos-sdk/docs/spec directory'
        in: query
        name: message.action
        type: string
        x-example: send
      - description: 'transaction tags with sender: ''GET /txs?message.action=send&message.sender=cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv'''
        in: query
        name: message.sender
        type: string
        x-example: cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv
      - description: Page number
        in: query
        name: page
        type: integer
        x-example: 1
      - description: Maximum number of items per page
        in: query
        name: limit
        type: integer
        x-example: 1
      - description: transactions on blocks with height greater or equal this value
        in: query
        name: tx.minheight
        type: integer
        x-example: 25
      - description: transactions on blocks with height less than or equal this value
        in: query
        name: tx.maxheight
        type: integer
        x-example: 800000
      produces:
      - application/json
      responses:
        "200":
          description: All txs matching the provided events
          schema:
            $ref: '#/definitions/PaginatedQueryTxs'
        "400":
          description: Invalid search events
        "500":
          description: Internal Server Error
      summary: Search transactions
      tags:
      - Transactions
    post:
      consumes:
      - application/json
      description: Broadcast a signed tx to a full node
      parameters:
      - description: The tx must be a signed StdTx. The supported broadcast modes include '"block"'(return after tx commit), '"sync"'(return afer CheckTx) and '"async"'(return right away).
        in: body
        name: txBroadcast
        required: true
        schema:
          properties:
            mode:
              example: block
              type: string
            tx:
              $ref: '#/definitions/StdTx'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: Tx broadcasting result
          schema:
            $ref: '#/definitions/BroadcastTxCommitResult'
        "500":
          description: Internal Server Error
      summary: Broadcast a signed tx
      tags:
      - Transactions
  /txs/{hash}:
    get:
      description: Retrieve a transaction using its hash.
      parameters:
      - description: Tx hash
        in: path
        name: hash
        required: true
        type: string
        x-example: BCBE20E8D46758B96AE5883B792858296AC06E51435490FBDCAE25A72B3CC76B
      produces:
      - application/json
      responses:
        "200":
          description: Tx with the provided hash
          schema:
            $ref: '#/definitions/TxQuery'
        "500":
          description: Internal Server Error
      summary: Get a Tx by hash
      tags:
      - Transactions
  /txs/decode:
    post:
      consumes:
      - application/json
      description: Decode a transaction (signed or not) from base64-encoded Amino serialized bytes to JSON
      parameters:
      - description: The tx to decode
        in: body
        name: tx
        required: true
        schema:
          properties:
            tx:
              example: SvBiXe4KPqijYZoKFFHEzJ8c2HPAfv2EFUcIhx0yPagwEhTy0vPA+GGhCEslKXa4Af0uB+mfShoMCgVzdGFrZRIDMTAwEgQQwJoM
              type: string
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: The tx was successfully decoded
          schema:
            $ref: '#/definitions/StdTx'
        "400":
          description: The tx was malformated
        "500":
          description: Server internal error
      summary: Decode a transaction from the Amino wire format
      tags:
      - Transactions
  /txs/encode:
    post:
      consumes:
      - application/json
      description: Encode a transaction (signed or not) from JSON to base64-encoded Amino serialized bytes
      parameters:
      - description: The tx to encode
        in: body
        name: tx
        required: true
        schema:
          properties:
            tx:
              $ref: '#/definitions/StdTx'
          type: object
      produces:
      - application/json
      responses:
        "200":
          description: The tx was successfully decoded and re-encoded
          schema:
            properties:
              tx:
                example: The base64-encoded Amino-serialized bytes for the tx
                type: string
            type: object
        "400":
          description: The tx was malformated
        "500":
          description: Server internal error
      summary: Encode a transaction to the Amino wire format
      tags:
      - Transactions
  /validatorsets/{height}:
    get:
      parameters:
      - description: Block height
        in: path
        name: height
        required: true
        type: number
        x-example: 1
      produces:
      - application/json
      responses:
        "200":
          description: The validator set at a specific block height
          schema:
            properties:
              block_height:
                type: string
              validators:
                items:
                  $ref: '#/definitions/TendermintValidator'
                type: array
            type: object
        "400":
          description: Invalid height
        "404":
          description: Block at height not available
        "500":
          description: Server internal error
      summary: Get a validator set a certain height
      tags:
      - Tendermint RPC
  /validatorsets/latest:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: The validator set at the latest block height
          schema:
            properties:
              block_height:
                type: string
              validators:
                items:
                  $ref: '#/definitions/TendermintValidator'
                type: array
            type: object
        "500":
          description: Server internal error
      summary: Get the latest validator set
      tags:
      - Tendermint RPC
  /vm/compile:
    get:
      consumes:
      - application/json
      description: Compile script / module code using VM and return byteCode
      operationId: vmCompile
      parameters:
      - description: Code with metadata
        in: body
        name: getRequest
        required: true
        schema:
          $ref: '#/definitions/rest.compileReq'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.VmRespCompile'
        "400":
          description: Returned if the request doesn't have valid query params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get compiled byteCode
      tags:
      - VM
  /vm/data/{accountAddr}/{vmPath}:
    get:
      consumes:
      - application/json
      description: Get data from data source by accountAddr and path
      operationId: vmGetData
      parameters:
      - description: account address (Libra HEX  Bech32)
        in: path
        name: accountAddr
        required: true
        type: string
      - description: VM path (HEX string)
        in: path
        name: vmPath
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.VmData'
        "422":
          description: Returned if the request doesn't have valid path params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get data from data source
      tags:
      - VM
  /vm/tx/{txHash}:
    get:
      consumes:
      - application/json
      description: Get tx VM execution status by tx hash
      operationId: vmTxStatus
      parameters:
      - description: transaction hash
        in: path
        name: txHash
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/rest.VmTxStatus'
        "422":
          description: Returned if the request doesn't have valid path params
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
        "500":
          description: Returned on server error
          schema:
            $ref: '#/definitions/rest.ErrorResponse'
      summary: Get tx VM execution status
      tags:
      - VM
schemes:
- https
securityDefinitions:
  kms:
    type: basic
swagger: "2.0"
tags:
- description: Search, encode, or broadcast transactions.
  name: Transactions
- description: Tendermint APIs, such as query blocks, transactions and validatorset
  name: Tendermint RPC
- description: Authenticate accounts
  name: Auth
- description: Create and broadcast transactions
  name: Bank
- description: Stake module APIs
  name: Staking
- description: Governance module APIs
  name: Governance
- description: Slashing module APIs
  name: Slashing
- description: Fee distribution module APIs
  name: Distribution
- description: Supply module APIs
  name: Supply
- name: version
- description: Minting module APIs
  name: Mint
- description: Query app version
  name: Misc
`
