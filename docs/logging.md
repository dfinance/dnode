# Logging

## Helpful links
* [Tendermint. How to read logs / default modules](https://github.com/tendermint/tendermint/blob/v0.33.3/docs/tendermint-core/how-to-read-logs.md)
* [Tendermint. Prod logging](https://github.com/tendermint/tendermint/blob/v0.33.3/docs/tendermint-core/running-in-production.md#logging)

## Changing the log level

    dnode start --log_level "main:info,state:info,*:error,x/vm:debug"
    
Argument above defines:
* `info` log level for `main` (the main app) and `state` modules;
* `debug` log level for `x/vm` module;
* `error` as a default log level for other modules;

Default log level value: `main:info,state:info,*:error`

## DNode specific modules

* `x/vm` - VM module;
* `x/vm/dsserver` - VM data server module;
* `x/multisig` - multisig module;
