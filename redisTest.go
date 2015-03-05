package main

import (
	"fmt"
	"github.com/fzzy/radix/redis"
	//"time"
)

func main() {

	client, err := redis.Dial("tcp", "178.62.74.225:6379")
	handleErr(err)
	defer client.Close()

	foo, err := client.Cmd("PING").Str()
	handleErr(err)
	fmt.Println(foo)

	foo, err = client.Cmd("FLUSHALL").Str()
	handleErr(err)
	fmt.Println(foo)

	info := map[string]string{
		"name":        "test vote",
		"description": "test vote descrip",
		"V1":          "pizza",
		"V2":          "pasta",
		"V3":          "soup",
		"V4":          "turds",
	}

	setBallot(client, "1234", info, 5)
	//time.Sleep(time.Second * 11)
	setVote(client, "1234", "102.169.123", "V1")
	getBallot(client, "1234")
	//setVote(client, "1234", "102.169.123", "V1")
	//setVote(client, "1234", "102.169.123", "V1")
}

func setBallot(client *redis.Client, dbName string, info map[string]string, voteTimeout int) {

	//set the info hash and give a extra day of timeout
	_, err := client.Cmd("HMSET", dbName+"Info", info).Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"Info", voteTimeout+86400).Int()
	handleErr(err)

	//set the IP set and also give a extra day of timeout
	_, err = client.Cmd("SADD", dbName+"IpSet", "1").Int()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"IpSet", voteTimeout+86400).Int()
	handleErr(err)

	//set v1 - v5 and add the timeout
	_, err = client.Cmd("SET", dbName+"V1", "0").Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"V1", voteTimeout+86400).Int()
	handleErr(err)
	_, err = client.Cmd("SET", dbName+"V2", "0").Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"V2", voteTimeout+86400).Int()
	handleErr(err)
	_, err = client.Cmd("SET", dbName+"V3", "0").Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"V3", voteTimeout+86400).Int()
	handleErr(err)
	_, err = client.Cmd("SET", dbName+"V4", "0").Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"V4", voteTimeout+86400).Int()
	handleErr(err)
	_, err = client.Cmd("SET", dbName+"V5", "0").Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"V5", voteTimeout+86400).Int()
	handleErr(err)

	//set the write timeout
	_, err = client.Cmd("SET", dbName+"Write", "1").Str()
	handleErr(err)
	_, err = client.Cmd("EXPIRE", dbName+"Write", voteTimeout).Int()
	handleErr(err)

}

func setVote(client *redis.Client, dbName string, ipAddress string, v string) {
	//check the IP set if the IP address has already voted or not
	ipPresent, err := client.Cmd("SADD", dbName+"IpSet", ipAddress).Int()
	handleErr(err)
	if ipPresent == 0 {
		fmt.Println("IP " + ipAddress + " already voted, go away!")
		return
	}

	//check write to see if timedout
	write, err := client.Cmd("EXISTS", dbName+"Write").Int()
	fmt.Println(write)
	if write == 0 {
		fmt.Println("Ballot " + dbName + " write timeout, no more voting, go away!")
		return
	}

	//add vote code here
	fmt.Println("Vote MUTHER FUCKER!")
	_, err = client.Cmd("INCR", dbName+v).Int()
	handleErr(err)
}

func getBallot(client *redis.Client, dbName string) {
	//check if the ballot info is present in redis
	stillThere, err := client.Cmd("EXISTS", dbName+"V1").Int()
	handleErr(err)
	if stillThere == 0 {
		fmt.Println("Ballot not in memory, either deleted or not made to start with!")
		return
	}

	//get the votes
	V1, _ := client.Cmd("GET", dbName+"V1").Str()
	V2, _ := client.Cmd("GET", dbName+"V2").Str()
	V3, _ := client.Cmd("GET", dbName+"V3").Str()
	V4, _ := client.Cmd("GET", dbName+"V4").Str()
	V5, _ := client.Cmd("GET", dbName+"V5").Str()

	//get the names and descrips
	var info map[string]string
	info, err = client.Cmd("HGETALL", dbName+"Info").Hash()

	//insert all of the info into a map
	Ballot := map[string]string{
		"name":        info["name"],
		"description": info["description"],
		info["V1"]:    V1,
		info["V2"]:    V2,
		info["V3"]:    V3,
		info["V4"]:    V4,
		"NotA":        V5,
	}

	fmt.Println(Ballot)
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
