package console

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/pkg/database"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/urfave/cli/v3"
	"io"
	"os"
)

func prepareGeoCommands(appName, currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "import-city",
			Description: "import city from csv file",
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

				dataCSV, err := os.ReadFile(cl.String("file"))
				if err != nil {
					l.Fatalln("couldn't open or read the file cienaa.csv", err)
				}

				reader := csv.NewReader(bytes.NewReader(dataCSV))
				reader.Comma = ','
				reader.Comment = '#'

				var dataRead [][]any
				for {
					record, err := reader.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						l.Fatalln("couldn't read the record", err)
					}

					var recordAny []any
					for _, v := range record {
						recordAny = append(recordAny, v)
					}

					dataRead = append(dataRead, recordAny)
				}
				dataRead = dataRead[1:]

				columns := []string{
					"address",
					"postal_code",
					"country",
					"federal_district",
					"region_type",
					"region",
					"area_type",
					"area",
					"city_type",
					"city",
					"settlement_type",
					"settlement",
					"kladr_id",
					"fias_id",
					"fias_level",
					"capital_marker",
					"okato",
					"oktmo",
					"tax_office",
					"timezone",
					"geo_lat",
					"geo_lon",
					"population",
					"foundation_year",
				}
				_, err = dbClient.DB.Query(ctx, "TRUNCATE TABLE cities")
				if err != nil {
					return err
				}
				copyCount, err := dbClient.DB.CopyFrom(
					context.Background(),
					pgx.Identifier{"cities"},
					columns,
					pgx.CopyFromSlice(len(dataRead), func(i int) ([]any, error) {
						row := make([]any, len(dataRead[i]))
						for j := range dataRead[i] {
							row[j] = dataRead[i][j]
						}
						return row, nil
					}),
				)
				if err != nil {
					return err
				}

				l.Info(copyCount)
				return nil
			},
		},
	}
}
