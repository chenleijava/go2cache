/*
 * Copyright (c) 2018. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package go2cache

import (
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"log"
)

type PubSub struct {
	Client       *redis.Pool
	Channel      string
	CacheChannel *CacheChannel
	Region       string
}

//j2cache  发布 订阅消息模块 封装
const (
	OPT_JOIN      = iota + 1 // 加入集群
	OPT_EVICT_KEY            // 删除集群
	OPT_CLEAR_KEY            // 清理缓存
	OPT_QUIT                 // 退出集群环境
)

//j2cache command
type Command struct {
	Region   string   `json:"region"`
	Operator int      `json:"operator"`
	Keys     []string `json:"keys"`
	Src      int      `json:"src"`
}


//发送清楚缓存的广播命令
func (p *PubSub) SendEvictCmd(region string, keys ...string) {
	data, _ := json.Marshal(&Command{Region: region, Keys: keys, Operator: OPT_EVICT_KEY})
	conn := p.Client.Get()
	defer conn.Close()                            // close
	_, err := conn.Do("PUBLISH", p.Channel, data) // 指Channel 发布 信息
	if err != nil {
		log.Printf("error in pubish , info:%s", err)
	}
}

//发送清除缓存的广播命令
func (p *PubSub) SendClearCmd(region string) {
	data, _ := json.Marshal(&Command{Region: region, Keys: nil, Operator: OPT_CLEAR_KEY})
	conn := p.Client.Get()
	defer conn.Close()
	_, err := conn.Do("PUBLISH", p.Channel, data) // 指Channel 发布 信息
	if err != nil {
		log.Printf("error in pubish , info:%s", err)
	}
}

//初始化订阅
func (p *PubSub) Subscribe() {
	psc := redis.PubSubConn{Conn: p.Client.Get()}
	err := psc.Subscribe(p.Channel)
	if err != nil {
		log.Fatalf("subscribe error:%s", err)
	}
	//在子协程中完成订阅 发布的监听，防止主线程挂起
	go func() {
		for {
			switch message := psc.Receive().(type) {
			case redis.Message:
				var cmd Command
				e := json.Unmarshal(message.Data, &cmd)
				if e != nil {
					log.Printf("command unmarshl json error:%s", e)
				}
				if cmd.Operator == OPT_EVICT_KEY { //删除一级缓存数据
					//log.Printf("evict key :%s region:%s", cmd.Keys, cmd.Region)
					p.CacheChannel.Evict(cmd.Region, cmd.Keys)
				} else if cmd.Operator == OPT_CLEAR_KEY { //  清除缓存
					log.Printf("clear cache  key :%s region:%s", cmd.Keys, cmd.Region)
				} else if cmd.Operator == OPT_JOIN { // 节点加入
					log.Printf("node-%d join cluster ", cmd.Src)
				} else if cmd.Operator == OPT_QUIT { //节点离开
					log.Printf("node-%d quit cluster ", cmd.Src)
				}
			case redis.Subscription:
				// Kind is "subscribe", "unsubscribe", "psubscribe" or "punsubscribe"
				log.Printf("subscription  channel :%s  %s  count:%d\n", message.Channel, message.Kind, message.Count)
			case error:
				//redis restart or down
				//retry to connect redis server
				log.Printf("redis done :%s begin retry Subscribe ... ...", message)
				psc.Close()               // close error connection
				psc.Conn = p.Client.Get() // 如果redis链接失败将会发起重新链接
				psc.Subscribe(p.Channel)
			}
		}
	}()
}
