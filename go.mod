module github.com/vipernet-xyz/viper-network

go 1.18

replace github.com/tendermint/tendermint => github.com/vipernet-xyz/tendermint v0.34.0

replace github.com/tendermint/tm-db => github.com/vipernet-xyz/tm-db v0.1.0

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/core v0.3.2
	cosmossdk.io/depinject v1.0.0-alpha.3
	cosmossdk.io/errors v1.0.0-beta.7
	cosmossdk.io/math v1.0.0-rc.0
	github.com/99designs/keyring v1.2.2
	github.com/armon/go-metrics v0.3.10
	github.com/bgentry/speakeasy v0.1.0
	github.com/cockroachdb/errors v1.9.1
	github.com/cometbft/cometbft v0.37.0
	github.com/cosmos/btcutil v1.0.5
	github.com/cosmos/cosmos-proto v1.0.0-beta.3
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/cosmos/gogoproto v1.4.4
	github.com/cosmos/ics23/go v0.9.1-0.20221207100636-b1abd8678aab
	github.com/cosmos/ledger-cosmos-go v0.12.1
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1
	github.com/go-kit/kit v0.12.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hdevalence/ed25519consensus v0.1.0
	github.com/jordanorelli/lexnum v0.0.0-20141216151731-460eeb125754
	github.com/julienschmidt/httprouter v1.3.0
	github.com/magiconair/properties v1.8.6
	github.com/mattn/go-isatty v0.0.16
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.37.0
	github.com/regen-network/cosmos-proto v0.3.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/stretchr/testify v1.8.2
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/tendermint/tm-db v0.5.1
	github.com/vipernet-xyz/utils-go v0.1.0
	github.com/willf/bloom v2.0.3+incompatible
	golang.org/x/crypto v0.5.0
	google.golang.org/genproto v0.0.0-20230202175211-008b39050e57
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8
	gopkg.in/h2non/gock.v1 v1.1.2
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible
	sigs.k8s.io/yaml v1.3.0
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/ChainSafe/go-schnorrkel v0.0.0-20200405005733-88cbf1b4c40d // indirect
	github.com/Workiva/go-datastructures v1.0.52 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/getsentry/sentry-go v0.12.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gojektech/heimdall v5.0.2+incompatible // indirect
	github.com/gojektech/valkyrie v0.0.0-20190210220504-8f62c1e7ba45 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/gtank/ristretto255 v0.1.2 // indirect
	github.com/h2non/parth v0.0.0-20190131123155-b4df798d6542 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	github.com/willf/bitset v1.1.10 // indirect
	github.com/zondax/hid v0.9.1 // indirect
	github.com/zondax/ledger-go v0.14.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
