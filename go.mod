module transmutate.io/pkg/atomicswap

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/gcash/bchd v0.15.2
	github.com/gcash/bchutil v0.0.0-20191012211144-98e73ec336ba
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/yaml.v2 v2.2.4
	transmutate.io/pkg/cryptocore v0.0.1
	transmutate.io/pkg/reflection v0.0.1
)

replace (
	transmutate.io/pkg/cryptocore => ../cryptocore
	transmutate.io/pkg/reflection => ../reflection
)
