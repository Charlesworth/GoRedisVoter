package main

import (
	"encoding/json"
	"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

type Message struct {
	Name string
	Desc string
	V1   string
	V2   string
	V3   string
	V4   string
	Time string
}

type Vote struct {
	Vote string
}

var Client redis.Client

func main() {

	Client, err := redis.Dial("tcp", "178.62.74.225:6379")
	handleErr(err)
	defer Client.Close()

	foo, err := Client.Cmd("PING").Str()
	handleErr(err)
	fmt.Println(foo)

	foo, err = Client.Cmd("FLUSHALL").Str()
	handleErr(err)
	fmt.Println(foo)

	rtr := mux.NewRouter()
	rtr.HandleFunc("/ballot/{id:[0-9]+}", getBallot(Client)).Methods("GET")
	rtr.HandleFunc("/ballot/{id:[0-9]+}", postVote(Client)).Methods("POST")
	//rtr.PathPrefix("/make").Methods("GET").Handler(http.FileServer(http.Dir("./website")))
	rtr.HandleFunc("/make", postBallot(Client)).Methods("POST")
	rtr.PathPrefix("/").Handler(http.FileServer(http.Dir("./website")))

	http.Handle("/", rtr)

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

// http req: http://localhost:3000/make
// {"Name":"Alice","Desc":"Hello","V1":"Pizza","V2":"Pasta","V3":"Burger","V4":"Fish","Time":"FUCK"}
func postBallot(Client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println("response Body:", string(body))
		var d Message
		json.Unmarshal(body, &d)
		fmt.Println(d)

		dbName := randID()
		voteTimeout := 1234
		info := map[string]string{
			"name":        d.Name,
			"description": d.Desc,
			"V1":          d.V1,
			"V2":          d.V2,
			"V3":          d.V3,
			"V4":          d.V4,
		}

		//set the info hash and give a extra day of timeout
		Client.Append("HMSET", dbName+"Info", info)
		Client.Append("EXPIRE", dbName+"Info", voteTimeout+86400)

		//set the IP set and also give a extra day of timeout
		Client.Append("SADD", dbName+"IpSet", "1")
		Client.Append("EXPIRE", dbName+"IpSet", voteTimeout+86400)

		//set v1 - v5 and add the timeout
		Client.Append("SET", dbName+"V1", "0")
		Client.Append("EXPIRE", dbName+"V1", voteTimeout+86400)
		Client.Append("SET", dbName+"V2", "0")
		Client.Append("EXPIRE", dbName+"V2", voteTimeout+86400)
		Client.Append("SET", dbName+"V3", "0")
		Client.Append("EXPIRE", dbName+"V3", voteTimeout+86400)
		Client.Append("SET", dbName+"V4", "0")
		Client.Append("EXPIRE", dbName+"V4", voteTimeout+86400)
		Client.Append("SET", dbName+"V5", "0")
		Client.Append("EXPIRE", dbName+"V5", voteTimeout+86400)

		//set the write timeout
		Client.Append("SET", dbName+"Write", "1")
		Client.Append("EXPIRE", dbName+"Write", voteTimeout)

		//go through each command and panic if theres an error
		for i := 0; i < 14; i++ {
			n := Client.GetReply()
			handleErr(n.Err)
			a := fmt.Sprintf("for %d", i) //tester code
			fmt.Println(a)                //tester code
		}

		w.Write([]byte("vote created: http://localhost:3000/ballot/" + dbName))
	}
}

// http req: http://localhost:3000/ballot/1234
func getBallot(Client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		dbName := params["id"]
		log.Println("GET request on ballot id: " + dbName)
		w.Write([]byte("Hello fart " + dbName))

		//check if the ballot info is present in redis
		stillThere, err := Client.Cmd("EXISTS", dbName+"V1").Int()
		handleErr(err)
		if stillThere == 0 {
			fmt.Println("Ballot not in memory, either deleted or not made to start with!")
			return
		}

		//get the votes
		V1, _ := Client.Cmd("GET", dbName+"V1").Str()
		V2, _ := Client.Cmd("GET", dbName+"V2").Str()
		V3, _ := Client.Cmd("GET", dbName+"V3").Str()
		V4, _ := Client.Cmd("GET", dbName+"V4").Str()
		V5, _ := Client.Cmd("GET", dbName+"V5").Str()

		//get the names and descrips
		var info map[string]string
		info, err = Client.Cmd("HGETALL", dbName+"Info").Hash()

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
}

// http req: http://localhost:3000/ballot/1234
// {"Vote":"V1"}
func postVote(Client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		dbName := params["id"]
		log.Println("GET request on ballot id: " + dbName)
		w.Write([]byte("Hello fart " + dbName))

		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println("response Body:", string(body))
		var d Vote
		json.Unmarshal(body, &d)
		fmt.Println(d)

		//ipAddress := r.RemoteAddr
		v := d.Vote

		//check the IP set if the IP address has already voted or not
		//		ipPresent, err := Client.Cmd("SADD", dbName+"IpSet", ipAddress).Int()
		//		handleErr(err)
		//		if ipPresent == 0 {
		//			fmt.Println("IP " + ipAddress + " already voted, go away!")
		//			return
		//		}

		//check write to see if timedout
		write, err := Client.Cmd("EXISTS", dbName+"Write").Int()
		fmt.Println(write)
		if write == 0 {
			fmt.Println("Ballot " + dbName + " write timeout, no more voting, go away!")
			return
		}

		//add vote code here
		fmt.Println("Vote MUTHER FUCKER!")
		_, err = Client.Cmd("INCR", dbName+v).Int()
		handleErr(err)
	}
}

//returns a string of 10 random numbers
func randID() (ID string) {
	return strconv.Itoa(rand.Int())[2:12]
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
