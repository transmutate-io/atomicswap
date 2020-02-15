#!/usr/bin/env bash

exec > cryptos.gen.go

header() {
  printf "package params\n\n"
}

constants() {
  iota=" Crypto = iota"
  printf "const (\n"
  for i in "$@"
    do printf "\t$i$iota\n"
    iota=""
  done
  printf ")\n\n"
}

constants-names() {
  printf "var _cryptoNames = map[Crypto]string{\n"
  for i in "$@"
    do printf "\t$i: \"$i\",\n"
  done
  printf "}\n\n"
}

constants-ticker() {
  printf "var _cryptoTickers = map[Crypto]string{\n"
  for i in "$@"
    do printf "\t$(echo $i|cut -d : -f 1): \"$(echo $i|cut -d : -f 2)\",\n"
  done
  printf "}\n\n"
}

cryptos="$(cat cryptos.txt)"
cryptos_names="$(cat cryptos.txt|cut -d : -f 1)"

header
constants $cryptos_names
constants-names $cryptos_names
constants-ticker $cryptos