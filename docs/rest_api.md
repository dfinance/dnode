# Rest API

Launch REST API:

    dncli rest-server --chain-id dn-testnet --trust-node

To configure Swagger web application you should change the following parameters in the `~/.dncli/config` file:
* `swagger-host = "127.0.0.1:1317"` - host that handles HTTP requests from the Swagger interface;
* `swagger-allowed-urls = ["http://127.0.0.1:1317"]` - CORS allowed URLs for Swagger web interface (if `swagger-host`
matches the Swagger host, this parameter should be empty `[]`);
* `swagger-schemes = ["http", "https"]` - available schemes (those would appear on the web interface); 

Swagger-UI API specification endpoint is available when REST server is up and running:
* `http://localhost:1317/swagger-ui/` 
* `http://localhost:1317/swagger-ui/?urls.primaryName=Cosmos%20SDK%20API` - Cosmos SDK / Tendermint specification 
* `http://localhost:1317/swagger-ui/?urls.primaryName=Dfinance%20API`     - Dfinance custom modules specification 

All REST API returns JSON.

Multisig:

* `/multisig/call/{id}` - Get call by id.
* `/multisig/calls` - Get array of active calls (that waiting for confirmations)
* `/multisig/unique/{unique}` - Get call by unique id.

Currencies:

* `/currencies/issue/{issueID}` - Get issue operation by issue id.
* `/currencies/currency/{symbol}` - Get currency info by symbol.
* `/currencies/withdraw/{withdrawID}` - Get withdraw info by id.
* `/currencies/withdraws?page={page}&limit={limit}` - Get withdraw list, page/limit parameters are optional.

PoA:

* `/poa/validators` - PoA validators list.

Oracle:
* `/oracle/rawprices` - Post price from Oracle.
* `/oracle/rawprices/{assetCode}/{blockHeight}` - Get unprocessed prices for assetCode and blockHeight.
* `/oracle/currentprice/{assetCode}` - Get current price for assetCode.
* `/oracle/assets` - Get array of assets.

VM:
* `/vm/compile-script` - Get compiled bytecode for VM script.
* `/vm/compile-module` - Get compiled bytecode for VM module.