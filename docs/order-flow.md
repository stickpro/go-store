# Флоу заказа — план реализации

Статус: черновик. Раздел «Оплата (v2)» — отдельным этапом после выбора провайдера.

## Контекст проекта

- Чистая архитектура: Delivery → Service → Repository → Storage.
- Корзина живёт в **Redis** (`internal/service/cart`), ключ по `user_id` или `session_id`,
  позиция = `{product_id, variant_id, quantity}`. Корзина эфемерна.
- Цена и остаток — на таблице `products`: `price_retail / price_business / price_wholesale`,
  `quantity`, `subtract` (списывать ли остаток), `minimum` (мин. кол-во в заказе), `stock_status`.
  `product_variants` — витринный SKU без собственной цены/остатка.
- Деньги — `shopspring/decimal`, никогда не float.
- Транзакции: `repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {...})`
  + `s.storage.X(repository.WithTx(tx))`.
- Генерация БД-слоя: миграции в `sql/postgres/migrations/`, запросы в
  `sql/postgres/queries/<table>/<table>.sql` (+ `<table>_gen.sql` для CRUD-конфига в `sql/pgxgen.yaml`),
  прогон `make gen-sql`. **Сгенерированные файлы не редактируем руками.**
- Инфраструктура: Kafka-консьюмеры, mail-очередь с ретраями и dead-letter, воркеры
  (`internal/messaging/worker`).
- Ответы — только через `internal/delivery/http/response/<domain>_response`, `response.OkByData`.
- Таблица `cities` уже есть — под адрес доставки.

## Ключевые принципы

1. **Заказ — персистентная сущность в PostgreSQL**, не в Redis.
2. **Полный снапшот в `order_items`**: цена, имя, sku, картинка на момент оформления.
   Историю заказа нельзя строить джойном к живому товару.
3. **Все суммы считаются на сервере.** Клиент может прислать `expected_total` только
   для сверки — при расхождении ответ `409` с актуальной корзиной.
4. **Идемпотентность checkout** — заголовок `Idempotency-Key`.
5. **Списание остатка — в той же транзакции**, что и создание заказа,
   с `SELECT ... FOR UPDATE` по товарам, сортировка по `id` (против дедлоков).
6. **Явный state machine** + таблица истории статусов.
7. **Побочные эффекты (письмо, Kafka) — после коммита транзакции**, ошибки только логируем.

## Схема БД

Новая миграция `sql/postgres/migrations/<ts>_create_orders.up.sql`:

```sql
create sequence orders_number_seq start 100000;

create table orders (
    id              uuid primary key default gen_random_uuid(),
    order_number    bigint       not null unique default nextval('orders_number_seq'),
    user_id         uuid         references users(id) on delete set null,  -- null = гость
    status          varchar(32)  not null default 'pending',
    payment_status  varchar(32)  not null default 'unpaid',
    payment_method  varchar(32),
    currency        char(3)      not null default 'RUB',

    email           varchar(255) not null,
    phone           varchar(32),

    ship_city_id    uuid references cities(id),
    ship_city_name  varchar(255) not null,
    ship_address    varchar(512) not null,
    ship_postcode   varchar(16),
    ship_recipient  varchar(255) not null,
    shipping_method varchar(64),

    subtotal        decimal(15,4) not null,
    discount_total  decimal(15,4) not null default 0,
    shipping_total  decimal(15,4) not null default 0,
    tax_total       decimal(15,4) not null default 0,
    grand_total     decimal(15,4) not null,

    comment         text,
    idempotency_key varchar(80),
    created_at      timestamp not null default current_timestamp,
    updated_at      timestamp null default current_timestamp,
    paid_at         timestamp,
    cancelled_at    timestamp
);
create unique index uq_orders_idem on orders (user_id, idempotency_key) where idempotency_key is not null;
create index idx_orders_user on orders (user_id, created_at desc);
create index idx_orders_open on orders (status) where status in ('pending','paid');

create table order_items (
    id          uuid primary key default gen_random_uuid(),
    order_id    uuid not null references orders(id) on delete cascade,
    product_id  uuid,                       -- nullable, для аналитики; НЕ FK с каскадом удаления
    variant_id  uuid,
    sku         varchar(255),
    name        varchar(512) not null,
    slug        varchar(255),
    image_path  varchar(255),
    unit_price  decimal(15,4) not null,
    quantity    bigint not null,
    line_total  decimal(15,4) not null
);
create index idx_order_items_order on order_items (order_id);

create table order_status_history (
    id          uuid primary key default gen_random_uuid(),
    order_id    uuid not null references orders(id) on delete cascade,
    from_status varchar(32),
    to_status   varchar(32) not null,
    actor       varchar(64) not null,      -- 'customer' | 'admin:<id>' | 'system' | 'payment:<provider>'
    comment     text,
    created_at  timestamp not null default current_timestamp
);
```

`order_items` **намеренно без FK** на `products` / `product_variants`: `product_id` и
`variant_id` — nullable `uuid` без ссылки, поэтому исторические заказы переживают удаление
из каталога (весь нужный для отображения набор — `name`, `sku`, `slug`, `image_path`,
`unit_price` — лежит в снапшоте строки). Полный soft-delete товара — отдельная задача,
для целостности заказов не требуется.

### Статусы

- `status` (фулфилмент): `pending → paid → processing → shipped → delivered`,
  плюс `cancelled`, `refunded`.
- `payment_status`: `unpaid / paid / refunded / failed`.
- Оплату и фулфилмент держим раздельно.

### State machine

```go
var allowed = map[string][]string{
    "pending":    {"paid", "cancelled"},
    "paid":       {"processing", "refunded", "cancelled"},
    "processing": {"shipped", "cancelled", "refunded"},
    "shipped":    {"delivered", "refunded"},
    "delivered":  {"refunded"},
}
```

`transition(ctx, tx, order, to, actor, comment)` — проверка перехода, `UPDATE orders`,
запись в `order_status_history`, возврат доменного события для публикации после коммита.
Отмена/возврат из `paid` и далее → restock остатка.

## Структура кода

```
internal/service/order/
  service.go     IOrderService, New(), DI
  checkout.go    CreateOrder — основная транзакция
  query.go       GetByNumber, ListForUser (пагинация через internal/storage/base.Paginate)
  status.go      transition() + Cancel/restock
  pricing.go     resolveUnitPrice(user, product) — retail/business/wholesale
  errors.go      ErrCartEmpty, ErrInsufficientStock, ErrPriceChanged, ErrInvalidTransition, ...
internal/dto/order.go
internal/messaging/contracts/order.go            OrderCreatedPayload, OrderPaidPayload, OrderCancelledPayload
internal/delivery/http/handlers/order.go
internal/delivery/http/request/order_request/
internal/delivery/http/response/order_response/
internal/storage/repository/repository_orders/   (сгенерированный)
sql/postgres/queries/orders/orders.sql  +  orders_gen.sql
internal/messaging/worker/order_expiry.go
```

Проводка:

- `internal/storage/repository/repository.go` — `IRepository` + поле структуры + `InitRepository`
  + accessor-метод `Orders(opts ...Option)` с веткой `WithTx`.
- `internal/service/service.go` — `Services.OrderService` + создание в `InitService`.
- `internal/delivery/http/handlers/handler.go` — `initOrderRoutes`:
  - `POST /api/v1/orders` — открытая группа, owner через `h.cartOwner(c)` (гость по session или юзер);
  - `GET /api/v1/orders`, `GET /api/v1/orders/:number` — в группе `secured` (только юзер).
- `sql/pgxgen.yaml` — блок `tables:` для `orders`, `order_items`, `order_status_history`.
- `internal/app/app.go` — запуск воркера `order_expiry`, подписка на `order.paid`.
- Общий `pricing.go` заодно закрывает `TODO: get price by user group` в `cart.enrichCart` —
  вынести туда и переиспользовать.

## Транзакция checkout

`POST /api/v1/orders` → `CreateOrder(ctx, owner dto.Owner, d dto.CreateOrderDTO)`:

```
1. d.IdempotencyKey задан и заказ найден → вернуть существующий (200)
2. loadCart(owner) из Redis; пусто → ErrCartEmpty
3. BeginTxFunc(PSQLConn, pgx.TxOptions{}):
     a. SELECT id, quantity, subtract, minimum, is_enable, price_*
        FROM products WHERE id = ANY($productIDs) ORDER BY id FOR UPDATE
     b. по каждой позиции корзины:
        - товар и вариант is_enable, quantity_в_заказе >= minimum
        - subtract=true → quantity >= требуемого, иначе ErrInsufficientStock(variantID)
        - unitPrice := pricing.resolveUnitPrice(user, product)
     c. subtotal = Σ line_total; shipping (v1 — фикс/по методу); tax; grand_total
        d.ExpectedTotal задан и != grand_total → ErrPriceChanged (409 + актуальная корзина)
     d. INSERT orders (status=pending, payment_status=unpaid)
     e. INSERT order_items (снапшот name/sku/slug/image/unit_price/line_total)
     f. subtract=true → UPDATE products SET quantity = quantity - $n WHERE id=$id
     g. INSERT order_status_history (null → pending, actor='customer')
4. после commit (вне транзакции):
     - CartService.ClearCart(owner)
     - MailService.Enqueue(email, OrderConfirmation{...})   // новый mail.Kind
     - Kafka publish "order.created"
   ошибки этих шагов только логируем — заказ уже создан
5. вернуть order_response.OrderResponse + данные для оплаты (v2: confirmation_url)
```

## Оплата

### v1 (быстрый старт)

`payment_method = cod` (оплата при получении) либо ручное подтверждение админом.
Заказ висит в `pending`; админ переводит `pending → paid → processing`.

### v2 (платёжный провайдер — ЮKassa / Stripe)

- `CreateOrder` создаёт платёж у провайдера, сохраняет `external_payment_id`,
  возвращает `confirmation_url`.
- Вебхук `POST /api/v1/payments/webhook/:provider`:
  - проверка подписи провайдера;
  - найти заказ по `external_payment_id`;
  - `transition(paid)` → `payment_status=paid`, `paid_at=now()`;
  - publish `order.paid` → письмо «оплачен».
- **Вебхук идемпотентен**: провайдер ретраит, повторный `paid` для оплаченного заказа — no-op `200`.
- Отдельная миграция под `orders.external_payment_id`, `payments` (если нужен лог попыток).

## Фоновый воркер экспирации

`internal/messaging/worker/order_expiry.go` (по образцу mail/image воркеров), тик раз в минуту:

```sql
SELECT id FROM orders
WHERE status = 'pending' AND created_at < now() - interval '30 minutes'
FOR UPDATE SKIP LOCKED;
```

Для каждого — в транзакции: `transition(cancelled, actor='system')`, вернуть остаток
(`UPDATE products SET quantity = quantity + qty` по позициям, где `subtract=true`),
publish `order.cancelled`. Окно (30 мин) — в конфиг `WorkersConfig`.

## Kafka-контракты

`internal/messaging/contracts/order.go`:

- `order.created` — `{order_number, user_id?, email, grand_total, currency, items[], created_at}`
- `order.paid` — `{order_number, paid_at, payment_method}`
- `order.cancelled` — `{order_number, reason, cancelled_at}`

Потребители: письма (через mail-очередь), аналитика, синхронизация с ERP/1С.

## Пограничные случаи

- **Деньги** — только серверный расчёт; `expected_total` → `409` при расхождении.
- **Порядок блокировок** — всегда `ORDER BY id` при `FOR UPDATE`, иначе дедлоки при
  пересечении товаров в параллельных заказах.
- **`FOR UPDATE SKIP LOCKED`** — только в фоновых воркерах, не в checkout (там строка нужна).
- **Гостевой заказ** — храним `email`; при последующей регистрации можно привязать по email.
- **Удаление товара** — soft-delete / `RESTRICT`; `order_items.product_id` nullable без каскада.
- **Расхождение корзины** — корзина в Redis могла измениться/истечь между просмотром и
  оформлением: checkout заново валидирует цены и остатки, при проблеме — `409` с актуальным составом.
- **Gaps в `order_number`** — последовательности не транзакционны, пропуски номеров допустимы.
- **Идемпотентность письма** — `Enqueue` после коммита; при падении процесса между коммитом и
  `Enqueue` письмо не уйдёт (допустимо для текущего тира; при необходимости — outbox-таблица).

## Порядок реализации

1. ✅ **Миграция + запросы + `make gen-sql`.** Сделано:
   - `sql/postgres/migrations/20260903171212_create_orders.{up,down}.sql`;
   - `sql/pgxgen.yaml` — CRUD-конфиг для `orders` / `order_items` / `order_status_history`;
   - хенд-запросы: `queries/orders/orders.sql`, `queries/order_items/order_items.sql`,
     `queries/order_status_history/order_status_history.sql`;
   - `queries/products/products.sql` — `GetOrderLinesByVariantIDs` (enrich + `FOR UPDATE OF p`),
     `DecrementProductStock` (`:execrows`, guarded), `RestockProduct`;
   - сгенерированы `repository_orders` / `repository_order_items` / `repository_order_status_history`;
   - проводка в `internal/storage/repository/repository.go` (`IRepository` + accessor'ы с `WithTx`).
   - Позже добавлено на шаге 2: `queries/orders/orders.sql` → `GetByNumberForUpdate`;
     `queries/products/products.sql` → `RestockOrderItems`.
   - ⚠️ В коммите шага 1 миграция `..._create_orders.up.sql` уехала с мусорным `2` в начале
     первой строки (битый SQL) — исправлено в изменениях шага 2.
2. ✅ **Сервис `order`.** `internal/service/order/`:
   - `service.go` — `IOrderService`, DI (`cfg`, `logger`, `storage`, `cart`, `user`, `mail`,
     `EventPublisher`), парсинг `config.OrderConfig` (валюта, flat/free shipping, `PendingTTL`);
   - `checkout.go` — `CreateOrder`: идемпотентность → `RawCart` → одна транзакция
     (`GetOrderLinesByVariantIDs` с `FOR UPDATE OF p` → валидация/цены → `orders.Create` →
     `order_items.Create` по строкам → `DecrementProductStock` по товарам → первая запись
     истории), после коммита best-effort (`ClearCart`, письмо `order_confirmation`,
     `publisher.OrderCreated`); гонка по idempotency-key ловится через unique-violation и
     перечитывается;
   - `calc.go` (+`calc_test.go`) — `resolveUnitPrice` (v1: всегда retail, хук под группы),
     `computeTotals`, `shippingFor`;
   - `status.go` (+`status_test.go`) — таблица переходов + `canTransition` / `restocksOn`;
   - `cancel.go` — `Cancel`: `GetByNumberForUpdate` → проверка перехода → `RestockOrderItems`
     → `MarkCancelled` → история → `publisher.OrderCancelled`;
   - `query.go` — `GetByNumber` (проверка владельца), `ListForUser` (пагинация + батч-загрузка
     строк одним запросом);
   - `publisher.go` — `EventPublisher` + `NoopPublisher` (Kafka-продюсер пока не в DI).
   - Новое: `internal/constant/order.go` (статусы), `internal/dto/order.go`,
     `internal/messaging/contracts/order.go`, `mail.OrderConfirmation` + шаблон,
     `cart.ICartService.RawCart`, поле `Services.OrderService`, `config.OrderConfig`.
   - Юнит-тесты: `calc_test.go`, `status_test.go` (чистая логика). **Интеграционные тесты
     транзакции (нехватка остатка, `ErrPriceChanged`, идемпотентность, конкурентные заказы)
     требуют БД-харнесса (`testcontainers`), которого в проекте нет — отдельная задача.**
3. ✅ **Хендлеры + `order_request` / `order_response` + Swagger.**
   - `internal/delivery/middleware/optional_auth_middleware.go` — `OptionalAuthMiddleware`:
     кладёт `user` в locals при валидном Bearer, иначе просто `Next()` (для checkout гостя/юзера).
   - `handlers/order.go`:
     - `POST /api/v1/orders` — `OptionalAuthMiddleware` + `cartOwner`; email берётся из аккаунта
       или из тела (гость); заголовок `Idempotency-Key` → `d.IdempotencyKey`;
     - `GET /api/v1/orders` — `AuthMiddleware`, пагинация;
     - `GET /api/v1/orders/:number` — `AuthMiddleware`, проверка владельца (чужой/несуществующий → 404);
     - `orderError()` — маппинг: `LineError`/`ErrCartEmpty`/`ErrEmailRequired` → 422,
       `ErrPriceChanged`/`ErrInvalidTransition` → 409, `ErrNotFound`/`ErrForbidden` → 404.
   - `request/order_request/` — `CreateOrderRequest` (валидация адреса/оплаты), `ListOrdersRequest`;
     мапперы `RequestToCreateOrderDTO` / `RequestToListOrdersDTO` в `internal/dto/order.go`.
   - `response/order_response/` — `OrderResponse` + `NewFromDTO` + `NewPaginated`.
   - `IOrderService.ListForUser` теперь возвращает
     `*base.FindResponseWithFullPagination[*dto.OrderDTO]` (единый контракт пагинации).
   - `make gen-swag` — 3 эндпоинта в `docs/`. Гостевой `GET` заказа отложен (нужен guest-токен/ссылка).
4. **Kafka-продюсер `order.*` в DI** (реализация `EventPublisher` поверх `kafka.Producer`) +
   темплейт письма уже есть.
5. **Воркер экспирации** (`ListExpiredPending` + `Cancel`, actor=`system`) — `Cancel` готов.
6. **Админские эндпоинты**: список заказов, смена статуса, отмена/возврат.
7. **Оплата (v2)** — после выбора провайдера.
