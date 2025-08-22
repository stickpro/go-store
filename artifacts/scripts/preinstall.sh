#!/bin/sh

echo "Executing preinstall script"

if systemctl list-unit-files | grep "go-store.service"
 then
   systemctl stop go-store.service
fi

echo "Preinstall script done"