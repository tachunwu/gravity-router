# gravity-router
gravity-router 主要實作 [Scaling Memcache at Facebook](https://www.usenix.org/system/files/conference/nsdi13/nsdi13-final170_update.pdf) 這篇 Distirbuted Cache 的概念，不過用 Go 語言實作，並且設計可以部署在 kubernetes 之上。

## Architecture & Design



## Component
### Aggregate Layer
Aggregate Layer 主要由 NATS 所組成，扮演了 Leaf 的 Ambassador，詳細概念可以參考 Brendan Burns 的論文 [Design patterns for container-based distributed systems](https://static.googleusercontent.com/media/research.google.com/zh-TW//pubs/archive/45406.pdf)。
![](./asserts/ambassador.png)
還有如果直接用 Application Server 和 N 個 Leaf Node 溝通，連線的數量是 O(n^2)，所以加入一層 Proxy 可以大幅減低內部網路擁塞。

#### Replicated Pool
Replicated Pool 其中一種模型，在 Replicated Pool 中的 Leaf Node 都擁有相同的副本，這樣有幾個好處。首先，可以做 Hedged requests，當有多個副本的時候，可以讓最低延遲的服務先回應，可以大幅減低 P99 的延遲，詳細參考 Jeff Dean 的 Google 伺服器管理論文 [The Tail at Scale](https://cseweb.ucsd.edu/classes/sp18/cse124-a/post/schedule/p74-dean.pdf)。
![](./asserts/hedged.png)

再來，就算是 Leaf 故障，也可以讓沒有故障的副本繼續服務達到 HA 的效果，最後如果有 Hot Spot 也可以經由分散 Requests 來應付高流量。

#### Sharded Pool
Sharded Pool 則是採用一致性雜湊將特定的 Key 分配到特定的 Shard 上，可以達到水平擴展的需求。

### Leaf Layer
Leaf Layer 單純是獨立的 In-memory 負責儲存 Cache 的資料，原始論文使用的是 memcached，不過這裡採用 Go 的實作 [ristretto](https://github.com/dgraph-io/ristretto)，擁有非常高的 Throughput，被許多 Go 的 Database 所使用，如下列表:
* Badger - Embeddable key-value DB in Go
* Dgraph - Horizontally scalable and distributed GraphQL database with a graph backend
* Vitess - Database clustering system for horizontal scaling of MySQL
* SpiceDB - Horizontally scalable permissions database
#### Benchmark
![](./asserts/mixed.svg)
![](./asserts/read.svg)
![](./asserts/write.svg)

## Deploy

## Getting Started
```go
func main() {

	client := NewClient()

	key := "key"
	value := []byte("value")

	err := client.Set(context.Background(), key, value, 3)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set operation completed")

	client.Get(key)

	err = client.Del(context.Background(), key, 3)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Del operation completed")

}
```

## Reference
* [Scaling Memcache at Facebook](https://www.usenix.org/system/files/conference/nsdi13/nsdi13-final170_update.pdf)
* [Introducing mcrouter: A memcached protocol router for scaling memcached deployments](https://engineering.fb.com/2014/09/15/web/introducing-mcrouter-a-memcached-protocol-router-for-scaling-memcached-deployments/)
* [Turning Caches into Distributed Systems with mcrouter - Data@Scale](https://www.youtube.com/watch?v=e9lTgFO-ZXw&list=PLb0IAmt7-GS0HarXUJP6v4I5IPaCRkX3c&index=10)

