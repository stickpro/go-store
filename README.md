## GO store

# Init project

Download geodb
```shell 
 make download-geodb
```
Download geo cities
```shell
 make download-geo-cities
 ```
Build a project
```shell
 make build
 ```
Run migrations
```shell
 ./.bin/go-store migrate up
```
Import cities
```shell
./.bin/go-store geo  import-city -f ./storage/geo/city.csv
 ```