#!/bin/bash

go build -o bookings cmd/web/*.go 
./bookings -dbname=bookings -dbuser=postgres -dbpass=dont4get -cache=false -production=false