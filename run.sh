#!/bin/bash

go build -o app cmd/web/*.go
./app -dbname=postgres -dbuser=postgres -secret= -dbpassword=password -production=false -cache=false -smtpuser= -smtphost= -frontend=localhost:8080 -smtppass=
