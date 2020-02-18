module transmutate.io/pkg/swapper

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/golang/protobuf v1.2.0
	github.com/ltcsuite/ltcd v0.20.1-beta
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20200115085410-6d4e4cb37c7d
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	gopkg.in/yaml.v2 v2.2.2
	transmutate.io/pkg/btccore v0.0.0-00010101000000-000000000000
)

replace transmutate.io/pkg/btccore => ../btccore
