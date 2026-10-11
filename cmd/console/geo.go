package console

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stickpro/go-store/internal/tools/garimport"
	"github.com/stickpro/go-store/pkg/database"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/urfave/cli/v3"
)

// hflabsColumns are the city.csv columns of github.com/hflabs/city, in file order.
var hflabsColumns = []string{
	"address", "postal_code", "country", "federal_district", "region_type", "region",
	"area_type", "area", "city_type", "city", "settlement_type", "settlement",
	"kladr_id", "fias_id", "fias_level", "capital_marker", "okato", "oktmo",
	"tax_office", "timezone", "geo_lat", "geo_lon", "population", "foundation_year",
}

// hflabsUpsertSQL inserts cities GAR doesn't know yet and, for existing rows,
// only fills in what GAR lacks (coordinates, population, timezone, foundation
// year, missing codes), so names stay in the GAR spelling whatever the order
// of imports.
const hflabsUpsertSQL = `
INSERT INTO cities (address, postal_code, country, federal_district, region_type, region,
                    area_type, area, city_type, city, settlement_type, settlement, kladr_id, fias_id, fias_level,
                    capital_marker, okato, oktmo, tax_office, timezone, geo_lat, geo_lon, population, foundation_year)
SELECT address, postal_code, country, federal_district, region_type, region,
       area_type, area, city_type, city, settlement_type, settlement, kladr_id, fias_id, fias_level,
       capital_marker, okato, oktmo, tax_office, timezone, geo_lat, geo_lon, population, foundation_year
FROM cities_hflabs_import
ON CONFLICT (fias_id) DO UPDATE SET
    timezone        = excluded.timezone,
    geo_lat         = excluded.geo_lat,
    geo_lon         = excluded.geo_lon,
    population      = excluded.population,
    foundation_year = excluded.foundation_year,
    capital_marker  = greatest(cities.capital_marker, excluded.capital_marker),
    postal_code     = coalesce(cities.postal_code, excluded.postal_code),
    kladr_id        = coalesce(cities.kladr_id, excluded.kladr_id),
    okato           = coalesce(cities.okato, excluded.okato),
    oktmo           = coalesce(cities.oktmo, excluded.oktmo),
    tax_office      = coalesce(cities.tax_office, excluded.tax_office)`

func prepareGeoCommands(appName, currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "import-city",
			Description: "import cities (with coordinates and population) from hflabs city.csv; upserts by fias_id",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "file",
					Aliases: []string{"f"},
					Usage:   "path to csv file with city data",
				},
			},
			Action: func(ctx context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))

				l := logger.NewExtended(loggerOpts...)
				dbClient, err := database.NewPostgresClient(ctx, conf.Postgres.DSN(), conf.Postgres.MinOpenConns, conf.Postgres.MaxOpenConns)
				if err != nil {
					return fmt.Errorf("failed to create database client: %w", err)
				}
				defer func() { _ = dbClient.Close() }()

				rows, err := readHflabsCSV(cl.String("file"))
				if err != nil {
					return err
				}
				if err := upsertHflabs(ctx, dbClient.DB, rows); err != nil {
					return err
				}

				l.Infow("cities imported", "rows", len(rows))
				return nil
			},
		},
		{
			Name: "import-gar",
			Description: "import every Russian settlement from the ФНС ГАР БД ФИАС export (gar_xml.zip) into cities; " +
				"by default streams only the needed files (~3 GB) of the latest export straight from fias-file.nalog.ru",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "source",
					Aliases: []string{"s"},
					Usage:   "gar_xml.zip path or URL (default: latest full export from fias.nalog.ru)",
				},
				&cli.StringSliceFlag{
					Name:    "region",
					Aliases: []string{"r"},
					Usage:   "region folder code(s) to import, e.g. -r 77 -r 50 (default: all)",
				},
				&cli.IntFlag{Name: "workers", Value: 4, Usage: "regions read concurrently"},
				&cli.BoolFlag{Name: dryRunFlag, Usage: "parse and report without writing to the database"},
			},
			Action: func(ctx context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))
				l := logger.NewExtended(loggerOpts...)
				defer func() { _ = l.Sync() }()

				dbClient, err := database.NewPostgresClient(ctx, conf.Postgres.DSN(), conf.Postgres.MinOpenConns, conf.Postgres.MaxOpenConns)
				if err != nil {
					return fmt.Errorf("failed to create database client: %w", err)
				}
				defer func() { _ = dbClient.Close() }()

				report, err := garimport.Run(ctx, dbClient.DB, l, garimport.Options{
					Source:  cl.String("source"),
					Regions: cl.StringSlice("region"),
					Workers: cl.Int("workers"),
					DryRun:  cl.Bool(dryRunFlag),
				})
				if report != nil {
					fmt.Printf("source: %s\n", report.Source)
					fmt.Printf("regions: %d, settlements: %d (without hierarchy: %d)\n",
						report.Regions, report.Settlements, report.NoHierarchy)
					fmt.Printf("inserted: %d, updated: %d\n", report.Inserted, report.Updated)
					if len(report.Skipped) > 0 {
						fmt.Printf("skipped regions: %s\n", strings.Join(report.Skipped, ", "))
					}
				}
				if err != nil {
					return fmt.Errorf("import gar: %w", err)
				}
				fmt.Println("run `search reindex --cities` to rebuild the city search index")
				return nil
			},
		},
	}
}

// readHflabsCSV reads city.csv; empty cells become NULLs.
func readHflabsCSV(path string) ([][]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	reader.Comment = '#'
	reader.FieldsPerRecord = len(hflabsColumns)

	var rows [][]any
	for header := true; ; header = false {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		if header {
			continue
		}
		row := make([]any, len(record))
		for i, v := range record {
			if v != "" {
				row[i] = v
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func upsertHflabs(ctx context.Context, pool *pgxpool.Pool, rows [][]any) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE cities_hflabs_import (LIKE cities INCLUDING DEFAULTS) ON COMMIT DROP`); err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"cities_hflabs_import"}, hflabsColumns, pgx.CopyFromRows(rows)); err != nil {
		return fmt.Errorf("copy cities: %w", err)
	}
	if _, err := tx.Exec(ctx, hflabsUpsertSQL); err != nil {
		return fmt.Errorf("upsert cities: %w", err)
	}
	return tx.Commit(ctx)
}
