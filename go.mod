module tez-ton-bridge-relay

go 1.17

require (
	blockwatch.cc/tzgo v0.11.2-0.20220216144900-9a2fcbe99e1b
	github.com/move-ton/ton-client-go v1.28.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/decred/dcrd/dcrec/secp256k1 v1.0.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v2 v2.0.0 // indirect
	github.com/echa/log v1.1.0 // indirect
	github.com/go-bson/bson v0.0.0-20171017145622-6d291e839eca // indirect
	github.com/kr/pretty v0.3.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
)

replace blockwatch.cc/tzgo => github.com/Big-Kotik/tzgo v0.11.2-0.20220220212758-1445eb27de54
