#!/bin/bash

go build -o GoApp cmd/web/*.go && ./GoApp
./GoApp -dbname=bookings -dbuser=sekitakeru -cache=false -production=false 