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

### Run in Docker
`make up`

### Run as standalone application
`microservices/saiETHContractInteraction/build/sai-eth-interaction