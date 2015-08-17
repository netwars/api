Netwars API [![Build Status](https://travis-ci.org/netwars/api.svg)](https://travis-ci.org/netwars/api)
=============

Prosta aplikacja udostępniająca aktualną zawartość forum http://netwars.pl w postaci API. Po pobraniu i przeparsowaniu dane są zachowywane w pamięci. 
Jeżeli w przeciągu 24 godzin nie zostanie wysłane żadne żądanie o dany `Topic`, zostaje on wymazany.
Ponadto każdy `Topic` jest odświeżany co 30 sekund aż do momentu wygaśnięcia.

Instalacja
------------
1. Należy skonfigurować środowisko wg tego dokumentu (http://golang.org/doc/code.html#GOPATH)
2. `go get github.com/netwars/api`
3. `go get` aby pobrać zależności

Komendy
--------

#### Kompilacja
```bash
go build
```

#### Uruchamianie
Po skompilowaniu możemy uruchomić aplikacje. Nie wymaga ona żadnych dodatkowych zależności, takich jak np baza danych.
Opcjonalnie możemy podać flagę `-warmup`. Definiuje ona ile stron tematów ma zostać pobranych zaraz po uruchomieniu.

```bash
./api -warmup=0
```

API
---------
* lista forów: `GET:/forums`
* temat wraz z postami: `GET:/topic/<id>`
* list tematów posortowanych wg daty: `GET:/topics?offset=0&limit=10`
