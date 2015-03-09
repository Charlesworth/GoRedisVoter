# GoRedisVoter

Fast, concurrent voting application written with Go, using Redis as a database. Every connection call spins up a new Go routine, making this highly concurrent, allowing 1000s of requests per second. Easy API and in development web interface. I plan to host this as a free service shortly, please get in touch if you would be interested.

##TODO list

- Combine with Redis in a docker image, with auto start using a docker file
- Produce a web interface and templating for ballot calling
- Allow any number of voting catagories
- Make option to allow the votes to be viewed while the ballot is still open
- Link to my comment service (also in development) to allow a comment attachement to each vote

## Usage

There are 3 main things you can do with this application:

- Make a Ballot
- Query the Ballot
- Vote on a Ballot

Ballots are given up to 4 voting options, a short name, a long description and a duration for the ballot to be open (in seconds). Before the end time is met people can place votes on any of the 4 voting options or "None of the above" option. After the end time is met, voting closes and the result is made available. Each IP address can only vote once and the finished result is available for 24 hours after the Ballot closes, at which time it is deleted.

## API
All parameters must be sent in the request body as JSON, not as part of a query string. 

### Make a Ballot: `Post` [appUrl]/make 
This is used to make a Ballot. Feed it a JSON document in the HTTP request body with the following parameters:

| Name | Value Description | Value Example |
| ---- | ---- | ---- |
| "Name" | Display the help window | "Dinner Time" |
| "Description" | Closes a window | "What should I have for dinner tonight?" |
| "V1" | Vote catagory 1 | "Pizza" |
| "V2" | Vote catagory 2 | "Pasta" |
| "V3" | Vote catagory 3| "Burger" |
| "V4" | Vote catagory 4 | "Fish" |
| "Time" | Time in seconds for ballot to be open | 600 |
  
Example:
- POST URL: _localhost:3000/make_
- POST Body: _{"Name":"Dinner Time","Desc":"What should I have for dinner?","V1":"Pizza","V2":"Pasta","V3":"Burger","V4":"Fish","Time":600}_
- Response: 303 redirect to localhost:3000/ballot/*BallotID*

### Query the Ballot: `Get` [appUrl]/ballot/[ballotID]
This API request is used to query the ballot. if the ballot is still active and accepting votes, you will only recieve back the ballot name, description and vote catagory names. Once the vote is over you will also get back the vote results. This requires no JSON in the body of the request, just the ballot ID in the URL.

Example:
- GET URL: _localhost:3000/ballot/7700679194_
- Response if ballot open: 200 {"V1":"Pizza","V2":"Pasta","V3":"Burger","V4":"Fish","description":"What should I have for dinner tonight","name":"Dinner Time"}
- Response if ballot closed: 200 {"Pizza":"0","Pasta":"0","Burger":"0","Fish":"0","NotA":"0","description":"What should I have for dinner tonight","name":"Dinner Time"}

###Vote on a Ballot: `Post` [appUrl]/ballot/[ballotID]
This POST request sends a vote to the application. On each vote the requester IP address is logged, so each IP can only cast a single vote. If the ballot is closed the vote will be refused also. The user can vote for V1 to V4 and additionaly V5, the "None of the above" option.

| Name | Value Description | Value Example |
| ---- | ---- | ---- |
| "Vote" | Vote catagory you wish to vote for, V1 - V5 | "V3" |

Example:
- POST URL: _localhost:3000/ballot/7700679194_
- Post Body: {"Vote":"V1"}
- Response if ballot open: 200
- Response if ballot closed: 403 Ballot 7700679194 write timeout
