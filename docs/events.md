# Events

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
