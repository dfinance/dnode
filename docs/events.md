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