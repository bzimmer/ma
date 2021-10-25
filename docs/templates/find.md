`find` will look nodes for the specified query. If `--scope` is not specified, the currently
authenticated user's scope is used.

```sh
$ ma find --scope "/api/v2/user/cmac" Event
2021-10-24T18:38:52-07:00 INF find name=Events nodeID=kTR76 type=Folder
2021-10-24T18:38:52-07:00 INF find albumKey=mjXDhW imageCount=79 name="Harley Pan America event" nodeID=sBt5dp type=Album
2021-10-24T18:38:52-07:00 INF find albumKey=dDMKTz imageCount=6 name="SmugMug camera straps!" nodeID=vGtbt type=Album
2021-10-24T18:38:52-07:00 INF counters count=2 metric=ma.find.album
2021-10-24T18:38:52-07:00 INF counters count=1 metric=ma.find.node
```

```sh
$ ma find --scope "/api/v2/user/cmac" SmugMug
2021-10-24T18:40:21-07:00 INF find name=SmugMug nodeID=XWx8t type=Folder
2021-10-24T18:40:23-07:00 INF find albumKey=fqH3fM imageCount=16 name="SmugMug videos" nodeID=99SCh type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=TrBCmb imageCount=99 name="SmugMug Heroes" nodeID=Bpj6s type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=bKfV9Z imageCount=5 name="SmugMug Stickers!" nodeID=XWhVs type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=fk58cR imageCount=6 name="Powered by smugmug" nodeID=CxKrQ type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=Qp6Vk5 imageCount=35 name="SmugMug Halloween 2012" nodeID=NzST7 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=dDMKTz imageCount=6 name="SmugMug camera straps!" nodeID=vGtbt type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=Q4JCgb imageCount=64 name="SmugMug Tahoe Soiree" nodeID=gQX9t type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=wKjKNd imageCount=1 name="SmugMug August 2012" nodeID=kDCZZ type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=9nmSBj imageCount=104 name="Fark/smugmug photoshop contest" nodeID=tZgNn type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=GxmcTJ imageCount=24 name="SmugMug homepage slide show" nodeID=SH5wj type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=bMh8jp imageCount=81 name="SmugMug Photo booth 2013" nodeID=kdCn8 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=dgfgxg imageCount=4 name="SmugMug on an iPhone" nodeID=Q6qvZ type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=pXPZ4k imageCount=90 name="SmugMug Christmas party 2012" nodeID=wWSwG type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=D4nm6Z imageCount=6 name="Past smugmug home pages" nodeID=zLWT2 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=XqxN46 imageCount=44 name="Tahoe Soiree" nodeID=VC5j4 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=pZ9XFH imageCount=7 name="smugmug 1st annual Christmas brekkie" nodeID=KK9Dx type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=TJstLc imageCount=37 name="Possible smugmug home page thumbs" nodeID=cpVJr type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=243FWz imageCount=1 name="SmugMug Gym Shoot 2016 with Von Wong" nodeID=Kwz5kc type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=4hxRpd imageCount=0 name="My SmugMug Site Files (Do Not Delete)" nodeID=fPR7p type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=bT2KLC imageCount=129 name="Green hair!" nodeID=4n2G7 type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=WJvpCp imageCount=19 name="SmugMug's green hair" nodeID=7CMcv type=Album
2021-10-24T18:40:23-07:00 INF find albumKey=NjqVwL imageCount=20 name="SmugMug's Windows 7 party with seven 7x7 burgers" nodeID=HcHRq type=Album
2021-10-24T18:40:23-07:00 INF counters count=1 metric=ma.find.node
2021-10-24T18:40:23-07:00 INF counters count=22 metric=ma.find.album
```
