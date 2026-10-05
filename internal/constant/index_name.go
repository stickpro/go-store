package constant

const (
	CitiesIndex          string = "cities"
	ProductVariantsIndex string = "product_variants"
	AttributesIndex      string = "attributes"
	AttributeGroupsIndex string = "attributes_groups"
)

// ProductVariantStaticFilterableAttributes are the ProductVariantsIndex
// filterable fields that come from the variant/product schema itself, not
// from the dynamic per-attribute slugs (color, size, …). Shared between
// product.Service (full index build) and attribute.Service (incremental
// settings push) so the two never drift apart.
var ProductVariantStaticFilterableAttributes = []string{
	"price", "category_id", "category_ids", "manufacturer_id", "is_enable", "stock_status", "model",
}
