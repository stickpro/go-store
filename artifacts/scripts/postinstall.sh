#!/bin/sh

echo "Executing postinstall script"

go_store="${GO_STORE}"

if [ -z $go_store ]; then
  go_store="gostore"
fi

id -u $go_store || useradd $go_store

usermod -aG $go_store gostore

setcap cap_net_bind_service+ep /etc/go-store/go-store

if [ -e /etc/go-store/go-store.service  ] && ! [ -e /etc/systemd/system/go-store.service ]
 then
   echo "Unit file not exists. Сopying..."
   cp  /etc/go-store/go-store.service /etc/systemd/system/go-store.service
fi

if [ ! -e /usr/bin/go-store ]; then
  echo "Creating symlink /usr/bin/go-store -> /etc/go-store/go-store.service"
  ln -s /etc/go-store/go-store.service /usr/bin/go-store
fi

systemctl enable go-store.service
systemctl restart go-store.service

echo "Postinstall scripts done"