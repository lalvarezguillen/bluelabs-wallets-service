# BlueLabs Wallets Service Challenge


## Part 1

### Creating a wallet

```
$ curl -i -X POST host:port/wallets -H 'Content-Type:application/json' -d '{"name":"name for the wallet"}' 

Content-Type:application/json'
HTTP/1.1 201 Created
Content-Type: application/json; charset=UTF-8
Date: Sun, 12 Sep 2021 17:01:59 GMT
Content-Length: 86

{"id":1,"created_at":"0001-01-01T00:00:00Z","name":"name for the wallet","balance":0}
```

### Adding funds from a wallet

```
$ curl -i -X POST host:port/wallets/1/balance-changes -H 'Content-Type:application/json' -d '{"operation": "ADD", "amount": 300}

HTTP/1.1 201 Created
Content-Type: application/json; charset=UTF-8
Date: Sun, 12 Sep 2021 17:04:26 GMT
Content-Length: 144

{"id":3,"created_at":"0001-01-01T00:00:00Z","amount":300,"operation":"ADD","balance_before":0,"balance_after":300,"reference":"","wallet_id":1}
```

### Removing funds from a wallet

* Returns 400 if the wallet doesn't have enough balance (IE: prevents the wallet from having negative balance)
* Locks the Wallet DB entry at the begining of the operation, making sure that "a user can’t spend the same funds twice"
* A DB transaction guarantees that either the whole operation succeeds (creating the `BalanceChange` entry and updating the
`Wallet`'s balance, or no part of the operation suceeds)

```
$ curl -i -X POST host:port/wallets/1/balance-changes -H 'Content-Type:application/json' -d '{"operation": "SUBSTRACT", "amount": 100, "reference": "important payment"}'

HTTP/1.1 201 Created
Content-Type: application/json; charset=UTF-8
Date: Sun, 12 Sep 2021 17:06:10 GMT
Content-Length: 152

{"id":4,"created_at":"0001-01-01T00:00:00Z","amount":100,"operation":"SUBSTRACT","balance_before":300,"balance_after":200,"reference":"important payment","wallet_id":1}
```

### Performance

Modifying balance does involve creating an extra `BalanceChange` DB entry, besides the update to the `Wallet` entry. Which means trading
some performance in favor of being able to track balance changes, which might prove important on the second part of this challenge

### Graceful shutdown

The server handles `SIGINT` signals (the ones sent when `ctrl+c` is hit), and `SIGTERM` (the one sent by `docker stop`), and
schedules a graceful shutdown in those scenarios.


## Part 2

Lets use the following names:

* The service built on Part 1 of this challenge will be called `Wallets Service`
* Our service that deals with PayPal will be called `PayPalWithdrawals Service`

In all the scenarios below, `PayPalWithdrawals Service` will be making use of `PayPal`'s [API idempotency capabilities](https://developer.paypal.com/docs/platforms/develop/idempotency/). So it'd be safe to retry transient failures, at least for a configured number of times

### Creating a PayPal Withdrawal, happy path

1. `PayPalWithdrawals Service` creates an internal DB entry to track the state of the withdrawal.
2. `PayPalWithdrawals Service` calls `Wallets Service` API to reduce the `Wallet`'s balance. `BalanceChange.Reference` field is used for bookkeeping
3. `PayPalWithdrawals Service` updates its internal state, signaling that the balance was successfully deducted from the Wallet. `BalanceChange.ID` is stored internally
3. `PayPalWithdrawals Service` calls `PayPal` to schedule a Withdrawal.
4. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal was completed

### Creating a PayPal Withdrawal, fails because the user doesn't have enough funds

1. `PayPalWithdrawals Service` creates an internal DB entry to track the state of the withdrawal.
2. `PayPalWithdrawals Service` calls `Wallets Service` API to reduce the `Wallet`'s balance. Which **fails because the user doesn't have enough balance**
3. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal was impossible.

### Creating a Paypal Withdrawal, PayPal temporary fails for a **transient cause**

1. `PayPalWithdrawals Service` creates an internal DB entry to track the state of the withdrawal.
2. `PayPalWithdrawals Service` calls `Wallets Service` API to reduce the `Wallet`'s balance. `BalanceChange.Reference` field is used for bookkeeping
3. `PayPalWithdrawals Service` updates its internal state, signaling that the balance was successfully deducted from the Wallet. `BalanceChange.ID` is stored internally
3. `PayPalWithdrawals Service` calls `PayPal` to schedule a Withdrawal. Which fails for a **transient cause**
4. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal failed once
5. `PayPalWithdrawals Service` retries the call to `PayPal` API a configurable number of times, until it succeeds.
6. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal was completed

### Creating a Paypal Withdrawal, PayPal fails for a **transient cause** for longer than we're confortable with

1. `PayPalWithdrawals Service` creates an internal DB entry to track the state of the withdrawal.
2. `PayPalWithdrawals Service` calls `Wallets Service` API to reduce the `Wallet`'s balance. `BalanceChange.Reference` field is used for bookkeeping
3. `PayPalWithdrawals Service` updates its internal state, signaling that the balance was successfully deducted from the Wallet. `BalanceChange.ID` is stored internally
3. `PayPalWithdrawals Service` calls `PayPal` to schedule a Withdrawal.  Which fails for a **transient cause**
4. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal failed once
5. `PayPalWithdrawals Service` retries the call to `PayPal` API a configured number of times, but it keeps failing for a **transient cause**
6. `PayPalWithdrawals Service` calls `Wallets Service` to increase the `Wallet`'s balance (refund). `BalanceChange.Reference` field is used for bookkeeping
7. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal was impossible. `BalanceChange.ID` is stored internally

### Creating a Paypal Withdrawal, PayPal fails for a **final cause**

1. `PayPalWithdrawals Service` creates an internal DB entry to track the state of the withdrawal.
2. `PayPalWithdrawals Service` calls `Wallets Service` API to reduce the `Wallet`'s balance. `BalanceChange.Reference` field is used for bookkeeping
3. `PayPalWithdrawals Service` updates its internal state, signaling that the balance was successfully deducted from the Wallet. `BalanceChange.ID` is stored internally
3. `PayPalWithdrawals Service` calls `PayPal` to schedule a Withdrawal. Which fails for a **final cause**
4. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal failed and it makes no sense to retry
5. `PayPalWithdrawals Service` calls `Wallets Service` to increase the `Wallet`'s balance (refund). `BalanceChange.Reference` field is used for bookkeeping
6. `PayPalWithdrawals Service` updates its internal state, signaling that the Withdrawal was impossible. `BalanceChange.ID` is stored internally


## Running this project

The system is delivered as a `docker-compose.yaml` that spins up a `postgres` instance, and an instance of our `Wallets Service`.

Start up with

```
docker-compose up
```

The `Wallets Service` instance will run schema migrations at startup, if necessary. It makes use of [golang-migrate](https://github.com/golang-migrate/migrate) for that purpose

By default, the service is configured to listen on the host's port `9000`. The following API call should return HTTP 404 if everything is in place (because wallet with ID 20 doesn't exist when we first spin up the service):

```
curl -i -X http://localhost:9000/wallets/20
```

## Testing this project

Some unittests are in place. They don't make use of DB. To run them, simply run:

```
go test -v
```

## Corners cut

To avoid spending a ton of time on this challenge, some corners were cut:

* Proper logging, specially around unexpected exceptions
* Properly isolating some exceptions at the right layer: some DB-specific exceptions are only handled on the service layer, instead of being handled on the storage layer
* Second pass or careful review of the `.sql` schema migration files. So some types or constraints might not be ideal
* Second pass or careful review of the interfaces used on the storage layer. I feel they could be polished
* Extending unittests: we're missing some unittests around the less-important endpoints and services. Also on the store layer.
* Idempotency: Although not explicitly required, I think an idempotency mechanism was important. It would have taken some time though
* Second pass or review of Echo's idioms: some of the code might not be Echo-idiomatic