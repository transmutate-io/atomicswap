module transmutate.io/pkg/atomicswap

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/davecgh/go-spew v1.1.0
	github.com/golang/protobuf v1.3.3
	github.com/ltcsuite/ltcd v0.20.1-beta
	github.com/stretchr/testify v1.5.0
	golang.org/x/crypto v0.0.0-20200214034016-1d94cc7ab1c6
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	gopkg.in/yaml.v2 v2.2.2
	transmutate.io/pkg/btccore v0.0.0-20200225214617-043e78c86284
)

replace transmutate.io/pkg/btccore => ../btccore
