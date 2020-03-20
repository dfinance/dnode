# Oracle App

Oracle app fetching data from exchanges and place prices to Dfinance blockchain.

## Installation 

To install oracle app as binary:

    git clone https://github.com/dfinance/dnode.git
    cd node
    make install-oracleapp
    mkdir ~/.oracle-app
    oracle-app init
    
You can try help:

    oracle-app help    
    
### Launch

**IMPORTANT:** this actions works only if you have access to oracle account of specific asset.

To launch oracle app first open a configuration:

    nano ~/.oracle-app/config.yaml

And put next lines to end of file to enable oracles for **btc_usdt**, **eth_usdt** (binance) pairs,
and simulation of **dfi_eth** and **dfi_btc** prices:

```yaml
exchanges:
  - binance:
      - code: eth_usdt
        pair:
          base: ETH
          quote: USDT
      - code: btc_usdt
        pair:
          base: BTC
          quote: USDT
  - dfinance-test:
      - code: dfi_eth
        pair:
          base: dfi
          quote: eth
        simulate:
          enabled: true
          min_price: 869000
          max_price: 871000
          period_s: 1
      - code: dfi_btc
        pair:
          base: dfi
          quote: btc
        simulate:
          enabled: true
          min_price: 18000
          max_price: 20000
          period_s: 1
```

Save file.
Now it's possible to launch:

    DN_ORACLEAPP_MNEMONIC="<mnemonic>" oracle-app start
    
Where `<mnemonic>` must be replaced with your mnemonic from account with oracle status.
The rest of parameters:

    * DN_ORACLEAPP_ACCOUNT - number of account in standard hd path ((44'/118'/{account}'/0/{index}).
    * DN_ORACLEAPP_INDEX - number of index of account in standard hd path ((44'/118'/{account}'/0/{index}).

