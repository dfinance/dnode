# Peg Zone

## Add/remove/replace validator by multisignature

Before we start managing validators by PoA, let's remember that minimum amount of validators is 3, maximum is 11.

To add new validator use next command:

    dncli tx poa ms-add-validator [validator-address] [eth-address] --validator-1

* **[validator-address]** - DN bench32 validator address;
* **[eth-address]** - validator ethereum address;

To remove:

    dncli tx poa ms-remove-validator [validator-address] --from validator1

To replace:

    dncli tx poa ms-replace-validator [old-address] [new-address] [eth-address] --from validator-1

To get validators list (include their amount and required confirmations amount to execute call):

    dncli query poa validators

To get minimum/maximum validators amount:

    dncli query poa minmax

To get validator:

    dncli query poa validator [address]

Where `[address]` is Bech32 DN address.

## Confirm multisignature call

To confirm multisignature call you need to extract call id from transaction execution output and confirm this call
by other validators:

    dncli tx multisig confirm-call [call-id]

Once call submited under multisignature, there is **86400** blocks interval to confirm it by other validators, if call
not confirmed by that time, it will be marked as rejected.

To revoke confirmation from call:

    dncli tx multisig revoke-confirm [call-id]

Once call reaches **N/2+1** amount of confirmations, message inside call will be executed.

To get call information:

    dncli query multisig call [call-id]

To get calls amount:

    dncli query multisig lastId

## Issuing new currency by multisig

To issue new currency:

    dncli tx currencies ms-issue [issueID] [coin] [recipient]  --from validators1

* **issueID** - call unique ID, required to prevent double spend on issuing new currencies, usually it's sha256(chainId + symbol + txHash), serialized to hex;
* **coin** - issue denomination symbol and amount in coin format (100xfi);
* **recipient** - DN address of account who's receiving coins;

To withdraw currency from any account call:

    dncli tx currencies withdraw [denom] [coin] [pegZoneSpender] --from account

To get issued currencies demons/symbols:

    dncli query currencies currency [symbol]

To get specific issue info:

    dncli query currencies issue [issueID]

To get withdraw list:

    dncli query currencies withdraws --page=1 --limit=100

* **[page]** - page number (optional)
* **[limit]** - limit of objects per page (optional)

To get withdraw by ID:

    dncli query currencies withdraw [withdrawID]

* **[withdrawID]** - withdraw ID, usually just from 0 to N.
