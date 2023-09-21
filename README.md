# TCG.MY
Server for [TCG.MY](https://github.com/hollandgeng/TCG.MY). API for retrieving Yugioh OCG cards' price by scraping data from [bigweb](https://bigweb.co.jp/) and [YUYU-TEI](https://yuyu-tei.jp/).

## Design
1. web scraping with [go-colly](https://github.com/gocolly/colly)
2. graphql http caching with [apq](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq) and lru cache
3. graphql response caching with redis
4. structured logging with slog

## Redis
Redis is a key-value store. In this instance, it is used as [response caching](https://www.apollographql.com/docs/apollo-server/features/caching),
which would let us effectively bypass the resolver for one or more fields and use the cached value
instead (until it expires or is invalidated). It acts as a cache layer between GraphQL and our data source i.e. the websites being scraped.
### Data model
#### Cache structure
Each card is cached in Redis using a combination of the card's attributes. The cache key follows the format: `<Rarity>||<Name>||<Code>`
#### Queries
Queries can be based on a card's code, name, or associated booster pack. To accommodate this, we use different search patterns:
- Code: `*||<Query>`
- Name: `*||<Query>||*`
- Booster: `*||<Query>-*`

This approach helps us retrieve cached cards without knowing the exact attributes of a card.
#### Cache markers for queries
To prevent unnecessary external service calls, we keep a separate cache entry for each unique query. These entries are not the card data itself but just markers indicating that a particular query has been made before. These are stored with the key format: `query:<Query>`.
The value for these markers is just the string "true".

## Reference Documentation

* [Scraping the Web in Golang with Colly and Goquery](https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/)
* [Web Scraping in Golang: Complete Guide 2023](https://www.zenrows.com/blog/web-scraping-golang#scrape-product-data)
* [web crawl nested HTML elements](https://github.com/gocolly/colly/issues/376)
* [Automatic persisted queries (APQ)](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq)
* [APQ with gqlgen](https://gqlgen.com/reference/apq/)
* [Research project: Which is the best caching strategy with GraphQL for a big relational database?](https://medium.com/@niels.onderbeke.no/research-project-which-is-the-best-caching-strategy-with-graphql-for-a-big-relational-database-56fedb773b97)
* [Load real env vars in production and use a .env locally](https://github.com/joho/godotenv/issues/40)
