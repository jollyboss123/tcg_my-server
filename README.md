# TCG.MY
Server for [TCG.MY](https://github.com/hollandgeng/TCG.MY). API for retrieving Yugioh OCG cards' price by scraping data from [bigweb](https://bigweb.co.jp/) and [YUYU-TEI](https://yuyu-tei.jp/).

### Design
1. web scraping with [go-colly](https://github.com/gocolly/colly)
2. graphql http caching with [apq](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq) and lru cache
3. graphql response caching with redis

### Todo
- [ ] data loading with graphql
- [ ] currency conversion query
- [ ] jwt authentication on APIs
- [ ] structured logging with slog

### Reference Documentation

* [Scraping the Web in Golang with Colly and Goquery](https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/)
* [Web Scraping in Golang: Complete Guide 2023](https://www.zenrows.com/blog/web-scraping-golang#scrape-product-data)
* [web crawl nested HTML elements](https://github.com/gocolly/colly/issues/376)
* [Automatic persisted queries (APQ)](https://www.apollographql.com/docs/resources/graphql-glossary/#automatic-persisted-queries-apq)
* [APQ with gqlgen](https://gqlgen.com/reference/apq/)
* [Research project: Which is the best caching strategy with GraphQL for a big relational database?](https://medium.com/@niels.onderbeke.no/research-project-which-is-the-best-caching-strategy-with-graphql-for-a-big-relational-database-56fedb773b97)
