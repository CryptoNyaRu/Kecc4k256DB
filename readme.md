## 动机

- [4byte](https://www.4byte.directory/) 提供 EVM 签名查询服务, 但不开放完整数据库下载
- 为满足本地高频查询需求, Kecc4k256DB 支持对 [4byte API](https://www.4byte.directory/docs/) 数据的完整爬取与增量更新

## 数据库信息

- 维护时间: ***2025-05-13***
- 方法签名: ***1122273***
- 事件签名: ***263810***
- 文件大小: ***84.3 MB (88,440,832 字节)***

**数据库随缘维护, 请自行定期进行增量更新!!!**

## 完整爬取/增量更新

```golang
kecc4k256DB, err := kecc4k256db.Open("./kecc4k256.db")
if err != nil {
	log.Fatalf("Failed to open: %s\n", err)
}
	
// UpdateSync
log.Println("UpdateSync will start in 3 seconds...")
time.Sleep(3 * time.Second)

kecc4k256DB.UpdateSync(&kecc4k256db.Logger{
    Info:    log.Println,
    Success: log.Println,
    Warning: log.Println,
    Error:   log.Println,
})

log.Println("UpdateSync done")
```

- 不存在的数据库被打开时将会自动创建
- 若你未在自己的项目中维护数据库, 进行增量更新需要将数据库复制到 ***cmd/kecc4k256.db*** 并运行 ***cmd/main.go***
- 若你已在自己的项目中维护数据库, 推荐使用 **UpdateSync(logger *Logger)**, 在需要打印 log 时传入你的 logger

## 注意事项

- **请不要使用别人编译好的可执行文件, 务必自行编译!!!**
- **请不要使用别人编译好的可执行文件, 务必自行编译!!!**
- **请不要使用别人编译好的可执行文件, 务必自行编译!!!**

## 鸣谢

#### [4byte TOS and Licensing](https://www.4byte.directory/#:~:text=Search-,TOS%20and%20Licensing,-The%20data%20from)

> The data from this service is given free of any license or restrictions on how it may be used.
> 
> Usage of the API is also granted with the single restriction that your usage should not disrupt the service itself. If you wish to scrape this data, feel free, but please do so with limited concurrency. You are encouraged to scrape your own copy of this data if your use case is likely to produce heavy load on the API.

- Kecc4k256DB 的数据来源于 [4byte API](https://www.4byte.directory/docs/)
- [4byte](https://www.4byte.directory/) 允许用户在遵守公平使用原则的情况下自由地使用其数据, Kecc4k256DB 遵守了这一规则, 未使用高并发更新数据库
  