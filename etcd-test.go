package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"path/filepath"
	"time"
)

type Upstream struct {
	Name string   `json:"name"`
	List []string `json:"list"`
}

type Dy_Upstream struct {
	Action string
	Name   string
	List   []string
}

var service map[string][]string

func NewCli(endpoints []string) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("eeee", err)
	}
	return cli
}

func Get(cli *clientv3.Client, prefix string) {
	resp, _ := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	fmt.Printf("Getting prefix:%s now...\n", prefix)
	for _, ev := range resp.Kvs {
		//fmt.Printf("%s: %s\n", ev.Key, ev.Value)
		var t Upstream
		json.Unmarshal([]byte(ev.Value), &t)
		service[t.Name] = t.List
	}
    fmt.Println(service)
}

func Watch(cli *clientv3.Client, prefix string, ch chan Dy_Upstream) {
	rch := cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	fmt.Printf("Watching prefix:%s now...\n", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				//fmt.Printf("PUT %s - %s\n", string(ev.Kv.Key), string(ev.Kv.Value))
				var t Upstream
				json.Unmarshal([]byte(ev.Kv.Value), &t)
				ch <- Dy_Upstream{
					Action: "PUT",
					Name:   t.Name,
					List:   t.List,
				}
			case mvccpb.DELETE:
				key := filepath.Base(string(ev.Kv.Key))
				ch <- Dy_Upstream{
					Action: "DEL",
					Name:   key,
				}
			}
		}
	}
}
func main() {
	service = make(map[string][]string)
	ch := make(chan Dy_Upstream)
	ed := []string{"localhost:2379"}
	src := NewCli(ed)
	Get(src, "/srv/")
	go Watch(src, "/srv/", ch)
	for {
		select {
		case t := <-ch:
			if t.Action == "PUT" {
				service[t.Name] = t.List
			} else if t.Action == "DEL" {
				delete(service, t.Name)
			}
		}
		fmt.Println(service)
	}
}
