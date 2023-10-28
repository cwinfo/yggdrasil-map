package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

var waitgroup sync.WaitGroup

const MAX_RETRY = 3
const N_PARALLEL_REQ = 32

var semaphore chan struct{}

func init() {
	semaphore = make(chan struct{}, N_PARALLEL_REQ)
}

func dial() (net.Conn, error) {
	return net.DialTimeout("unix", "/var/run/yggdrasil.sock", time.Second)
}

func getRequest(key, request string) map[string]interface{} {
	return map[string]interface{}{
		"keepalive": true,
		"request":   request,
		"arguments": map[string]interface{}{
			"key": key,
		},
	}
}

func doRequest(request map[string]interface{}) map[string]interface{} {
	req, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	var res map[string]interface{}
	for idx := 0; idx < MAX_RETRY; idx++ {
		sock, err := dial()
		if err != nil {
			panic(err)
		}
		if _, err = sock.Write(req); err != nil {
			panic(err)
		}
		bs := make([]byte, 65535)
		deadline := time.Now().Add(6 * time.Second)
		sock.SetReadDeadline(deadline)
		n, err := sock.Read(bs)
		sock.Close()
		if err != nil {
			continue
			panic(bs)
		}
		bs = bs[:n]
		if err = json.Unmarshal(bs, &res); err != nil {
			return nil
			panic(err)
		}
		// TODO parse res, check if there's an error
		if res, ok := res["response"]; ok {
			if _, ok := res.(map[string]interface{}); !ok {
				continue
			}
			if _, isIn := res.(map[string]interface{})["error"]; isIn {
				continue
			}
		}
		break
	}
	return res
}

func getNodeInfo(key string) map[string]interface{} {
	return doRequest(getRequest(key, "getNodeInfo"))
}

func getSelf(key string) map[string]interface{} {
	return doRequest(getRequest(key, "debug_remoteGetSelf"))
}

func getPeers(key string) map[string]interface{} {
	return doRequest(getRequest(key, "debug_remoteGetPeers"))
}

func getDHT(key string) map[string]interface{} {
	return doRequest(getRequest(key, "debug_remoteGetDHT"))
}

func getPaths() map[string]interface{} {
	req := map[string]interface{}{
		"keepalive": true,
		"request":   "getPaths",
	}
	return doRequest(req)
}

type rumorResult struct {
	key string
	res map[string]interface{}
}

func doRumor(key string, out chan rumorResult) {
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		semaphore <- struct{}{}
		defer func() { <-semaphore }()
		results := make(map[string]interface{})
		if res, ok := getNodeInfo(key)["response"]; ok {
			if _, ok := res.(map[string]interface{}); !ok {
				return
			}
			for addr, v := range res.(map[string]interface{}) {
				vm, ok := v.(map[string]interface{})
				if !ok {
					return
				}
				results["address"] = addr
				results["nodeinfo"] = vm
			}
		}
		
		if res, ok := getPaths()["response"]; ok {
			if _, ok := res.(map[string]interface{}); !ok {
				return
			}
			psi := res.(map[string]interface{})["paths"]
			if _, ok := psi.([]interface{}); !ok {
				return
			}
			ps := psi.([]interface{})
			for _, pi := range ps {
				if _, ok := pi.(map[string]interface{}); !ok {
					return
				}
				p := pi.(map[string]interface{})
				if _, ok := p["key"]; !ok {
					return
				}
				ki := p["key"]
				if _, ok := ki.(string); !ok {
					return
				}
				k := ki.(string)
				if k != key {
					continue
				}
				if _, ok := p["path"]; !ok {
					return
				}
				results["coords"] = fmt.Sprintf("%v", p["path"])
			}
		}
		/*
		if res, ok := getSelf(key)["response"]; ok {
			if _, ok := res.(map[string]interface{}); !ok {
				return
			}
			for _, v := range res.(map[string]interface{}) {
				vm, ok := v.(map[string]interface{})
				if !ok {
					return
				}
				if coords, ok := vm["coords"]; ok {
					results["coords"] = coords
				}
			}
		}
		*/
		if res, ok := getPeers(key)["response"]; ok {
			if _, ok := res.(map[string]interface{}); !ok {
				return
			}
			for _, v := range res.(map[string]interface{}) {
				vm, ok := v.(map[string]interface{})
				if !ok {
					return
				}
				if keys, ok := vm["keys"]; ok {
					results["peers"] = keys
				}
			}
		}
		/*
		if res, ok := getDHT(key)["response"]; ok {
			if _, ok := res.(map[string]interface{}); !ok {
				return
			}
			for _, v := range res.(map[string]interface{}) {
				vm, ok := v.(map[string]interface{})
				if !ok {
					return
				}
				if keys, ok := vm["keys"]; ok {
					results["dht"] = keys
				}
			}
		}
		*/
		if len(results) > 0 {
			results["time"] = time.Now().Unix()
			out <- rumorResult{key, results}
			waitgroup.Add(1)
		}
	}()
}

func doPrinter() (chan rumorResult, chan struct{}) {
	results := make(chan rumorResult)
	done := make(chan struct{})
	go func() {
		defer close(done)
		visited := make(map[string]struct{})
		fmt.Println("{\"yggnodes\": {")
		var notFirst bool
		for result := range results {
			// TODO correct output
			if _, isIn := visited[result.key]; isIn {
				waitgroup.Done()
				continue
			}
			visited[result.key] = struct{}{}
			res, err := json.Marshal(result.res)
			if err != nil {
				panic(err)
			}
			if notFirst {
				fmt.Println(",")
			}
			fmt.Printf("\"%s\": %s", result.key, res)
			notFirst = true
			toVisit := make(map[string]struct{})
			if peers, isIn := result.res["peers"]; isIn {
				if ks, ok := peers.([]interface{}); ok {
					for _, k := range ks {
						if key, ok := k.(string); ok {
							toVisit[key] = struct{}{}
						}
					}
				}
			}
			if dht, isIn := result.res["dht"]; isIn {
				if ks, ok := dht.([]interface{}); ok {
					for _, k := range ks {
						if key, ok := k.(string); ok {
							toVisit[key] = struct{}{}
						}
					}
				}
			}
			for k := range toVisit {
				if _, isIn := visited[k]; !isIn {
					doRumor(k, results)
				}
			}
			waitgroup.Done()
		}
		fmt.Println("\n}}")
	}()
	return results, done
}

func main() {
	self := doRequest(map[string]interface{}{"keepalive": true, "request": "getSelf"})
	res := self["response"].(map[string]interface{})
	var key string
	key = res["key"].(string)
	results, done := doPrinter()
	doRumor(key, results)
	waitgroup.Wait()
	close(results)
	<-done
}
