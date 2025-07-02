module github.com/kingparks/cursor-vip

go 1.23.0

toolchain go1.23.9

require (
	github.com/astaxie/beego v1.12.3
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/eiannone/keyboard v0.0.0-20220611211555-0d226195f203
	github.com/gofrs/flock v0.12.1
	github.com/kingparks/cursor-vip/auth v0.0.0-00010101000000-000000000000
	github.com/kingparks/cursor-vip/auth/sign v0.0.0-00010101000000-000000000000
	github.com/kingparks/cursor-vip/authtool v0.0.0-00010101000000-000000000000
	github.com/mattn/go-colorable v0.1.13
	github.com/tidwall/gjson v1.17.1
	github.com/unknwon/i18n v0.0.0-20210904045753-ff3a8617e361
	golang.org/x/sys v0.33.0
	howett.net/plist v1.0.1
)

require (
	github.com/lqqyt2423/go-mitmproxy v1.8.5 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	gopkg.in/ini.v1 v1.46.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace (
	github.com/denisbrodbeck/machineid => ./auth/machineid
	github.com/kingparks/cursor-vip/auth => ./auth
	github.com/kingparks/cursor-vip/auth/sign => ./auth/sign
	github.com/kingparks/cursor-vip/authtool => ./authtool
	github.com/lqqyt2423/go-mitmproxy => ./auth/go-mitmproxy
	github.com/ugorji/go => github.com/ugorji/go v1.2.12
	github.com/ugorji/go/codec => github.com/ugorji/go/codec v1.2.12
)
