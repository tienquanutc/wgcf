```shell
docker service create --replicas 5 --publish published=6688,target=6688  --name wgcf-server  wgcf
```
