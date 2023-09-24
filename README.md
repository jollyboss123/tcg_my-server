# TCG.MY

Server for [TCG.MY](https://github.com/hollandgeng/TCG.MY). GraphQL APIs for retrieving Yugioh OCG cards' price by
scraping data
from [bigweb](https://bigweb.co.jp/) and [YUYU-TEI](https://yuyu-tei.jp/).

## High Level Design

1. web scraping with [go-colly](https://github.com/gocolly/colly)
2. graphql http caching
   with [apq](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq) and lru
   cache
3. graphql response caching with redis
4. structured logging with slog

## Endpoints

| endpoint   | route     |
|:-----------|:----------|
| playground | /graphiql |
| query      | /query    |
| health     | /health   |

## Redis

Redis is a key-value store. In this instance, it is used
as [response caching](https://www.apollographql.com/docs/apollo-server/features/caching),
which would let us effectively bypass the resolver for one or more fields and use the cached value
instead (until it expires or is invalidated). It acts as a cache layer between GraphQL and our data source i.e. the
websites being scraped.

### Data model

#### Cache structure

Each card is cached in Redis using a unique combination of the card's attributes. The cache key follows the
format: `<Rarity>||<Name>||<Code>`

#### Queries

Queries can be based on a card's code, name, or associated booster pack. To accommodate this, we use different search
patterns:

- Code or Booster: `*||*||*<Query>*`
- Name: `*||*<Query>*||*`

This approach helps us retrieve cached cards without knowing the exact attributes of a card.

#### Cache markers for queries

To optimize our external service calls, we maintain a separate cache entry for each unique query made.
These entries are markers indicating prior queries and do not store the actual card data.
They follow the format: `query:<Game>:<Query>`

#### Caching logic

Redis's sets and hashes is utilized to enhance our caching strategy.
Once we have the card identifiers from the sets, we can retrieve the full card data from the hashes.

This combination of sets and hashes allow for faster membership checks and storing card data
in hashes can be more space-efficient than simple k-v pair when there's lots of data.

- Sets:
  Sets are employed to track unique card identifiers for a specific game. Each game has its own set where each member is
  a unique combination of card attributes, formatted as `<Rarity>||<Name>||<Code>`. This makes it quick and efficient to
  determine if a card exists in the cache and to perform pattern-based searches across cards.
- Hashes:
  While sets give us a fast way to identify the existence of a card or to search across them, hashes store the actual
  data of these cards.
  For each game, there's a hash where the key is the unique card identifier and the value is the serialized card data.

## Reference Documentation

* [Scraping the Web in Golang with Colly and Goquery](https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/)
* [Web Scraping in Golang: Complete Guide 2023](https://www.zenrows.com/blog/web-scraping-golang#scrape-product-data)
* [web crawl nested HTML elements](https://github.com/gocolly/colly/issues/376)
* [Automatic persisted queries (APQ)](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq)
* [APQ with gqlgen](https://gqlgen.com/reference/apq/)
* [Research project: Which is the best caching strategy with GraphQL for a big relational database?](https://medium.com/@niels.onderbeke.no/research-project-which-is-the-best-caching-strategy-with-graphql-for-a-big-relational-database-56fedb773b97)
* [Load real env vars in production and use a .env locally](https://github.com/joho/godotenv/issues/40)
