package main

import (
	"encoding/json"
	//"fmt"
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
	Time int
}

type Vote struct {
	Vote string
}

var Client redis.Client

func main() {

	log.Println("   __    __")
	log.Println("  |  _  |      | /")
	log.Println("  |__|o | edis |/oter by Charles Cochrane")
	log.Println("https://github.com/Charlesworth/GoRedisVoter")
	log.Println("--------------------------------------------")

	Client, err := redis.Dial("tcp", "178.62.74.225:6379")
	handleErr(err)
	defer Client.Close()

	foo, err := Client.Cmd("PING").Str()
	handleErr(err)
	log.Println("Redis Connection Reply: " + foo + " (we're good to go)")

	foo, err = Client.Cmd("FLUSHALL").Str() //test code
	handleErr(err)                          //test code

	rtr := mux.NewRouter()
	rtr.HandleFunc("/ballot/{id:[0-9]+}", getBallot(Client)).Methods("GET")
	rtr.HandleFunc("/ballot/{id:[0-9]+}", postVote(Client)).Methods("POST")
	//rtr.PathPrefix("/make").Methods("GET").Handler(http.FileServer(http.Dir("./website")))
	//rtr.HandleFunc("/make", testMeth).Methods("GET")
	//rtr.PathPrefix("/make").Handler(http.FileServer(http.Dir("./website")))
	rtr.HandleFunc("/make", postBallot(Client)).Methods("POST")
	//rtr.PathPrefix("/").Handler(http.FileServer(http.Dir("./website")))

	http.Handle("/", rtr)

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

// http req: http://localhost:3000/make
// {"Name":"Alice","Desc":"Hello","V1":"Pizza","V2":"Pasta","V3":"Burger","V4":"Fish","Time":"FUCK"}
func postBallot(Client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		body, _ := ioutil.ReadAll(r.Body)
		var d Message
		json.Unmarshal(body, &d)

		dbName := randID()
		voteTimeout := d.Time
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
		}

		//output to HTTP request and to the programs logs, 303 to redirect to get ballot
		w.Header().Add("Location", "http://localhost:3000/ballot/"+dbName)
		w.WriteHeader(303)
		log.Println("[request: postBallot, IP " + r.RemoteAddr + "] [responce: status 303; id " + dbName + " created]")
	}
}

// http req: http://localhost:3000/ballot/1234
func getBallot(Client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		dbName := params["id"]

		//check if the ballot info is present in redis
		stillThere, err := Client.Cmd("EXISTS", dbName+"V1").Int()
		handleErr(err)
		if stillThere == 0 {
			//output to HTTP request and to the programs logs then return
			w.WriteHeader(404)
			w.Write([]byte("Ballot " + dbName + " not in memory, either deleted or never made"))
			log.Println("[request: getBallot, IP " + r.RemoteAddr + "] [responce: status 404; id " + dbName + " not found in redis]")
			return
		}

		//get the names and descrips
		var info map[string]string
		info, err = Client.Cmd("HGETALL", dbName+"Info").Hash()

		var Ballot map[string]string

		//check write to see if write timedout, and reply with vote numbers depending
		write, err := Client.Cmd("EXISTS", dbName+"Write").Int()
		handleErr(err)
		if write == 1 {
			//The ballot is still open, show don't show votes
			Ballot = map[string]string{
				"name":        info["name"],
				"description": info["description"],
				"V1":          info["V1"],
				"V2":          info["V2"],
				"V3":          info["V3"],
				"V4":          info["V4"],
			}

		} else {
			//The ballot is closed, get the votes
			V1, _ := Client.Cmd("GET", dbName+"V1").Str()
			V2, _ := Client.Cmd("GET", dbName+"V2").Str()
			V3, _ := Client.Cmd("GET", dbName+"V3").Str()
			V4, _ := Client.Cmd("GET", dbName+"V4").Str()
			V5, _ := Client.Cmd("GET", dbName+"V5").Str()

			//insert all of the info into a map
			Ballot = map[string]string{
				"name":        info["name"],
				"description": info["description"],
				info["V1"]:    V1,
				info["V2"]:    V2,
				info["V3"]:    V3,
				info["V4"]:    V4,
				"NotA":        V5,
			}
		}

		jsonResponce, _ := json.Marshal(Ballot)

		//output to HTTP request and to the programs logs
		w.WriteHeader(200)
		w.Write(jsonResponce)
		log.Println("[request: getBallot, IP " + r.RemoteAddr + "] [responce: status 200; id " + dbName + "]")
	}
}

// http req: http://localhost:3000/ballot/1234
// {"Vote":"V1"}
func postVote(Client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		dbName := params["id"]

		body, _ := ioutil.ReadAll(r.Body)
		var d Vote
		json.Unmarshal(body, &d)

		ipAddress := r.RemoteAddr
		v := d.Vote

		//check if the ballot info is present in redis
		stillThere, err := Client.Cmd("EXISTS", dbName+"V1").Int()
		handleErr(err)
		if stillThere == 0 {
			//output to HTTP request and to the programs logs then return
			w.WriteHeader(404)
			w.Write([]byte("Ballot " + dbName + " not in memory, either deleted or never made"))
			log.Println("[request: postVote, IP " + r.RemoteAddr + "] [responce: status 404; id " + dbName + " not found in redis]")
			return
		}

		//check the IP set if the IP address has already voted or not
		ipPresent, err := Client.Cmd("SADD", dbName+"IpSet", ipAddress).Int()
		handleErr(err)
		if ipPresent == 0 {
			w.WriteHeader(403)
			w.Write([]byte("IP " + ipAddress + " already voted on this ballot"))
			log.Println("[request: postVote, IP " + r.RemoteAddr + "] [responce: status 403; id " + dbName + " IP already votes]")
			return
		}

		//check write to see if timed-out
		write, err := Client.Cmd("EXISTS", dbName+"Write").Int()
		handleErr(err)
		if write == 0 {
			w.WriteHeader(403)
			w.Write([]byte("Ballot " + dbName + " write timeout"))
			log.Println("[request: postVote, IP " + r.RemoteAddr + "] [responce: status 403; id " + dbName + " write timeout]")
			return
		}

		_, err = Client.Cmd("INCR", dbName+v).Int()
		handleErr(err)

		//output to HTTP request and to the programs logs
		w.WriteHeader(200)
		log.Println("[request: postVote, IP " + r.RemoteAddr + "] [responce: status 200; id " + dbName + " vote " + d.Vote + " added]")
	}
}

//returns a string of 10 random numbers
func randID() (ID string) {
	return strconv.Itoa(rand.Int())[2:12]
}

//prints error to log then panics
func handleErr(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}
