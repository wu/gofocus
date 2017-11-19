## Synopsis

This project provides an HTTP API for OmniFocus, a task management
tool for OS X and iOS.  HTTP requests are translated into local
applescript commands or queries to the local omnifocus sqlite
database.

## Examples

```
# create a new task
curl -v -F name='test name' -F due=2016-01-01 -F parent=mytasks http://hostname:8080/create
# new task id returned as a Location code

# query task by id, returned as JSON
curl -v http://hostname:8080/id/some_id_here

# query tasks by name, returned as JSON array
curl -v 'http://hostname:8080/query/test'

# query task by name with wildcard '%', url-encoded for curl
curl -v 'http://hostname:8080/query/test%25'

# mark task done by id
curl -v http://hostname:8080/done/some_id_here

# mark task not done by id
curl -v http://hostname:8080/done/some_id_here
```

## Motivation

Omnifocus is a fabulous tool.  I use it on OS X, IOS, and even on my
watch.

I've written some other tools to automate omnifocus, but they all
suffer from one significant limitation--they need to be run on a
machine where omnifocus is running.  Providing a web API for omnifocus
will make it possible for tools that run on linux or other systems.


## Installation

```
# build
go build gofocus.go

# start listening on port 8080
./gofocus

# use non-standard database file path
DBFILE="/path/to/OmniFocusDatabase2" ./gofocus
```


## License

Copyright (c) 2017, Alex White
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

The views and conclusions contained in the software and documentation are those
of the authors and should not be interpreted as representing official policies,
either expressed or implied, of the FreeBSD Project.
