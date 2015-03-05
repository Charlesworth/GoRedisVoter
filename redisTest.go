package main

import (
	"fmt"
	"github.com/fzzy/radix/redis"
	//"github.com/gorilla/mux"
	//"log"
	"math/rand"
	//"net/http"
	"strconv"
)

//TODO
//add in a difference in what is returned if the ballot is open in getBallot()
//make the DB unique name
//

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

	fmt.Println(randID())

}

//returns a string of 10 random numbers
func randID() (ID string) {
	return strconv.Itoa(rand.Int())[2:12]
}

func setBallot(client *redis.Client, dbName string, info map[string]string, voteTimeout int) {

	//set the info hash and give a extra day of timeout
	client.Append("HMSET", dbName+"Info", info)
	client.Append("EXPIRE", dbName+"Info", voteTimeout+86400)

	//set the IP set and also give a extra day of timeout
	client.Append("SADD", dbName+"IpSet", "1")
	client.Append("EXPIRE", dbName+"IpSet", voteTimeout+86400)

	//set v1 - v5 and add the timeout
	client.Append("SET", dbName+"V1", "0")
	client.Append("EXPIRE", dbName+"V1", voteTimeout+86400)
	client.Append("SET", dbName+"V2", "0")
	client.Append("EXPIRE", dbName+"V2", voteTimeout+86400)
	client.Append("SET", dbName+"V3", "0")
	client.Append("EXPIRE", dbName+"V3", voteTimeout+86400)
	client.Append("SET", dbName+"V4", "0")
	client.Append("EXPIRE", dbName+"V4", voteTimeout+86400)
	client.Append("SET", dbName+"V5", "0")
	client.Append("EXPIRE", dbName+"V5", voteTimeout+86400)

	//set the write timeout
	client.Append("SET", dbName+"Write", "1")
	client.Append("EXPIRE", dbName+"Write", voteTimeout)

	//go through each command and panic if theres an error
	for i := 0; i < 14; i++ {
		r := client.GetReply()
		handleErr(r.Err)
		a := fmt.Sprintf("for %d", i) //tester code
		fmt.Println(a)                //tester code
	}
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
