// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Vladislav B",
            "email": "go-store@stick.sh"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/auth/login": {
            "post": {
                "description": "Auth a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Auth user",
                "parameters": [
                    {
                        "description": "Register account",
                        "name": "register",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/AuthRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-AuthResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/auth/register": {
            "post": {
                "description": "Register a new user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Register user",
                "parameters": [
                    {
                        "description": "Register account",
                        "name": "register",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-RegisterUserResponse"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/category/": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Get products",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "Get products",
                "parameters": [
                    {
                        "type": "integer",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_products_FindRow"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            },
            "post": {
                "description": "Create category",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "Category",
                "parameters": [
                    {
                        "description": "Create category",
                        "name": "create",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/CreateCategoryRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_category_response_CategoryResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/category/:id": {
            "put": {
                "description": "Update category",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "Category",
                "parameters": [
                    {
                        "description": "Update category",
                        "name": "update",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/UpdateCategoryRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_category_response_CategoryResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/category/:slug/": {
            "get": {
                "description": "Get category by slug",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Category",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Category Slug",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_category_response_CategoryResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/category/id/:id/": {
            "get": {
                "description": "Get category by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Category",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_category_response_CategoryResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/product/:slug/": {
            "get": {
                "description": "Get product by slug",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Product",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Product Slug",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_product_response_ProductResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        },
        "/v1/product/id/:id/": {
            "get": {
                "description": "Get product by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Product",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_product_response_ProductResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/APIErrors"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "APIError": {
            "type": "object",
            "properties": {
                "field": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "APIErrors": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "errors": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/APIError"
                    }
                }
            }
        },
        "AuthRequest": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                }
            }
        },
        "AuthResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "CreateCategoryRequest": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string",
                    "minLength": 1
                },
                "is_enabled": {
                    "type": "boolean"
                },
                "meta_description": {
                    "type": "string",
                    "minLength": 1
                },
                "meta_h1": {
                    "type": "string",
                    "minLength": 1
                },
                "meta_keyword": {
                    "type": "string",
                    "minLength": 1
                },
                "meta_title": {
                    "type": "string",
                    "minLength": 1
                },
                "name": {
                    "type": "string",
                    "maxLength": 255,
                    "minLength": 1
                },
                "parent_id": {
                    "type": "string"
                },
                "slug": {
                    "type": "string",
                    "maxLength": 255,
                    "minLength": 1
                }
            }
        },
        "FullPagingData": {
            "type": "object",
            "properties": {
                "last_page": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "page_size": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "JSONResponse-AuthResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "$ref": "#/definitions/AuthResponse"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "JSONResponse-RegisterUserResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "$ref": "#/definitions/RegisterUserResponse"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "JSONResponse-ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_categories_FindRow": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "$ref": "#/definitions/ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_categories_FindRow"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "JSONResponse-ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_products_FindRow": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "$ref": "#/definitions/ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_products_FindRow"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_category_response_CategoryResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "$ref": "#/definitions/github_com_stickpro_go-store_internal_delivery_http_response_category_response.CategoryResponse"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "JSONResponse-github_com_stickpro_go-store_internal_delivery_http_response_product_response_ProductResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "$ref": "#/definitions/github_com_stickpro_go-store_internal_delivery_http_response_product_response.ProductResponse"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "RegisterRequest": {
            "type": "object",
            "required": [
                "email",
                "language",
                "location",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "language": {
                    "type": "string",
                    "maxLength": 2,
                    "minLength": 2
                },
                "location": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                }
            }
        },
        "RegisterUserResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_categories_FindRow": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_stickpro_go-store_internal_storage_repository_repository_categories.FindRow"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/FullPagingData"
                }
            }
        },
        "ResponseWithFullPagination-github_com_stickpro_go-store_internal_storage_repository_repository_products_FindRow": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_stickpro_go-store_internal_storage_repository_repository_products.FindRow"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/FullPagingData"
                }
            }
        },
        "UpdateCategoryRequest": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string",
                    "minLength": 1
                },
                "is_enabled": {
                    "type": "boolean"
                },
                "meta_description": {
                    "type": "string",
                    "minLength": 1
                },
                "meta_h1": {
                    "type": "string",
                    "minLength": 1
                },
                "meta_keyword": {
                    "type": "string",
                    "minLength": 1
                },
                "meta_title": {
                    "type": "string",
                    "minLength": 1
                },
                "name": {
                    "type": "string",
                    "maxLength": 255,
                    "minLength": 1
                },
                "parent_id": {
                    "type": "string"
                },
                "slug": {
                    "type": "string",
                    "maxLength": 255,
                    "minLength": 1
                }
            }
        },
        "github_com_stickpro_go-store_internal_delivery_http_response_category_response.CategoryResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "is_enabled": {
                    "type": "boolean"
                },
                "meta_description": {
                    "type": "string"
                },
                "meta_h1": {
                    "type": "string"
                },
                "meta_keywords": {
                    "type": "string"
                },
                "meta_title": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "slug": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "github_com_stickpro_go-store_internal_delivery_http_response_product_response.ProductResponse": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "ean": {
                    "type": "string"
                },
                "height": {
                    "type": "number"
                },
                "id": {
                    "type": "string"
                },
                "image": {
                    "type": "string"
                },
                "is_enable": {
                    "type": "boolean"
                },
                "isbn": {
                    "type": "string"
                },
                "jan": {
                    "type": "string"
                },
                "length": {
                    "type": "number"
                },
                "location": {
                    "type": "string"
                },
                "manufacturer_id": {
                    "$ref": "#/definitions/uuid.NullUUID"
                },
                "meta_description": {
                    "type": "string"
                },
                "meta_h1": {
                    "type": "string"
                },
                "meta_keyword": {
                    "type": "string"
                },
                "meta_title": {
                    "type": "string"
                },
                "minimum": {
                    "type": "integer"
                },
                "model": {
                    "type": "string"
                },
                "mpn": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "price": {
                    "type": "number"
                },
                "quantity": {
                    "type": "integer"
                },
                "sku": {
                    "type": "string"
                },
                "slug": {
                    "type": "string"
                },
                "sort_order": {
                    "type": "integer"
                },
                "stock_status": {
                    "type": "string"
                },
                "subtract": {
                    "type": "boolean"
                },
                "upc": {
                    "type": "string"
                },
                "weight": {
                    "type": "number"
                },
                "width": {
                    "type": "number"
                }
            }
        },
        "github_com_stickpro_go-store_internal_storage_repository_repository_categories.FindRow": {
            "type": "object",
            "properties": {
                "created_at": {
                    "$ref": "#/definitions/pgtype.Timestamp"
                },
                "description": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "id": {
                    "type": "string"
                },
                "is_enable": {
                    "type": "boolean"
                },
                "meta_description": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "meta_h1": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "meta_keyword": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "meta_title": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "name": {
                    "type": "string"
                },
                "parent_id": {
                    "$ref": "#/definitions/uuid.NullUUID"
                },
                "slug": {
                    "type": "string"
                },
                "updated_at": {
                    "$ref": "#/definitions/pgtype.Timestamp"
                }
            }
        },
        "github_com_stickpro_go-store_internal_storage_repository_repository_products.FindRow": {
            "type": "object",
            "properties": {
                "created_at": {
                    "$ref": "#/definitions/pgtype.Timestamp"
                },
                "description": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "ean": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "height": {
                    "type": "number"
                },
                "id": {
                    "type": "string"
                },
                "image": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "is_enable": {
                    "type": "boolean"
                },
                "isbn": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "jan": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "length": {
                    "type": "number"
                },
                "location": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "manufacturer_id": {
                    "$ref": "#/definitions/uuid.NullUUID"
                },
                "meta_description": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "meta_h1": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "meta_keyword": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "meta_title": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "minimum": {
                    "type": "integer"
                },
                "model": {
                    "type": "string"
                },
                "mpn": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "name": {
                    "type": "string"
                },
                "price": {
                    "type": "number"
                },
                "quantity": {
                    "type": "integer"
                },
                "sku": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "slug": {
                    "type": "string"
                },
                "sort_order": {
                    "type": "integer"
                },
                "stock_status": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "subtract": {
                    "type": "boolean"
                },
                "upc": {
                    "$ref": "#/definitions/pgtype.Text"
                },
                "updated_at": {
                    "$ref": "#/definitions/pgtype.Timestamp"
                },
                "viewed": {
                    "type": "integer"
                },
                "weight": {
                    "type": "number"
                },
                "width": {
                    "type": "number"
                }
            }
        },
        "pgtype.InfinityModifier": {
            "type": "integer",
            "enum": [
                1,
                0,
                -1
            ],
            "x-enum-varnames": [
                "Infinity",
                "Finite",
                "NegativeInfinity"
            ]
        },
        "pgtype.Text": {
            "type": "object",
            "properties": {
                "string": {
                    "type": "string"
                },
                "valid": {
                    "type": "boolean"
                }
            }
        },
        "pgtype.Timestamp": {
            "type": "object",
            "properties": {
                "infinityModifier": {
                    "$ref": "#/definitions/pgtype.InfinityModifier"
                },
                "time": {
                    "description": "Time zone will be ignored when encoding to PostgreSQL.",
                    "type": "string"
                },
                "valid": {
                    "type": "boolean"
                }
            }
        },
        "uuid.NullUUID": {
            "type": "object",
            "properties": {
                "uuid": {
                    "type": "string"
                },
                "valid": {
                    "description": "Valid is true if UUID is not NULL",
                    "type": "boolean"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "GO-store",
	Description:      "This is an API for go-store",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
