`ls` returns all the nodes under the specified parent node.

```sh
$ ma ls node -R kTR76
2021-10-24T18:41:37-07:00 INF ls name=Events nodeID=kTR76 parentID=zx4Fx type=Folder
2021-10-24T18:41:38-07:00 INF ls albumKey=jbBNhR imageCount=16 name="Black lives matter protest" nodeID=q2qP7F parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=GNpNRf imageCount=11 name="47th annual electric vehicle show silicon valley" nodeID=6zFLgD parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=LPZThB imageCount=21 name="Humans melt and pets soak up love at the Bay Area Pet Fair" nodeID=MG5c5r parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=MGxZXq imageCount=36 name="San Francisco Reptile Expo" nodeID=kvh7p7 parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=KJXpSD imageCount=25 name="Muttville adoption ceremony" nodeID=JbnGdJ parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=pttBvK imageCount=18 name="SF Body Art Expo 2018" nodeID=xH9vPF parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=gJ7d6r imageCount=35 name="Maker Faire 2008" nodeID=BjcPR parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=2XrGxm imageCount=33 name="Santa Cruz fungus festival 2019" nodeID=jM739r parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=B8BLsX imageCount=30 name="Golden Gate dog show 2019" nodeID=LtqS2s parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=wqmpff imageCount=22 name="Spanish Fork High 50th reunion 2019" nodeID=5FD6k8 parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF ls albumKey=qH5CG3 imageCount=35 name="Spanish Fork Fiesta Days parade 2019" nodeID=J3wNMj parentID=kTR76 type=Album
2021-10-24T18:41:38-07:00 INF counters count=12 metric=ma.ls.node
```

If `ls` queries an album, only the album details are returned.

```sh
$ ma ls node -R q2qP7F
2021-10-24T18:42:52-07:00 INF ls albumKey=jbBNhR imageCount=16 name="Black lives matter protest" nodeID=q2qP7F parentID=kTR76 type=Album
2021-10-24T18:42:52-07:00 INF counters count=1 metric=ma.ls.node
```

To list the images for an album, use the `-i` flag.

```sh
$ ma ls node -R -i q2qP7F
2021-10-24T18:44:30-07:00 INF ls albumKey=jbBNhR imageCount=16 name="Black lives matter protest" nodeID=q2qP7F parentID=kTR76 type=Album
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--3.jpg imageKey=pjVHggG imageURI=/api/v2/album/jbBNhR/image/pjVHggG-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--6.jpg imageKey=Wzk7Lk5 imageURI=/api/v2/album/jbBNhR/image/Wzk7Lk5-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--7.jpg imageKey=mQRcX2V imageURI=/api/v2/album/jbBNhR/image/mQRcX2V-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--8.jpg imageKey=Vhhbw4j imageURI=/api/v2/album/jbBNhR/image/Vhhbw4j-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--9.jpg imageKey=3MgjDCT imageURI=/api/v2/album/jbBNhR/image/3MgjDCT-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--10.jpg imageKey=jfChpW8 imageURI=/api/v2/album/jbBNhR/image/jfChpW8-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--11.jpg imageKey=VDt5z9K imageURI=/api/v2/album/jbBNhR/image/VDt5z9K-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--12.jpg imageKey=S9jS4PW imageURI=/api/v2/album/jbBNhR/image/S9jS4PW-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--13.jpg imageKey=f6hWww5 imageURI=/api/v2/album/jbBNhR/image/f6hWww5-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--14.jpg imageKey=JdL4NbN imageURI=/api/v2/album/jbBNhR/image/JdL4NbN-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--15.jpg imageKey=F8QQhnN imageURI=/api/v2/album/jbBNhR/image/F8QQhnN-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--16.jpg imageKey=bvh5hgQ imageURI=/api/v2/album/jbBNhR/image/bvh5hgQ-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--17.jpg imageKey=g6F7kfS imageURI=/api/v2/album/jbBNhR/image/g6F7kfS-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--18.jpg imageKey=wd626B2 imageURI=/api/v2/album/jbBNhR/image/wd626B2-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--19.jpg imageKey=DLp2Mkt imageURI=/api/v2/album/jbBNhR/image/DLp2Mkt-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF ls albumKey=jbBNhR caption= filename=flowers--20.jpg imageKey=B9C6jjx imageURI=/api/v2/album/jbBNhR/image/B9C6jjx-0 keywords=["flowers"] type=Image version=0
2021-10-24T18:44:31-07:00 INF counters count=16 metric=ma.ls.image
2021-10-24T18:44:31-07:00 INF counters count=1 metric=ma.ls.node
```
