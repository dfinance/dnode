# Events

[Cosmos SDK Events overview](https://docs.cosmos.network/master/core/events.html).

Almost every DN module emits events on successful Tx execution.
There are some useful cross module queries:

* `message.module='{moduleName}'` - filter out module events;

## `Multisig` module

* A new Call submitted event

    Type: `multisig.submit_call`
    
    Attributes:
    - `msg_type` - multi signature message type [string];
    - `msg_route` - multi signature message routing [string];
    - `call_id` - created call ID [uint];
    - `unique_id` - call unique ID [string];
    - `sender` - call creator address [bech32 string];

* Call removed from the queue (call was executed or last vote was revoked)

    Type: `multisig.remove_call`
    
    Attributes:
    - `call_id` - removed call ID [uint];
    
* Call state updated

    Type: `multisig.update_call`
    
    Attributes:
    - `call_id` - removed call ID [uint];
    - `call_state` - removed call ID [approved / rejected / failed / executed];
    
    **Approved** - Call voting ended successfully, message handler would be executed;
    
    **Failed** - Call is Approved, but message handler finished with an error;
    
    **Executed** - Call is Approved and message handler finished without an error;
    
    **Rejected** - Call voting stopped by blocks timeout;

* Call confirmed (vote added)

    Type: `multisig.confirm_vote`
    
    Attributes:
    - `call_id` - call ID [uint];
    - `sender` - vote sender address [Bech32 string];

* Call vote revoked

    Type: `multisig.revoke_vote`
    
    Attributes:
    - `call_id` - call ID [uint];
    - `sender` - revoked vote sender address [Bech32 string];

## `PoA` module

* Validator added (also triggered on validator replacement)

    Type: `poa.add`
    
    Attributes:
    - `address` - sdk address [Bech32 string];
    - `eth_address` - Ethereum address [string];

* Validator removed (also triggered on validator replacement)

    Type: `poa.remove`
    
    Attributes:
    - `address` - sdk address [Bech32 string];
    - `eth_address` - Ethereum address [string];

## `Ccstorage` module

* Currency created after gov proposal approval

    Type: `ccstorage.create`
    
    Attributes:
    - `denom` - Currency denomination symbol [string];
    - `decimals` - Currency decimals count [uint];
    - `info_path` - VM CurrencyInfo path [HEX string];

## `Currencies` module

* Currency issued after multi signature approval

    Type: `currencies.issue`
    
    Attributes:
    - `issue_id` - Issue unique ID [string];
    - `denom` - Currency denomination symbol [string];
    - `amount` - Issue amount [uint];
    - `sender` - payee account [Bech32 string];

* Currency withdraw action

    Type: `currencies.withdraw`
    
    Attributes:
    - `withdraw_id` - Withdraw ID [uint];
    - `denom` - Currency denomination symbol [string];
    - `amount` - Withdraw amount [uint];
    - `sender` - spender account [Bech32 string];

## `Markets` module

* Market created

    Type: `markets.create`
    
    Attributes:
    - `market_id` - Market ID [uint];
    - `base_denom` - BaseAsset denomination symbol [string];
    - `quote_denom` - QuoteAsset denomination symbol [string];

## `Orders` module

* Order posted

    Type: `orders.post`
    
    Attributes:
    - `owner` - Order creator address [Bech32 string];
    - `market_id` - Market ID [uint];
    - `order_id` - Order ID [uint];
    - `direction` - Order type [bid/ask];
    - `price` - price [uint];
    - `quantity` - quantity [uint];

* Order canceled (manually revoked or by TTL)

    Type: `orders.cancel`
    
    Attributes:
    - `owner` - Order creator address [Bech32 string];
    - `market_id` - Market ID [uint];
    - `order_id` - Order ID [uint];
    - `direction` - Order type [bid/ask];
    - `price` - price [uint];
    - `quantity` - quantity [uint];

* Order fully filled and removed

    Type: `orders.full_fill`
    
    Attributes:
    - `owner` - Order creator address [Bech32 string];
    - `market_id` - Market ID [uint];
    - `order_id` - Order ID [uint];
    - `direction` - Order type [bid/ask];
    - `price` - price [uint];
    - `quantity` - quantity [uint];

* Order partially filled, quantity updated

    Type: `orders.partial_fill`
    
    Attributes:
    - `owner` - Order creator address [Bech32 string];
    - `market_id` - Market ID [uint];
    - `order_id` - Order ID [uint];
    - `direction` - Order type [bid/ask];
    - `price` - price [uint];
    - `quantity` - quantity [uint];

## `Orderbook` module

* ClearanceState updated for Market (matching event)

    Type: `orderbook.clearance`
    
    Attributes:
    - `market_id` - Market ID [uint];
    - `price` - clearance price [uint];

## `Oracle` module

* New asset added

    Type: `oracle.add_asset`
    
    Attributes:
    - `asset_code` - new asset assetCode [string];

* Price updated for assetCode

    Type: `oracle.price`
    
    Attributes:
    - `asset_code` - assetCode [string];
    - `price` - updated price [int];
    - `received_at` - price received UNIX timestamp (in seconds) by oracles system [int];

## `VM` module

Depending on VM execution status, module emits multiple events per Tx with variadic number of attributes.

* VM execution status `keep` received

    * Contract status event ("keep" event)
    
        Type: `vm.contract_status`
    
        Attributes:
        - `status` - `keep` [string const];

    * VM events (number of events depends on execution status)
    
        Type: `vm.contract_events`
        
        Attributes:
        - `sender_address` - VM event sender address [`0x1` for stdlib / Bech32 string for account resource];
        - `source` - VM event source [`script` for script source / `{moduleAddress}::{moduleName}` for module source];
        - `type` - VM event type string representation in Move format [string];
        - `data` - HEX string VM event data representation [string];

* VM execution status `keep` received (failed with an error)

    1. "keep" event
        
        Type: `vm.contract_status`
    
        Attributes:
        - `status` - `keep` [string const];

    2. "error" event

        Type: `vm.contract_status`
    
        Attributes:
        - `status` - `error` [string const];
        - `major_status` - error majorStatus [uint];
        - `sub_status` - error subStatus [uint];
        - `message` - error message [string];

* VM execution status `discard` received

    Type: `vm.contract_status`
    
    Attributes:
    - `status` - `discard` [string const];

* VM execution status `discard` received (failed with an error)

    Type: `vm.contract_status`
    
    Attributes:
    - `status` - `discard` [string const];
    - `major_status` - error majorStatus [uint];
    - `sub_status` - error subStatus [uint];
    - `message` - error message [string];
 