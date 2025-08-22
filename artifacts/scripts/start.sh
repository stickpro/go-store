#!/bin/sh

./go-store migrate up --disable-confirmations

./go-store start