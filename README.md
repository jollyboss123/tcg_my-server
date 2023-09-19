# TCG.MY
Server for [TCG.MY](https://github.com/hollandgeng/TCG.MY). API for retrieving Yugioh OCG cards' price by scraping data from [bigweb](https://bigweb.co.jp/) and [YUYU-TEI](https://yuyu-tei.jp/).

### Design
1. web scraping with [go-colly](https://github.com/gocolly/colly)
2. graphql http caching with [apq](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq) and lru cache
3. graphql response caching with redis
4. structured logging with slog

### Todo
- [ ] data loading with graphql
- [ ] jwt authentication on APIs
- [ ] add rate limit

### Setup
1. copy `docker-compose.yml` to your root directory
2. run docker at your root directory
```shell
docker compose up
```
3. access graphiql playground at http://0.0.0.0:8080/graphiql
4. available routes:
   1. graphiql playground: http://0.0.0.0:8080/graphiql
   2. query: http://0.0.0.0:8080/query
   3. health: http://0.0.0.0:8080/health
5. bring down docker when done
```shell
docker compose down
```

### Reference Documentation

* [Scraping the Web in Golang with Colly and Goquery](https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/)
* [Web Scraping in Golang: Complete Guide 2023](https://www.zenrows.com/blog/web-scraping-golang#scrape-product-data)
* [web crawl nested HTML elements](https://github.com/gocolly/colly/issues/376)
* [Automatic persisted queries (APQ)](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq)
* [APQ with gqlgen](https://gqlgen.com/reference/apq/)
* [Research project: Which is the best caching strategy with GraphQL for a big relational database?](https://medium.com/@niels.onderbeke.no/research-project-which-is-the-best-caching-strategy-with-graphql-for-a-big-relational-database-56fedb773b97)
* [Load real env vars in production and use a .env locally](https://github.com/joho/godotenv/issues/40)
