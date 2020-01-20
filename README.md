# Sagas

## Potential test cases

- 1 Transaction
    - Success
    - Fail
    - Saga coordinator fails, new coordinator finishes job
        - Before successful RPC
        - After successful RPC
    - Entity service fails, coordinator can't get response
        - Before processing request
        - After processing request

- 2 Transactions (parallel)

- 2 Transactions (sequential)