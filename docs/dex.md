# DEX

DEX is a combination of markets, orders and orderbook modules which provides market exchange feature.
Client can post market orders to buy (Bid order) / sell (Ask order) an asset (Base asset) in exchange of an other asset (Quote asset).

## Markets

All orders are linked to a corresponding Market.
Market defines which asset acts as a Base asset and which as a Quote asset.

**Base asset** - the target buy / sell asset; order quantity is specified in Base asset amount.

**Quote asset** - the payment asset (Base asset exchanged for); order price is specified in Quote asset amount.

### Add

To create a Market corresponding currencies (Base, Quote) must be registered in the chain.
Here is an example of creating the Market using CLI:

    dncli tx markets add btc xfi --from {accountAddress}

* `btc` - Base asset;
* `xfi` - Quote asset;

### Query

To query an existing Market(s) we have two options.

1. Query specific Market by ID:

        dncli query markets market {marketID}

2. Query all / filtered Markets:

        dncli markets list --page=1 --limit=10 --base-asset-denom=btc --quote-asset-denom=xfi

    * `page, limit` - pagination arguments (optional);
    * `base-asset-denom` - filter by Base asset (optional);
    * `quote-asset-denom` - filter by Base asset (optional);

## Orders

An Order defines will to buy / sell Base asset currency.

Order has the following fields:
* `owner` - order creator address;
* `direction` - `ask` (Sell order) / `bid` (Buy order);
* `price` - minimum (Ask) / maximum (Bid) amount of Quote asset currency;
* `quantity` - Base asset currency amount;
* `ttl` - time to live time interval (in seconds) after which order would be auto-canceled;
* `createdAt` - order creation timestamp;
* `updatedAt` - last partial fill timestamp (equal to `createdAt` for newly created orders);

Client can't post an order if his funds are insufficient to lock Base / Quote currency amount.
For Ask orders Base `quantity` is locked.
For Bid orders Base `quantity` multiplied by `price` of Quote currency is locked.

### Creation

Order creation example:

    dncli orders post 0 bid 500000 10000000000000 --from {accountAddress}
    
* `0` - marketID;
* `bid` - order direction;
* `500000` - price;
* `10000000000000` - quantity;

**Important**

Price / quantity values should be defined bering in mind currency decimals.
For example, we are posting an order for btc-xfi market with marketID `0`.
BTC is a Base asset currency with 8 decimals.
XFI is a Quote asset currency with 18 decimals.
In that case, the example above can interpreted as:
* `500000` BTCs -> `0.005` portion of BTC;
* `10000000000000` XFIs -> `0.00001` portions of XFI;

### Revoking

Order is auto-revoked by TTL timeout.
Order can also be manually revoked by its owner, example:

    dncli orders revoke 0 --from wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m
    
* `0` - orderID;

On revoking the unfilled currency quantity is unlocked.

### Query

To query an existing Order(s) we have two options.

1. Query specific Order by ID:

        dncli query orders order {orderID}

2. Query all / filtered Orders:

        dncli orders list --page=1 --limit=10 --direction=ask --owner={accountAddress} --market-id=0

    * `page, limit` - pagination arguments (optional);
    * `direction` - filter by Order direction (optional);
    * `owner` - filter by Order owner (optional);
    * `market-id` - filter by Market (optional);

## Matching

Matching is a process of acquiring a Clearance state.
Depending on Clearance state price and maximum bid / ask quantities, orders can be fully or partially filled.

### Full order fill

If 100% of order's quantity filled, order would be removed.
For Ask orders, Quote currency is transferred to client.
For Bif orders, Base currency is transferred to client.

### Partial order fill

If some % of order's quantity is filled, order is updated and preserved in the orders pool.
Order can be fully filled later or removed by TTL (manually).

### Refund

If Clearance state price is lower than Bid order's price, client gets the refund.
Price difference is transferred to client.

**Important**

If the refund amount is lower that the minimal Quote currency amount, refund is omitted.
