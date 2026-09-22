// Package opencartimport reads categories and products out of a legacy OpenCart
// (3.x/4.x) MySQL database and reports them in a shape the go-store service
// layer can consume directly, for a one-off catalog migration.
package opencartimport

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql" //nolint:revive // registers the "mysql" database/sql driver
	"github.com/shopspring/decimal"
)

// Config holds the connection details for the source OpenCart database.
type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	TablePrefix string
	StoreID     int64
	LanguageID  int64
}

// Client reads OpenCart rows over database/sql.
type Client struct {
	db     *sql.DB
	prefix string
	storeID,
	languageID int64
	seoSchema seoURLSchema
}

// seoURLSchema tracks which oc_seo_url layout this OpenCart install uses -
// it changed mid-3.x from a single "query" string column (e.g. "product_id=42")
// to separate "key"/"value" columns.
type seoURLSchema int

const (
	seoURLSchemaUnknown seoURLSchema = iota
	seoURLSchemaKeyValue
	seoURLSchemaQueryString
	seoURLSchemaNone
)

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	// Sized generously so concurrent import workers each get their own
	// connection instead of queueing behind Go's tiny (2-idle-conn) default.
	db.SetMaxOpenConns(32)
	db.SetMaxIdleConns(32)

	// TablePrefix is taken as-is (including empty, for installs with no
	// prefix) - the "oc_" default lives on the --oc-prefix CLI flag, not here.
	return &Client{db: db, prefix: cfg.TablePrefix, storeID: cfg.StoreID, languageID: cfg.LanguageID}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) table(name string) string {
	return c.prefix + name
}

// CategoryRow is one OpenCart category joined with its description.
type CategoryRow struct {
	ID              int64
	ParentID        int64
	Name            string
	Image           string
	Status          bool
	SortOrder       int32
	MetaTitle       string
	MetaH1          string
	MetaDescription string
	MetaKeyword     string
}

func (c *Client) FetchCategories(ctx context.Context) ([]CategoryRow, error) {
	query := fmt.Sprintf(`
		SELECT c.category_id, c.parent_id, c.image, c.status, c.sort_order,
		       cd.name, cd.seo_title, cd.seo_h1, cd.meta_description, cd.meta_keyword
		FROM %s c
		JOIN %s cd ON cd.category_id = c.category_id AND cd.language_id = ?
		ORDER BY c.parent_id, c.sort_order`,
		c.table("category"), c.table("category_description"))

	rows, err := c.db.QueryContext(ctx, query, c.languageID)
	if err != nil {
		return nil, fmt.Errorf("fetch categories: %w", err)
	}
	defer rows.Close()

	var result []CategoryRow
	for rows.Next() {
		var (
			r                                                     CategoryRow
			image, name, metaTitle, metaH1, metaDesc, metaKeyword sql.NullString
		)
		if err := rows.Scan(&r.ID, &r.ParentID, &image, &r.Status, &r.SortOrder,
			&name, &metaTitle, &metaH1, &metaDesc, &metaKeyword); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		r.Image = nullStr(image)
		r.Name = nullStr(name)
		r.MetaTitle = nullStr(metaTitle)
		r.MetaH1 = nullStr(metaH1)
		r.MetaDescription = nullStr(metaDesc)
		r.MetaKeyword = nullStr(metaKeyword)
		result = append(result, r)
	}
	return result, rows.Err()
}

// ManufacturerRow is one OpenCart manufacturer (brand).
type ManufacturerRow struct {
	ID    int64
	Name  string
	Image string
}

func (c *Client) FetchManufacturers(ctx context.Context) ([]ManufacturerRow, error) {
	query := fmt.Sprintf(`SELECT manufacturer_id, name, image FROM %s ORDER BY manufacturer_id`, c.table("manufacturer"))
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fetch manufacturers: %w", err)
	}
	defer rows.Close()

	var result []ManufacturerRow
	for rows.Next() {
		var (
			r           ManufacturerRow
			name, image sql.NullString
		)
		if err := rows.Scan(&r.ID, &name, &image); err != nil {
			return nil, fmt.Errorf("scan manufacturer row: %w", err)
		}
		r.Name = nullStr(name)
		r.Image = nullStr(image)
		result = append(result, r)
	}
	return result, rows.Err()
}

// ProductRow is one OpenCart product joined with its description.
type ProductRow struct {
	ID              int64
	ManufacturerID  int64
	Model           string
	Sku             string
	Image           string
	Status          bool
	Quantity        int64
	SortOrder       int32
	Price           decimal.Decimal
	Weight          decimal.Decimal
	Length          decimal.Decimal
	Width           decimal.Decimal
	Height          decimal.Decimal
	Name            string
	Description     string
	MetaTitle       string
	MetaH1          string
	MetaDescription string
	MetaKeyword     string
}

// productModelBatchSize bounds how many `model` values go into one IN (...)
// list for FetchProductsByModels - keeps each query well under MySQL's
// max_allowed_packet / placeholder limits even for a huge go-store catalog.
const productModelBatchSize = 500

// FetchProductsByModels reads only the OpenCart products whose `model` is one
// of the given values, batched. This is what the variant import actually
// needs: go-store's existing catalog (matched by sku == model) is normally
// far smaller than OpenCart's full legacy catalog, so driving the fetch from
// go-store's side - instead of pulling every row of a potentially huge
// `product`/`product_description` (large TEXT columns) - is the difference
// between a few targeted queries and minutes of scanning+transferring rows
// nothing will ever match. onProgress, if non-nil, is called every
// progressEvery rows scanned.
func (c *Client) FetchProductsByModels(ctx context.Context, models []string, onProgress func(int)) ([]ProductRow, error) {
	var result []ProductRow
	n := 0

	for i := 0; i < len(models); i += productModelBatchSize {
		end := min(i+productModelBatchSize, len(models))
		batch := models[i:end]

		placeholders := make([]string, len(batch))
		args := make([]any, 0, len(batch)+1)
		args = append(args, c.languageID)
		for j, m := range batch {
			placeholders[j] = "?"
			args = append(args, m)
		}

		query := fmt.Sprintf(`
			SELECT p.product_id, p.manufacturer_id, p.model, p.sku, p.image, p.status, p.quantity, p.sort_order,
			       p.price, p.weight, p.length, p.width, p.height,
			       pd.name, pd.description, pd.seo_title, pd.seo_h1, pd.meta_description, pd.meta_keyword
			FROM %s p
			JOIN %s pd ON pd.product_id = p.product_id AND pd.language_id = ?
			WHERE p.model IN (%s)`,
			c.table("product"), c.table("product_description"), strings.Join(placeholders, ","))

		batchResult, err := func() ([]ProductRow, error) {
			rows, err := c.db.QueryContext(ctx, query, args...)
			if err != nil {
				return nil, fmt.Errorf("fetch products by model: %w", err)
			}
			defer rows.Close()

			var out []ProductRow
			for rows.Next() {
				r, err := scanProductRow(rows)
				if err != nil {
					return nil, err
				}
				out = append(out, r)
				if n++; onProgress != nil && n%progressEvery == 0 {
					onProgress(n)
				}
			}
			return out, rows.Err()
		}()
		if err != nil {
			return nil, err
		}
		result = append(result, batchResult...)
	}

	return result, nil
}

func scanProductRow(rows *sql.Rows) (ProductRow, error) {
	var (
		r                                                                              ProductRow
		price, weight, length, width, height                                           sql.NullString
		model, sku, image, name, description, metaTitle, metaH1, metaDesc, metaKeyword sql.NullString
	)
	if err := rows.Scan(&r.ID, &r.ManufacturerID, &model, &sku, &image, &r.Status, &r.Quantity, &r.SortOrder,
		&price, &weight, &length, &width, &height,
		&name, &description, &metaTitle, &metaH1, &metaDesc, &metaKeyword); err != nil {
		return ProductRow{}, fmt.Errorf("scan product row: %w", err)
	}
	r.Model = nullStr(model)
	r.Sku = nullStr(sku)
	r.Image = nullStr(image)
	r.Name = nullStr(name)
	r.Description = nullStr(description)
	r.MetaTitle = nullStr(metaTitle)
	r.MetaH1 = nullStr(metaH1)
	r.MetaDescription = nullStr(metaDesc)
	r.MetaKeyword = nullStr(metaKeyword)
	r.Price = parseDecimal(nullStr(price))
	r.Weight = parseDecimal(nullStr(weight))
	r.Length = parseDecimal(nullStr(length))
	r.Width = parseDecimal(nullStr(width))
	r.Height = parseDecimal(nullStr(height))
	return r, nil
}

// progressEvery is how many scanned rows pass between onProgress calls in the
// FetchAll* bulk readers below - these can be single queries over hundreds of
// thousands of rows on a large catalog, so silence for the whole call looks
// like a hang without some feedback.
const progressEvery = 50000

// FetchAllProductToCategory reads the whole product_to_category table in one
// round trip (instead of one query per product, which is what makes a 40k+
// product catalog painfully slow over a remote MySQL connection). onProgress,
// if non-nil, is called every progressEvery rows scanned.
func (c *Client) FetchAllProductToCategory(ctx context.Context, onProgress func(int)) (map[int64][]int64, error) {
	query := fmt.Sprintf(`SELECT product_id, category_id FROM %s`, c.table("product_to_category"))
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fetch product_to_category: %w", err)
	}
	defer rows.Close()

	result := map[int64][]int64{}
	n := 0
	for rows.Next() {
		var productID, categoryID int64
		if err := rows.Scan(&productID, &categoryID); err != nil {
			return nil, fmt.Errorf("scan product_to_category row: %w", err)
		}
		result[productID] = append(result[productID], categoryID)
		if n++; onProgress != nil && n%progressEvery == 0 {
			onProgress(n)
		}
	}
	return result, rows.Err()
}

// AttributeGroupRow is one OpenCart attribute group.
type AttributeGroupRow struct {
	ID        int64
	Name      string
	SortOrder int32
}

func (c *Client) FetchAttributeGroups(ctx context.Context) ([]AttributeGroupRow, error) {
	query := fmt.Sprintf(`
		SELECT g.attribute_group_id, g.sort_order, gd.name
		FROM %s g
		JOIN %s gd ON gd.attribute_group_id = g.attribute_group_id AND gd.language_id = ?
		ORDER BY g.sort_order, g.attribute_group_id`,
		c.table("attribute_group"), c.table("attribute_group_description"))

	rows, err := c.db.QueryContext(ctx, query, c.languageID)
	if err != nil {
		return nil, fmt.Errorf("fetch attribute groups: %w", err)
	}
	defer rows.Close()

	var result []AttributeGroupRow
	for rows.Next() {
		var (
			r    AttributeGroupRow
			name sql.NullString
		)
		if err := rows.Scan(&r.ID, &r.SortOrder, &name); err != nil {
			return nil, fmt.Errorf("scan attribute group row: %w", err)
		}
		r.Name = nullStr(name)
		result = append(result, r)
	}
	return result, rows.Err()
}

// AttributeRow is one OpenCart attribute.
type AttributeRow struct {
	ID        int64
	GroupID   int64
	Name      string
	SortOrder int32
}

func (c *Client) FetchAttributes(ctx context.Context) ([]AttributeRow, error) {
	query := fmt.Sprintf(`
		SELECT a.attribute_id, a.attribute_group_id, a.sort_order, ad.name
		FROM %s a
		JOIN %s ad ON ad.attribute_id = a.attribute_id AND ad.language_id = ?
		ORDER BY a.sort_order, a.attribute_id`,
		c.table("attribute"), c.table("attribute_description"))

	rows, err := c.db.QueryContext(ctx, query, c.languageID)
	if err != nil {
		return nil, fmt.Errorf("fetch attributes: %w", err)
	}
	defer rows.Close()

	var result []AttributeRow
	for rows.Next() {
		var (
			r    AttributeRow
			name sql.NullString
		)
		if err := rows.Scan(&r.ID, &r.GroupID, &r.SortOrder, &name); err != nil {
			return nil, fmt.Errorf("scan attribute row: %w", err)
		}
		r.Name = nullStr(name)
		result = append(result, r)
	}
	return result, rows.Err()
}

// ProductAttributeRow is one free-text attribute value on an OpenCart product.
type ProductAttributeRow struct {
	AttributeID int64
	Text        string
}

// FetchAllProductAttributes reads product_attribute for every product in one
// round trip; see FetchAllProductToCategory for why.
func (c *Client) FetchAllProductAttributes(ctx context.Context, onProgress func(int)) (map[int64][]ProductAttributeRow, error) {
	query := fmt.Sprintf(`SELECT product_id, attribute_id, text FROM %s WHERE language_id = ?`, c.table("product_attribute"))
	rows, err := c.db.QueryContext(ctx, query, c.languageID)
	if err != nil {
		return nil, fmt.Errorf("fetch product_attribute: %w", err)
	}
	defer rows.Close()

	result := map[int64][]ProductAttributeRow{}
	n := 0
	for rows.Next() {
		var (
			productID int64
			r         ProductAttributeRow
			text      sql.NullString
		)
		if err := rows.Scan(&productID, &r.AttributeID, &text); err != nil {
			return nil, fmt.Errorf("scan product_attribute row: %w", err)
		}
		r.Text = strings.TrimSpace(nullStr(text))
		if r.Text != "" {
			result[productID] = append(result[productID], r)
		}
		if n++; onProgress != nil && n%progressEvery == 0 {
			onProgress(n)
		}
	}
	return result, rows.Err()
}

// FetchAllProductImages reads product_image for every product in one round
// trip; see FetchAllProductToCategory for why.
func (c *Client) FetchAllProductImages(ctx context.Context, onProgress func(int)) (map[int64][]string, error) {
	query := fmt.Sprintf(`SELECT product_id, image FROM %s ORDER BY product_id, sort_order`, c.table("product_image"))
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fetch product_image: %w", err)
	}
	defer rows.Close()

	result := map[int64][]string{}
	n := 0
	for rows.Next() {
		var (
			productID int64
			img       sql.NullString
		)
		if err := rows.Scan(&productID, &img); err != nil {
			return nil, fmt.Errorf("scan product_image row: %w", err)
		}
		if s := nullStr(img); s != "" {
			result[productID] = append(result[productID], s)
		}
		if n++; onProgress != nil && n%progressEvery == 0 {
			onProgress(n)
		}
	}
	return result, rows.Err()
}

// FetchSeoKeywords returns oldID -> sanitized SEO keyword for either "category_id"
// or "product_id" routes, from the OpenCart oc_seo_url table. It transparently
// handles both seo_url layouts (see seoURLSchema) and, since these keywords are
// only a "nicer slug" nice-to-have, returns an empty map instead of an error if
// the table can't be read at all - callers fall back to a slugified name.
func (c *Client) FetchSeoKeywords(ctx context.Context, key string) (map[int64]string, error) {
	schema, err := c.detectSeoURLSchema(ctx)
	if err != nil || schema == seoURLSchemaNone {
		return map[int64]string{}, nil //nolint:nilerr // best-effort: slugify(name) is the fallback
	}

	var query, prefix string
	switch schema {
	case seoURLSchemaKeyValue:
		query = fmt.Sprintf("SELECT value, keyword FROM %s WHERE store_id = ? AND language_id = ? AND `key` = ?", c.table("seo_url"))
	case seoURLSchemaQueryString:
		query = fmt.Sprintf("SELECT query, keyword FROM %s WHERE store_id = ? AND language_id = ? AND query LIKE ?", c.table("seo_url"))
		prefix = key + "=%"
	default:
		return map[int64]string{}, nil
	}

	var rows *sql.Rows
	if schema == seoURLSchemaQueryString {
		rows, err = c.db.QueryContext(ctx, query, c.storeID, c.languageID, prefix)
	} else {
		rows, err = c.db.QueryContext(ctx, query, c.storeID, c.languageID, key)
	}
	if err != nil {
		return map[int64]string{}, nil //nolint:nilerr // best-effort
	}
	defer rows.Close()

	result := make(map[int64]string)
	for rows.Next() {
		var value, keyword string
		if err := rows.Scan(&value, &keyword); err != nil {
			return nil, fmt.Errorf("scan seo_url row: %w", err)
		}
		// value/query looks like "product_id=42" (query-string schema) or just "42" (key/value schema).
		if idx := strings.LastIndex(value, "="); idx >= 0 {
			value = value[idx+1:]
		}
		// "path=5_10" style category chains - take the last (most specific) segment.
		if idx := strings.LastIndex(value, "_"); idx >= 0 {
			value = value[idx+1:]
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		result[id] = keyword
	}
	return result, rows.Err()
}

// detectSeoURLSchema inspects the oc_seo_url table's columns once and caches
// the result on the client.
func (c *Client) detectSeoURLSchema(ctx context.Context) (seoURLSchema, error) {
	if c.seoSchema != seoURLSchemaUnknown {
		return c.seoSchema, nil
	}

	rows, err := c.db.QueryContext(ctx, fmt.Sprintf("SHOW COLUMNS FROM %s", c.table("seo_url")))
	if err != nil {
		c.seoSchema = seoURLSchemaNone
		return c.seoSchema, fmt.Errorf("inspect seo_url columns: %w", err)
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var field, colType string
		var null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			c.seoSchema = seoURLSchemaNone
			return c.seoSchema, fmt.Errorf("scan seo_url column: %w", err)
		}
		cols[strings.ToLower(field)] = true
	}
	if err := rows.Err(); err != nil {
		c.seoSchema = seoURLSchemaNone
		return c.seoSchema, err
	}

	switch {
	case cols["key"] && cols["value"]:
		c.seoSchema = seoURLSchemaKeyValue
	case cols["query"]:
		c.seoSchema = seoURLSchemaQueryString
	default:
		c.seoSchema = seoURLSchemaNone
	}
	return c.seoSchema, nil
}

func nullStr(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func parseDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}
