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
Import all settlements of Russia (~155k cities, towns and villages) from ГАР БД ФИАС.
By default only the needed files (~3 GB of the ~58 GB `gar_xml.zip`) of the latest
export are streamed straight from fias-file.nalog.ru; `--source` takes a local
`gar_xml.zip` or another URL, `-r 77 -r 50` limits the regions, `--dry-run` only parses.
```shell
./.bin/go-store geo import-gar
 ```
Add coordinates, population and timezone for ~1100 cities (GAR has none of these).
Both imports upsert by `fias_id`, so they can be re-run in any order.
```shell
./.bin/go-store geo import-city -f ./storage/geo/city.csv
 ```
Rebuild the city search index
```shell
./.bin/go-store search reindex --cities
 ```