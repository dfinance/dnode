# Rest API

Launch REST API:

    dncli rest-server --chain-id dn-testnet --trust-node

Swagger-UI API specification endpoint is available when REST server is up and running:
* `http://localhost:1317/swagger-ui/` 
* `http://localhost:1317/swagger-ui/?urls.primaryName=Cosmos%20SDK%20API` - Cosmos SDK / Tendermint specification 
* `http://localhost:1317/swagger-ui/?urls.primaryName=Dfinance%20API`     - Dfinance custom modules specification 

All REST API returns JSON.

Multisig:

* `/multisig/call/{id}`  - get call by id.
* `/multisig/calls`      - get array of active calls (that waiting for confirmations)
* `/multisig/unique/{unique}` - get call by unique id.

Currencies:

* `/currencies/issue/{issueID}` - Get issue operation by issue id.
* `/currencies/currency/{symbol}` - Get currency info by symbol.
* `/currencies/destroy/{destroyID}` - Get destroy info by id.
* `/currencies/destroys/{page}?limit={limit}` - Get destroys list, limit parameter is optional, equal 100 by default.

PoA:

* `/poa/validators` - PoA validators list.

