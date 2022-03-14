# Tezos Everscale Bridge Relay server

## Setup testing environment 

- Running Tezos local node:

```shell
# Run node
docker run --rm --name my-sandbox --detach -p 20000:20000 tqtezos/flextesa:20210930 granabox start

# Clean config
tezos-client config reset

# Update entrypoint
tezos-client --endpoint http://localhost:20000 bootstrapped

# Get test account data
docker run --rm tqtezos/flextesa:20210930 granabox info

# Add new alias
tezos-client import secret key alice $key --force
```

- Running Everscale local node

```shell
# Firstly you need to install Everdev toolkit
npm i -g everdev

# Run node
everdev se start
```
