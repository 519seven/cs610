## Battleship

### Pre-requisites

**github.com/justinas/alice**

middleware - to chain HTTP events together

MIT license - https://github.com/justinas/alice/blob/master/LICENSE

**github.com/mattn/go-sqlite3**

driver for the sqlite3 datastore

MIT license - https://github.com/mattn/go-sqlite3/blob/master/LICENSE

**golang.org/x/xerrors**

a v1.13 'errors'-like module so you can future-proof your v1.10 and v1.11 code

Proprietary license - https://github.com/golang/xerrors/blob/master/LICENSE

## Installation/Deployment Steps Using `go get`

##### `go get` some libraries, including my project
`go get github.com/justinas/alice github.com/mattn/go-sqlite3 golang.org/x/xerrors github.com/519seven/cs610/battleship`

##### Ignore any errors about files missing.
###### The errors indicate there are no .go files in the top directory.  This is indeed true
```
Example:
user@student:~$ go get github.com/519seven/cs610/battleship
package github.com/519seven/cs610/battleship: no Go files in /cs/home/stu/akeypj/go/src/github.com/519seven/cs610/battleship
```
##### cd to my project's directory in your GOPATH
`cd $(go env GOPATH)/src/github.com/519seven/cs610/battleship`

##### run the go build command
`go build ./cmd/web`

##### view help dialog
```
 user@student:~/go/src/github.com/519seven/cs610/battleship$ ./web -h
Usage of ./web:
  -dsn string
        SQLite data source name (default "./battleship.db")
  -initialize
        Start with a fresh database
  -port string
        HTTP port on which to listen (default ":5033")
```
   
##### run web app with default settings (port 5033 is default)
`./web`

##### run web app with alternative port number
`./web -port=:5055`

Open your browser to http://<webserver_ip>:<port>


## Install by downloading the zip file

##### From your local machine where the zip was downloaded
`scp ~/Downloads/Battleship.zip <user>:<webserver>:`

##### On your web server
`mkdir -p $(go env GOPATH)/src/github.com/519seven/cs610`

`unzip -o -q Battleship.zip -d $(go env GOPATH)/src/github.com/519seven/cs610`

##### Get three git modules
`go get github.com/justinas/alice github.com/mattn/go-sqlite3 golang.org/x/xerrors`

##### cd into the battleship root directory and run
`cd $(go env GOPATH)/src/github.com/519seven/cs610/battleship`
`go build ./cmd/web`

##### run web app with default settings (default port number is 5033)
`./web`

##### run web app and specify a different port
`./web -port=:5055`

Open your browser to http://<webserver_ip>:<port>
