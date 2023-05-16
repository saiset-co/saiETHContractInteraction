## saiETHContractInteraction

### Config
#### config.yml

eth_server: "" //For all contracts for now  
log_mode: "debug" //Debug mode

#### contracts.json

{  
&emsp;    "name": "", //Contract name, uses in api commands  
&emsp;    "server": "", //Feature update, geth server per contract  
&emsp;    "abi": "", //Contract ABI, escaped json string  
&emsp;    "address": "", //Contract address  
&emsp;    "private": "", //Private key to sign commands  
&emsp;    "gas_limit": 0 //Gas limit for the command transaction  
}

### API
#### Contract command
- request:

curl --location --request GET 'http://localhost:8804' \
&emsp;    --header 'Token: SomeToken' \
&emsp;    --header 'Content-Type: application/json' \
&emsp;    --data-raw '{"method": "api", "data": {"contract":"$name","method":"$contract_method_name", "value": "$value", "params":[{"type":"$(int|string|float...)","value":"$some_value"}]}}'

- response: {"tx_0123"} //transaction hash
- response (if geth server is not available) {"Data":"07970bb0-2ec5-482f-bd92-6366f378810a","StatusCode":200,"Headers":null}

#### Add contracts
- request:

curl --location --request GET 'http://localhost:8804' \
&emsp;    --header 'Token: SomeToken' \
&emsp;    --header 'Content-Type: application/json' \
&emsp;    --data-raw '{"method": "add", "data": {"contracts": [{"name":"$name", "server": "$server", "address":"$address","abi":"$abi", "private": "$private", "gas_limit":100}]}}'

- response: {"ok"}

#### Delete contracts
- request:

curl --location --request GET 'http://localhost:8804' \
&emsp;    --header 'Token: SomeToken' \
&emsp;    --header 'Content-Type: application/json' \
&emsp;    --data-raw '{"method": "delete", "data": {"names": ["$name"]}}'

- response: {"ok"}

#### Сheck status of pending request (request, in which geth server was not available)
- request:

curl --location --request GET 'http://localhost:8804' \
&emsp;    --header 'Token: SomeToken' \
&emsp;    --header 'Content-Type: application/json' \
&emsp;    --data-raw '{"method": "checkStatus", "data": {"id": "$id"}}'

- response: {"Data":false,"StatusCode":200}

### Run in Docker
`make up`

### Run as standalone application
`microservices/saiETHContractInteraction/build/sai-eth-interaction

## Profiling
 `host:port/debug/pprof`
