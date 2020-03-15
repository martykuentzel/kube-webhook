# k8s-mutate-webhook

A playground to try build a crude k8s mutating webhook; the goal is to mutate a Pod CREATE request to _always_ use a debian image and by doing this, learning more about
the k8s api, objects, etc. - eventually figure out how scalable this is (could be made) if one had 1000 pods to schedule (concurrently) 

## build 

```
make
```

## test

```
make test
```

## docker

to create a docker image .. 

```
make docker
```

it'll be tagged with the current git commit (short `ref`) and `:latest`

don't forget to update `IMAGE_PREFIX` in the Makefile or set it when running `make`

### images




## watcher

useful during devving ... 


## kudos

- [https://github.com/morvencao/kube-mutating-webhook-tutorial](https://github.com/morvencao/kube-mutating-webhook-tutorial)
